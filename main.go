package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/krateoplatformops/authn/internal/certrenewal"
	"github.com/krateoplatformops/authn/internal/env"
	kubeconfig "github.com/krateoplatformops/authn/internal/helpers/kube/config"
	"github.com/krateoplatformops/authn/internal/helpers/kube/util"
	"github.com/krateoplatformops/authn/internal/helpers/restaction"
	"github.com/krateoplatformops/authn/internal/middlewares/cors"
	"github.com/krateoplatformops/authn/internal/routes"
	"github.com/krateoplatformops/authn/internal/routes/auth/basic"
	"github.com/krateoplatformops/authn/internal/routes/auth/info"
	"github.com/krateoplatformops/authn/internal/routes/auth/ldap"
	"github.com/krateoplatformops/authn/internal/routes/auth/oauth"
	"github.com/krateoplatformops/authn/internal/routes/auth/oidc"
	"github.com/krateoplatformops/authn/internal/routes/auth/strategies"
	"github.com/krateoplatformops/authn/internal/routes/health"
	xcontext "github.com/krateoplatformops/plumbing/context"
	"github.com/krateoplatformops/plumbing/jwtutil"
	"github.com/rs/zerolog"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

const (
	serviceName = "auth"
)

var (
	Version string
	Build   string
)

func main() {
	// Flags
	kconfig := flag.String(clientcmd.RecommendedConfigPathFlag, "", "absolute path to the kubeconfig file")
	debugOn := flag.Bool("debug", env.Bool("AUTHN_DEBUG", false), "dump verbose output")
	dumpEnv := flag.Bool("dump-env", env.Bool("AUTHN_DUMP_ENV", false), "dump environment variables")
	corsOn := flag.Bool("cors", env.Bool("AUTHN_CORS", true), "enable or disable CORS")
	servicePort := flag.Int("port", env.Int("AUTHN_PORT", 8082), "port to listen on")
	certExpiresIn := flag.Duration("cert-expires",
		env.Duration("AUTHN_KUBECONFIG_CRT_EXPIRES_IN", time.Hour*24), "requested duration of a login certificate (default: 24h)")
	serviceCertExpiresIn := flag.Duration("service-cert-expires",
		env.Duration("AUTHN_SERVICE_CRT_EXPIRES_IN", certrenewal.DefaultDuration),
		"requested duration of authn's own service certificate (default: 8760h)")
	certRenewalOn := flag.Bool("cert-renewal",
		env.Bool("AUTHN_CRT_RENEWAL_ENABLED", true),
		"keep authn's own service certificate renewed in the background")
	certRenewalThreshold := flag.Float64("cert-renewal-threshold",
		env.Float64("AUTHN_CRT_RENEWAL_THRESHOLD", certrenewal.DefaultThreshold),
		"fraction of the GRANTED certificate lifetime after which it is re-issued (default: 0.66)")

	clusterName := flag.String("kubeconfig-cluster-name",
		env.String("AUTHN_KUBECONFIG_CLUSTER_NAME", "krateo"), "cluster name for generated kubeconfig")
	kubernetesURL := flag.String("kubeconfig-server-url",
		env.String("AUTHN_KUBECONFIG_SERVER_URL", ""), "kubernetes api server url for generated kubeconfig")
	snowplowHOST := flag.String("snowplow-host",
		env.String("SNOWPLOW_SERVICE_HOST", ""), "snowplow host for restaction api calls")
	snowplowPORT := flag.String("snowplow-post",
		env.String("SNOWPLOW_SERVICE_PORT", "8081"), "snowplow port for restaction api calls")
	var snowplowURL *string
	temp := "http://" + *snowplowHOST + ":" + *snowplowPORT
	snowplowURL = &temp
	if *snowplowURL == "http://:8081" {
		snowplowURL = flag.String("snowplow-url", env.String("URL_SNOWPLOW", "http://snowplow.krateo-system.svc.cluster.local:8081"), "snowplow url for restaction api calls")
	}
	storageNamespace := flag.String("namespace",
		env.String("AUTHN_NAMESPACE", ""), "namespace where to store secrets with generated config")
	authnUsername := flag.String("authn-username",
		env.String("AUTHN_USERNAME", "authn"), "authn username for clientconfig for restaction api calls")
	signKey := flag.String("jwt-sign-key", env.String("JWT_SIGN_KEY", ""), "secret key used to sign JWT tokens")

	flag.Usage = func() {
		fmt.Fprintln(flag.CommandLine.Output(), "Flags:")
		flag.PrintDefaults()
	}

	flag.Parse()

	if len(*storageNamespace) > 0 {
		os.Setenv(util.NamespaceEnvVar, *storageNamespace)
	}

	// Initialize the logger
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix

	// Default level for this log is info, unless debug flag is present
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	if *debugOn {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	}

	log := zerolog.New(os.Stdout).With().
		Str("service", serviceName).
		Timestamp().
		Logger()

	if log.Debug().Enabled() {
		evt := log.Debug().
			Str("version", Version).
			Str("build", Build).
			Str("debug", fmt.Sprintf("%t", *debugOn)).
			Str("cors", fmt.Sprintf("%t", *corsOn)).
			Str("port", fmt.Sprintf("%d", *servicePort)).
			Str("clusterName", *clusterName).
			Str("kubernetesURL", *kubernetesURL).
			Dur("certExpire", *certExpiresIn).
			Dur("serviceCertExpire", *serviceCertExpiresIn).
			Bool("certRenewal", *certRenewalOn)

		if *dumpEnv {
			evt = evt.Strs("env-vars", os.Environ())
		}

		evt.Msg("configuration and env vars info")
	}

	log.Debug().Msgf("Snowplow URL from Service ENV: %s", temp)
	log.Debug().Msgf("Snowplow URL computed/from parameter: %s", *snowplowURL)

	// Kubernetes configuration
	var cfg *rest.Config
	var err error
	if len(*kconfig) > 0 {
		cfg, err = clientcmd.BuildConfigFromFlags("", *kconfig)
	} else {
		cfg, err = rest.InClusterConfig()
	}
	if err != nil {
		log.Fatal().Err(err).Msg("resolving kubeconfig for rest client")
	}

	gen := kubeconfig.NewGenerator(cfg,
		kubeconfig.KubernetesURL(*kubernetesURL),
		kubeconfig.CertDuration(*certExpiresIn),
		kubeconfig.ClusterName(*clusterName),
		kubeconfig.Log(log),
	)

	healthy := int32(0)

	all := []routes.Route{}
	all = append(all, strategies.List(cfg))
	all = append(all, info.Info(cfg))
	all = append(all, health.Check(&healthy, Version, serviceName))

	all = append(all, basic.Login(cfg, basic.LoginOptions{
		KubeconfigGenerator: gen,
		JwtDuration:         *certExpiresIn,
		JwtSingKey:          *signKey,
	}))

	all = append(all, ldap.Login(cfg, ldap.LoginOptions{
		KubeconfigGenerator: gen,
		JwtDuration:         *certExpiresIn,
		JwtSingKey:          *signKey,
	}))

	accessToken, err := jwtutil.CreateToken(jwtutil.CreateTokenOptions{
		Username:   *authnUsername,
		Groups:     []string{"authn"},
		SigningKey: *signKey,
		Duration:   time.Hour * 8760, // 1 year,
	})
	if err != nil {
		log.Fatal().Err(err).Msgf("cannot create jwt token for %s", *authnUsername)
	}

	all = append(all, oauth.Login(
		xcontext.BuildContext(context.Background(),
			xcontext.WithAccessToken(accessToken),
			func(ctx context.Context) context.Context {
				ctx = context.WithValue(ctx,
					restaction.RestActionContextKey("username"), *authnUsername)
				return context.WithValue(ctx,
					restaction.RestActionContextKey("snowplowURL"), *snowplowURL)
			},
		), cfg, oauth.LoginOptions{
			KubeconfigGenerator: gen,
			JwtDuration:         *certExpiresIn,
			JwtSingKey:          *signKey,
		}))

	all = append(all, oidc.Login(
		xcontext.BuildContext(context.Background(),
			xcontext.WithAccessToken(accessToken),
			func(ctx context.Context) context.Context {
				ctx = context.WithValue(ctx,
					restaction.RestActionContextKey("username"), *authnUsername)
				return context.WithValue(ctx,
					restaction.RestActionContextKey("snowplowURL"), *snowplowURL)
			},
		), cfg, oidc.LoginOptions{
			KubeconfigGenerator: gen,
			JwtDuration:         *certExpiresIn,
			JwtSingKey:          *signKey,
		}))

	handler := routes.Serve(all, log)
	if *corsOn {
		c := cors.New(cors.Options{
			AllowedOrigins:   []string{"*"},
			AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
			AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-Auth-Code"},
			ExposedHeaders:   []string{"Link"},
			AllowCredentials: true,
			MaxAge:           300, // Maximum value not ignored by any of major browsers
		})

		handler = c.Handler(handler)
	}

	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", *servicePort),
		Handler:      handler,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 50 * time.Second,
		IdleTimeout:  30 * time.Second,
	}

	ctx, stop := signal.NotifyContext(context.Background(), []os.Signal{
		os.Interrupt,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGKILL,
		syscall.SIGHUP,
		syscall.SIGQUIT,
	}...)
	defer stop()

	// authn's own clientconfig — the identity snowplow presents when it runs
	// RESTActions on authn's behalf. Issued synchronously here, as it always
	// was, and then kept alive: the signer may grant far less than the year we
	// ask for, and nothing but a pod restart used to re-issue it.
	renewer, err := certrenewal.New(certrenewal.Options{
		RestConfig: cfg,
		Namespace:  *storageNamespace,
		CAData:     string(cfg.CAData),
		ServerURL:  *kubernetesURL,
		Username:   *authnUsername,
		Groups:     []string{"authn"},
		Duration:   *serviceCertExpiresIn,
		Threshold:  *certRenewalThreshold,
		Log:        log,
	})
	switch {
	case err != nil:
		// A clientconfig authn cannot create has never been a reason to refuse
		// logins — only snowplow-backed identity enrichment degrades — so this
		// stays as non-fatal as the fire-and-forget signup it replaces.
		log.Err(err).Msg("client certificate renewal unavailable")

	default:
		// This read is the only one until renewal falls due: the wait it
		// returns is handed straight to the loop.
		wait, err := renewer.Ensure(context.Background())
		if err != nil {
			log.Err(err).Msg("issuing authn clientconfig")
		}

		if *certRenewalOn {
			go renewer.Run(ctx, wait)
		} else {
			log.Warn().
				Str("secret", renewer.SecretName()).
				Msg("client certificate renewal is disabled; the stored certificate will expire without a restart")
		}
	}

	go func() {
		atomic.StoreInt32(&healthy, 1)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal().Err(err).Msgf("could not listen on %s", server.Addr)
		}
	}()

	// Listen for the interrupt signal.
	log.Info().Msgf("server is ready to handle requests at @ %s", server.Addr)
	<-ctx.Done()

	// Restore default behavior on the interrupt signal and notify user of shutdown.
	stop()
	log.Info().Msg("server is shutting down gracefully, press Ctrl+C again to force")
	atomic.StoreInt32(&healthy, 0)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	server.SetKeepAlivesEnabled(false)
	if err := server.Shutdown(ctx); err != nil {
		log.Fatal().Err(err).Msg("server forced to shutdown")
	}

	log.Info().Msg("server gracefully stopped")
}
