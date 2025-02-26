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
	"github.com/krateoplatformops/snowplow/plumbing/signup"
	"github.com/rs/zerolog"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

const (
	serviceName = "auth-service"
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
		env.Duration("AUTHN_KUBECONFIG_CRT_EXPIRES_IN", time.Hour*24), "generated certificate duration (default: 24h)")

	clusterName := flag.String("kubeconfig-cluster-name",
		env.String("AUTHN_KUBECONFIG_CLUSTER_NAME", "krateo"), "cluster name for generated kubeconfig")
	kubernetesURL := flag.String("kubeconfig-server-url",
		env.String("AUTHN_KUBECONFIG_SERVER_URL", ""), "kubernetes api server url for generated kubeconfig")
	snowplowURL := flag.String("snowplow-url",
		env.String("URL_SNOWPLOW", "http://snowplow.krateo-system.svc.cluster.local:8081"), "snowplow url for restaction api calls")
	storageNamespace := flag.String("namespace",
		env.String("AUTHN_NAMESPACE", ""), "namespace where to store secrets with generated config")
	authnUsername := flag.String("authn-username",
		env.String("AUTHN_USERNAME", "authn"), "authn username for clientconfig for restaction api calls")

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
			Dur("certExpire", *certExpiresIn)

		if *dumpEnv {
			evt = evt.Strs("env-vars", os.Environ())
		}

		evt.Msg("configuration and env vars info")
	}

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

	ctx := context.Background()
	ctx = context.WithValue(ctx, restaction.RestActionContextKey("username"), *authnUsername)
	ctx = context.WithValue(ctx, restaction.RestActionContextKey("snowplowURL"), *snowplowURL)

	healthy := int32(0)

	all := []routes.Route{}
	all = append(all, health.Check(&healthy, Version, serviceName))
	all = append(all, basic.Login(cfg, gen))
	all = append(all, oauth.Login(ctx, cfg, gen))
	all = append(all, ldap.Login(cfg, gen))
	all = append(all, oidc.Login(ctx, cfg, gen))
	all = append(all, strategies.List(cfg))
	all = append(all, info.Info(cfg))

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

	ctx, stop := signal.NotifyContext(ctx, []os.Signal{
		os.Interrupt,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGKILL,
		syscall.SIGHUP,
		syscall.SIGQUIT,
	}...)
	defer stop()

	// Create authn clientconfig to call snowplow's RESTActions
	_, err = signup.Do(context.TODO(), signup.Options{
		RestConfig:   cfg,
		Namespace:    *storageNamespace,
		CAData:       string(cfg.CAData),
		ServerURL:    *kubernetesURL,
		CertDuration: time.Hour * 8760, // 1 year
		Username:     *authnUsername,
		UserGroups:   []string{"authn"},
	})

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
