# Krateo AuthN Service

## List available strategies

The `GET /strategies` endpoint shows all available authentication strategies.

> When a particular authentication strategy requires additional configuration parameters, these will be exposed under the _"extensions"_ key. Otherwise this attribute will not be present.

Example:

```sh
$ curl https://api.krateoplatformops.io/authn/strategies
```

```json
[
  {
    "kind": "basic",
    "path": "/basic/login"
  },
  {
    "kind": "ldap",
    "name": "forumsys",
    "path": "/ldap/login"
  },
  {
    "kind": "ldap",
    "name": "openldap",
    "graphics": {
      "icon": "key",
      "displayName": "Login with LDAP",
      "backgroundColor": "#ffffff",
      "textColor": "#000000"
    },
    "path": "/ldap/login"
  },
  {
    "kind": "oauth",
    "name": "github-example",
    "graphics": {
      "icon": "fa-brands fa-github",
      "displayName": "Login with Github",
      "backgroundColor": "#ffffff",
      "textColor": "#000000"
    },
    "path": "/oauth/login",
    "extensions": {
      "authCodeURL": "https://github.com/login/oauth/authorize?client_id=XXXX&redirect_uri=http%3A%2F%2Flocalhost%3A8888%2Fgithub%2Fgrant&response_type=code&scope=read%3Auser+read%3Aorg&state=YYYY",
      "redirectURL": "http://localhost:30080/auth?kind=oauth"
    }
  },
  {
    "kind": "oidc",
    "name": "oidc-example",
    "graphics": {
      "icon": "fa-brands fa-windows",
      "displayName": "Login with Azure",
      "backgroundColor": "#4444ff",
      "textColor": "#ffffff"
    },
    "path": "/oidc/login",
    "extensions": {
      "authCodeURL": "https://login.microsoftonline.com/XXXX/oauth2/v2.0/authorize?client_id=XXXX\u0026redirect_uri=http%3A%2F%2Flocalhost%3A8080%2Foidc%2Fcallbacl\u0026response_mode=query\u0026response_type=code\u0026scope=openid+email+profile+User.Read",
      "redirectURL": "http://localhost:8080/auth?kind=oidc"
    }
  }
]
```

## Authentication

Regardless of the strategy used, the response will always be a json with the following structure:

```json
{
   "code":200,
   "user":{
      "displayName":"John Doe",
      "username":"johndoe",
      "avatarURL":"https://avatars.githubusercontent.com/u/585381?v=4"
   },
   "groups": [
      "devs"
   ],
   "data":{
      "apiVersion":"v1",
      "clusters":[
         {
            "cluster":{
               "certificate-authority-data":"<base64-ca-cert-data>",
               "server":"https://127.0.0.1:51461"
            },
            "name":"krateo"
         }
      ],
      "contexts":[
         {
            "context":{
               "cluster":"krateo",
               "user":"johndoe"
            },
            "name":"krateo"
         }
      ],
      "current-context":"krateo",
      "kind":"Config",
      "users":[
         {
            "user":{
               "client-certificate-data":"<base64-user-cert-data>",
               "client-key-data":"<base64-user-cert-key-data>"
            },
            "name":"johndoe"
         }
      ]
   }
}
```

### Login with Basic Authentication

The Authorization header field is constructed as follows:

- username and password are combined with a single colon
  - this means that the username itself cannot contain a colon

- the resulting string is encoded using a variant of Base64 (+/ and with padding)

- the authorization method and a space character (e.g. "Basic ") is then prepended to the encoded string.

For example, if the username is Aladdin and the password is open sesame, then the field's value is the Base64 encoding of Aladdin:open sesame, or QWxhZGRpbjpvcGVuIHNlc2FtZQ==

Then the Authorization header field will appear as: _Authorization: Basic QWxhZGRpbjpvcGVuIHNlc2FtZQ==_

Example:

```sh
curl https://reqbin.com/echo
   -H "Authorization: Basic QWxhZGRpbjpvcGVuIHNlc2FtZQ=="
```

### Login with OAuth Authorization Code Flow

> Let's take _Github_ as example, the same concept applies to all authentication systems of this type (authorization code flow).

With a valid [authorization code grant](https://www.oauth.com/oauth2-servers/access-tokens/authorization-code-request/) invoke the related endpoint `path`, passing the `name` as query string parameter.

Example:

```sh
$ curl -H "X-Auth-Code: $(AUTH_CODE)" \
    https://api.krateoplatformops.io/authn/oauth/login?name=github-example
```

The `RestActionRef` is mandatory because, unlike OIDC, OAuth2 returns only the bearer token and no information regarding the user, therefore, the RESTAction is required to compile all fields. See the [RESTAction Configuration](#restaction-configuration) section for more information.

### Login with LDAP

To login using LDAP credentials must be sent as JSON using POST:

```sh
$ curl -X POST "https://api.krateoplatformops.io/authn/ldap/login?name=openldap" \
   -H 'Content-Type: application/json' \
   -d '{"username":"XXXXXX","password":"YYYYYY"}'
```

### Login with OIDC

To login using OIDC credentials, the authorization code must be sent throught the `X-Auth-Code` header field:

```sh
$ curl -H "X-Auth-Code: $(AUTH_CODE)" \
    https://api.krateoplatformops.io/authn/oidc/login?name=oidc-example
```

The authn application supports the Discovery endpoint. If you provide a Discovery endpoint the values for `authorizationURL`, `tokenURL` and `userInfoURL` are ignored and overwritten. If you do not provide a Discovery endpoint, the values for `authorizationURL`, `tokenURL` and `userInfoURL` are used.

To obtain proper groups mappings you need to configure the ID Token response on the application side. Likewise for the profile picture. Examples are listed below for Azure and KeyCloak. Alternatively, you can use the RESTAction. See the [RESTAction Configuration](#restaction-configuration) section for more information.

#### Azure
Azure can be configured to authenticate users through OIDC ([Official Azure Documentation for OIDC](https://learn.microsoft.com/en-us/entra/identity-platform/v2-protocols-oidc)). To achieve this, you need to create a new app registration as follows:

On Azure:
 - Go to "App registrations" and then hit "New registration";
 - Configure the display name, account types and Redirect URI. The redirect URI must point to Krateo's frontend with an HTTPS endpoint and the path `/auth/oidc`;
 - Create a client secret in "Certificates & secrets", save the value of the secret now as it cannot be visualized afterwards;
 - In the "Authentication" menu, find and activate `Access tokens` and `ID tokens`;
 - In the "API permissions" menu, add the following: `openid`, `email`, `profile`, `User.Read` and `User.ReadBasic.All`;

To obtain groups you can do one of the following:
- Include them in the OIDC ID Token response: in the "Manifest" menu in the Azure portal, modify the value `groupMembershipClaims` to `All` ([Official Azure documentation for the groupMembershipClaims](https://learn.microsoft.com/en-us/entra/identity-platform/reference-app-manifest#groupmembershipclaims-attribute));
- Use the following RESTAction and place it in the `RestActionRef` of the OIDC custom resource:
```yaml
apiVersion: templates.krateo.io/v1
kind: RESTAction
metadata:
  name: test-rest-action
  namespace: krateo-system
spec:
  api:
  - name: groups
    verb: GET
    headers:
    - 'Accept: application/json'
    path: "v1.0/me/memberOf?$select=displayName"
    endpointRef:
      name: azure-entra
      namespace: krateo-system
    filter: .groups.value | map(.displayName)
```
To obtain groups through the rest action, add `Directory.Read.All` in the `additionalScopes`. There is a more complex example for the Microsoft Entra API that also uses pagination in the testdata folder (requires Authn >= 0.22.0).

On AuthN:
 - To obtain the user avatar/profile image include `User.Read` in the `additionalScopes` field of the OIDCConfiguration custom resource;
 - You can now configure the Authn's CR by using Azure discovery URL, which will be in the following format:
 ```
https://login.microsoftonline.com/<your-tenant-id>/v2.0/.well-known/openid-configuration
 ```

Scopes in the `additionalScopes` field can be simply separated with a space. Example for Azure groups with the RESTAction:
```yaml
  additionalScopes: User.Read Directory.Read.All
```

##### Troubleshooting
If you do not get the correct groups in the AuthN response, please verify your Azure OIDC configuration: the manifest value "groupMembershipClaims:All" adds in the JWT ID Token the value "groups", which contains an array of UIDs of the groups the user belongs. You can check the JWT ID Token returned by Azure by simulating the calls to the Authorization and Token endpoints through Postman or curl. The exact endpoints are contained in the "well-known" endpoint (Authorization and Token).

On Azure, set the redirect URI to, for example, "http://localhost:8080" for testing without HTTPS, then open the following page in a web browser:
```
https://login.microsoftonline.com/<tenant-id>/oauth2/v2.0/authorize?client_id=<client-id>&response_type=code&redirect_uri=http://localhost:8080&response_mode=query&scope=openid email profile User.Read
```
Login, then Azure will redirect you to the redirect URI, which will error out, however, there will be a "code" query parameter in the URL. Copy this code parameter being careful not to copy other query parameters. Through cURL or Postman perform a POST to:
```
https://login.microsoftonline.com/<tentant-id>/oauth2/v2.0/token
```
with the following body values:
```
client_id=<client-id>
client_secret=<client-secret>
code=<the code from the redirect url>
redirect_uri=http://localhost:8080
grant_type=authorization_code
```
You can then decode the JWT and verify that "groups" is present with the [Microsoft offline decoder](https://jwt.ms/).

#### KeyCloak
To obtain groups, add a custom mapper of type "Group Membership" and give it the Token Claim Name "groups", uncheck `Full group path`. Add `groups` into the `additionalScopes` field of the OIDCConfiguration custom resource.
To obtain the user avatar/profile image, go to the realm settings, then "User profiles" tab, "Create Attribute", and add one with the name `picture`. Set the profile picture for the user to a URL pointing to a picture. Keycloak will now return the avatar during authentication.


## RESTAction Configuration
The `RESTActionRef` field in the OAuth2 and OIDC configs is mandatory and optional, respectively. It is used to compile the following fields, used to build the Kubernetes certificate required for authentication:
- `name`: string
- `email`: string
- `preferredUsername`: string - mandatory
- `avatarURL`: string
- `groups`: []string - mandatory

In the case of OAuth2, these fields are needed to compile the certificate and the ones marked as such are mandatory. They have to be included as top level fields in the RESTAction response. See the testdata folder for a Github example. In the OIDC case, these fields are all optional, and if included will overwrite the information obtained from the id token and the userinfo endpoint.

The authentication for the endpoints of the RESTAction is automatically set to bearer token, using the token obtained from the OAuth2/OIDC authentication. This token is passed to the RESTActio as a parameter, and can be used in the `Authorization` header as follows:
```yaml
    headers:
    - "${ \"Authorization: Bearer \" + .token }"
```
See [oidc-azure-pagination](./testdata/oidc-azure-pagination.yaml), [oidc-azure](./testdata/oidc-azure.yaml) and [oauth](./testdata/oauth.yaml) for more examples.

If the RESTAction does not accept a token parameter, then it will temporarily set the token in the respective endpoint.

## Client certificate renewal
authn mints x509 client certificates through the Kubernetes CSR API (signer `kubernetes.io/kube-apiserver-client`) and stores them in `<name>-clientconfig` Secrets. There are two kinds of identity and they are kept alive differently.

### The requested duration is not the granted duration
A signer grants `min(requested expirationSeconds, cluster-signing-duration, signer CA remaining life)`. Vanilla Kubernetes defaults `--cluster-signing-duration` to `8760h`, so the year authn asks for usually survives, which is why this stays invisible until you deploy elsewhere. OpenShift pins the signing duration to 720h and rotates the CSR signer CA every 30 days, so a certificate requested for a year comes back valid for at most 30 days, and for a single day when a rotation is imminent.

authn therefore reads `NotAfter` off the issued certificate instead of assuming the request held, and annotates every stored Secret with the window that was actually granted:
```sh
$ kubectl -n krateo-system get secret authn-clientconfig -o jsonpath='{.metadata.annotations}'
{"authn.krateo.io/certificate-not-before":"2026-09-04T10:00:00Z","authn.krateo.io/certificate-not-after":"2026-10-04T10:00:00Z"}
```
A `warn` is logged at issue time whenever the granted window is under 90% of the requested one, which is the only signal that a signer is capping you. Nothing reads these annotations back: the renewal loop decides off the stored certificate itself, and they exist so that an operator can answer "when does this credential really die" across a whole namespace without decoding anything, e.g. `kubectl get secret -o custom-columns=NAME:.metadata.name,EXPIRES:'.metadata.annotations.authn\.krateo\.io/certificate-not-after'`.

That clamp is what makes `NotAfter` a sufficient signal rather than a hopeful one, and it is worth knowing it is real: Kubernetes' signer sets the issued `NotAfter` to `min(now + ttl, signer CA NotAfter)` (`pkg/controller/certificates/authority/policies.go`, `if !tmpl.NotAfter.Before(signerNotAfter) { tmpl.NotAfter = signerNotAfter }`) and refuses to sign at all once the CA itself has expired, so a certificate can never outlive the CA that signed it. The signer hot-reloads its CA from disk, so signings after a rotation use the new CA. OpenShift's rotation is additive on the trust side: library-go's `manageCABundleConfigMap` prepends the new signer to the CA bundle and prunes only certificates that have genuinely expired, so a certificate issued by a since-rotated CA stays trusted right up to its own `NotAfter`. Renewing at a fraction of the granted window is therefore always in time. Note that a CA being *rotated* in five days is not the same as it *expiring* in five days: if the old CA cert is still valid for another 25 days you are granted up to those 25 days, and that is correct, because the rotated-out CA remains in the bundle until it expires.

The one case this cannot see is a signer CA regenerated out of band with its trust bundle rebuilt from scratch (disaster recovery, a manual signer wipe). Previously issued certificates then stop being accepted while their `NotAfter` is still in the future, which surfaces as `x509: certificate signed by unknown authority` rather than as an expiry. authn cannot detect it, because snowplow rather than authn is what presents the certificate; recovery is a pod restart or `kubectl delete secret authn-clientconfig`, which the loop re-issues on its next check.

### `authn-clientconfig` is renewed in the background
This is authn's own identity, the one snowplow presents when it runs RESTActions on authn's behalf. Nothing else ever re-issues it: unlike a user certificate there is no login to rewrite it, so it used to be created once at startup and then rot in place, and every RESTAction call failed with `x509: certificate has expired or is not yet valid` until the pod was restarted.

A background loop now issues it at startup and re-issues it once it passes `AUTHN_CRT_RENEWAL_THRESHOLD` of its granted lifetime. Because the threshold is a fraction of what was granted rather than of what was asked for, one setting covers both a full year (re-issued after about 8 months) and a certificate a rotating signer CA cut to an hour (re-issued after about 40 minutes). The certificate is read exactly once per issuance: the signing call returns it, and the clamp above makes its `NotAfter` deterministic, so there is nothing to re-read until renewal falls due. The loop then simply sleeps that long — a year-long certificate is one sleep of about eight months, not a poll. The knobs, both settable under `env` in the chart values:
- `AUTHN_SERVICE_CRT_EXPIRES_IN` (`--service-cert-expires`, default `8760h`): duration requested for authn's own certificate.
- `AUTHN_CRT_RENEWAL_THRESHOLD` (`--cert-renewal-threshold`, default `0.66`): fraction of the granted lifetime after which it is re-issued. This is the whole renewal policy; it must be between 0 and 1 exclusive, and an out-of-range value falls back to the default.
- `AUTHN_CRT_RENEWAL_ENABLED` (`--cert-renewal`, default `true`): set `false` to restore the old behaviour, where the certificate expires in place until the pod restarts.

Failures retry after a minute, and a failed read never triggers a re-issue: an apiserver hiccup is not evidence the credential is gone. Because nothing polls, a Secret deleted or replaced out of band is not noticed until the next scheduled re-issue; deleting it is still the way to force a fresh certificate, but it takes a pod restart to act on it immediately. No new RBAC is needed: the validity annotations are written with the `get` and `update` on `secrets` authn already holds. The loop does assume a single writer, since the CSR object is named after the username and two replicas would race on the CSR named `authn`.

### `<user>-clientconfig` is clamped, not renewed
Per-user certificates are deliberately not on the renewal loop. Instead the login JWT is now issued for `min(AUTHN_KUBECONFIG_CRT_EXPIRES_IN, time until the certificate's real NotAfter)`.

That closes the actual bug. The JWT lifetime used to come from the requested duration, so a truncated certificate left the user logged in while every user-scoped call failed with an x509 error, and with no refresh endpoint the only recovery was waiting out the JWT. With the clamp the certificate outlives every token issued against it by construction, and a user certificate is re-minted on every login, so there is no window left for a background loop to rescue.

Renewing them would also be actively worse: it would rewrite the Secret but not the kubeconfig copy the user is holding, and it would keep minting valid cluster credentials for identities long since deprovisioned upstream, which authn cannot check because it learns who someone is at login and holds no session state.

### Verifying it
Certificate handling differs per distribution, so it has to be measured rather than reasoned about. [`scripts/verify-cert-renewal.sh`](./scripts/verify-cert-renewal.sh) reports the granted `NotAfter`, checks that renewal fires before expiry, and proves the renewed certificate still authenticates against the apiserver. Run it on each flavour you support (kind, minikube, k3s/k3d, EKS, AKS, GKE, OpenShift) and record the row it prints.
```sh
$ scripts/verify-cert-renewal.sh -n krateo-system              # report only
$ scripts/verify-cert-renewal.sh -n krateo-system --wait       # force a renewal now
$ scripts/verify-cert-renewal.sh -n krateo-system --short 15m  # watch a full cycle in ~10 minutes
```

## Graphics Configuration
The OAuth2 and OIDC authentication methods also support a `graphics` object that allows to configure how the button for the redirect to the authentication provider portal is visualized in the frontend login screen.
```yaml
  graphics:
    icon: # icon name from the fontawesome library, for icons like github and windows, use "fa-brands fa-github" or "fa-brands fa-windows"
    displayName: # text to be visualized on the button
    backgroundColor: # color of the button in hexadecimal, e.g., #0022ff
    textColor: # color of the text in the button, also in hexadecimal, e.g., #0022ff
```