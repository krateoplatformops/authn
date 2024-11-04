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
    "path": "/ldap/login"
  },
  {
    "kind": "github",
    "name": "github-example",
    "path": "/github/login",
    "extensions": {
      "authCodeURL": "https://github.com/login/oauth/authorize?client_id=XXXX&redirect_uri=http%3A%2F%2Flocalhost%3A8888%2Fgithub%2Fgrant&response_type=code&scope=read%3Auser+read%3Aorg&state=YYYY",
      "redirectURL": "http://localhost:8888/github/grant"
    }
  },
  {
    "kind": "oidc",
    "name": "oidc-example",
    "path": "/oidc/login",
    "extensions": {
      "authCodeURL": "https://login.microsoftonline.com/XXXX/oauth2/v2.0/authorize?client_id=XXXX\u0026redirect_uri=http%3A%2F%2Flocalhost%3A8080%2Foidc%2Fcallbacl\u0026response_mode=query\u0026response_type=code\u0026scope=openid+email+profile+User.Read",
      "redirectURL": "http://localhost:8080/oidc/callback"
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

### Login with OAuth Authorization Code Flow (Github)

> Let's take _Github_ as example, the same concept applies to all authentication systems of this type (authorization code flow).

With a valid [authorization code grant](https://www.oauth.com/oauth2-servers/access-tokens/authorization-code-request/) invoke the related endpoint `path`, passing the `name` as query string parameter.

Example:

```sh
$ curl -H "X-Auth-Code: $(AUTH_CODE)" \
    https://api.krateoplatformops.io/authn/github/login?name=github-example
```

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
To obtain proper groups mappings you need to configure the ID Token response on the application side. Likewise for the profile picture. Examples are listed below for Azure and KeyCloak.

#### Azure
Azure can be configured to authenticate users through OIDC. To achieve this, you need to create a new app registration:
 - Go to "App registrations" and then hit "New registration";
 - Configure the display name, account types and Redirect URI. The redirect URI must point to Krateo's Authn;
 - Create a client secret in "Certificates & secrets", save the value of the secret now as it cannot by visualized afterwards;
 - In the "Authentication" menu, find and activate `Access tokens` and `ID tokens`;
 - In the "API permissions" menu, add the following: `openid`, `email`, `profile`, `User.Read` and `User.ReadBasic.All`;
 - To obtain groups in the OIDC ID Token response, modify the manifest value `groupMembershipClaims` to `all`;
 - To obtain the user avatar/profile image include `User.Read` in the `additionalScopes` field of the OIDCConfiguration custom resource;
 - You can now configure the Authn's CR by using Azure discovery URL, which will be in the following format:
 ```
https://login.microsoftonline.com/<your-tenant-id>/v2.0/.well-known/openid-configuration
 ```

#### KeyCloak
To obtain groups, add a custom mapper of type "Group Membership" and give it the Token Claim Name "groups", uncheck `Full group path`. Add `groups` into the `additionalScopes` field of the OIDCConfiguration custom resource.
To obtain the user avatar/profile image, go to the realm settings, then "User profiles" tab, "Create Attribute", and add one with the name `picture`. Set the profile picture for the user to a URL pointing to a picture. Keycloak will now return the avatar during authentication.

