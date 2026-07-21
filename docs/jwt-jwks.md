# authn JWT signing & JWKS

authn issues JSON Web Tokens for the Krateo platform. As of the asymmetric-signing
migration it signs with **RS256** using an RSA private key and publishes the matching
public key as a **JWKS** so that any validator — Snowplow, agentgateway, or a third
party — can verify tokens without sharing a secret.

## What a token looks like

- **Algorithm:** `RS256` (asymmetric). Symmetric `HS256` is no longer used.
- **Header:** carries a `kid` identifying the signing key (required by agentgateway).
- **Claims:** `username`, `groups`, `iss: "krateo.io"`, `sub`, `exp`, `iat`, `nbf`.
  No `aud`.

The signing itself lives in the shared `plumbing/jwtutil` library
(`CreateToken` / `Validate`); authn only supplies the key material and `kid`.

## Configuration

authn reads two settings at startup:

| Flag | Env | Meaning |
| --- | --- | --- |
| `--jwt-sign-key-file` | `JWT_SIGN_KEY_FILE` | Path to the PEM-encoded RSA **private** key. |
| `--jwt-kid` | `JWT_KID` | Key ID stamped into every token header and the JWKS. |

Both are required; authn exits at boot if the key file is missing/unparseable or
the `kid` is empty. The private key is read from a **file** (mounted from a Secret),
never passed as a raw env value.

## Creating the signing-key Secret

Generate an RSA keypair and store the private key in a Secret. authn derives the
public key (and the JWKS) from it at runtime — you only ever store the private half.

```sh
# 1. Generate a 2048-bit RSA private key.
openssl genrsa -out private.pem 2048

# 2. Create the Secret in authn's namespace.
kubectl create secret generic jwt-sign-key \
  --namespace krateo-system \
  --from-file=private.pem=./private.pem
```

The Helm chart mounts this Secret at `/etc/authn/jwt/private.pem` and sets
`JWT_SIGN_KEY_FILE` accordingly. Relevant `values.yaml`:

```yaml
jwt:
  signKeySecretName: jwt-sign-key   # Secret name
  signKeySecretKey: private.pem     # key within the Secret + mounted filename
  mountPath: /etc/authn/jwt         # mount directory
  kid: krateo-authn-key-1           # stable key ID (kid)
```

> Keep `kid` stable for the life of the key. Changing the key without changing the
> `kid` (or vice versa) makes previously issued tokens unverifiable.

## The JWKS endpoint

authn serves the public key set at:

```
GET /.well-known/jwks.json
```

on its normal service port (default `8082`) — no separate service or route change
is required. The response is a standard JWKS:

```json
{
  "keys": [
    {
      "kty": "RSA",
      "use": "sig",
      "alg": "RS256",
      "kid": "krateo-authn-key-1",
      "n": "<base64url-modulus>",
      "e": "AQAB"
    }
  ]
}
```

Quick check once deployed:

```sh
kubectl -n krateo-system port-forward svc/authn 8082:8082
curl -s http://localhost:8082/.well-known/jwks.json | jq
```

## Consuming the JWKS

### agentgateway (remote JWKS)

Point agentgateway at the endpoint via `jwks.remote` so it refetches on a cadence
(supports key rotation). The `issuer` matches authn's `iss`; omit `audiences`
because authn emits no `aud`.

```yaml
apiVersion: agentgateway.dev/v1alpha1
kind: AgentgatewayPolicy
metadata:
  name: authn-jwt
spec:
  targetRefs:
  - group: gateway.networking.k8s.io
    kind: HTTPRoute
    name: mcp
  traffic:
    jwtAuthentication:
      mode: Strict
      providers:
      - issuer: "krateo.io"
        jwks:
          remote:
            backendRef:
              name: authn            # authn's Service
              kind: Service
              namespace: krateo-system
              port: 8082
            jwksPath: /.well-known/jwks.json
            cacheDuration: 5m
```

`groups`-based RBAC is enforced by a separate authorization policy over the decoded
claims.

### Snowplow

Snowplow validates the same tokens through the shared `plumbing/jwtutil` (via the
`server/use.UserConfig` middleware), which now verifies **RS256 with a public key**
instead of the old shared secret. Snowplow therefore takes authn's **public** key
(PEM), mounted as a file just like authn's private key:

| Flag | Env | Meaning |
| --- | --- | --- |
| `--jwt-public-key-file` | `JWT_PUBLIC_KEY_FILE` | Path to the PEM-encoded RSA **public** key (authn's public half). |

Derive the public key from the same private key authn signs with and distribute it
to Snowplow (e.g. a Secret/ConfigMap mounted as a file):

```sh
openssl rsa -in private.pem -pubout -out public.pem
kubectl create secret generic authn-jwt-public-key \
  --namespace krateo-system \
  --from-file=public.pem=./public.pem
```

The key is parsed once at middleware setup, not per request. (A future option is
to have Snowplow consume authn's JWKS endpoint directly for rotation; the current
`UserConfig` uses a static public key.)

## Migration notes (HS256 → RS256)

- The old shared secret (`JWT_SIGN_KEY` / `AUTHN_JWT_SECRET`) is gone. Replace the
  `jwt-sign-key` Secret's contents with the PEM private key described above.
- **Version alignment:** authn, Snowplow, and any other validator must all build
  against the `plumbing` version that carries the asymmetric `jwtutil`/`UserConfig`
  API. A validator still on the HS256 build will reject the new RS256 tokens.
- Snowplow's `--jwt-sign-key` / `JWT_SIGN_KEY` must be repointed to authn's PEM
  **public** key once it is rebuilt against the new `plumbing`.
