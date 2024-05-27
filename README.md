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

