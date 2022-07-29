# simple-ingress-external-auth

[![CI](https://github.com/slok/simple-ingress-external-auth/actions/workflows/ci.yaml/badge.svg?branch=main)](https://github.com/slok/simple-ingress-external-auth/actions/workflows/ci.yaml)
[![Go Report Card](https://goreportcard.com/badge/github.com/slok/simple-ingress-external-auth)](https://goreportcard.com/report/github.com/slok/simple-ingress-external-auth)
[![Apache 2 licensed](https://img.shields.io/badge/license-Apache2-blue.svg)](https://raw.githubusercontent.com/slok/simple-ingress-external-auth/master/LICENSE)
[![GitHub release (latest SemVer)](https://img.shields.io/github/v/release/slok/simple-ingress-external-auth)](https://github.com/slok/simple-ingress-external-auth/releases/latest)

Easy and simple Kubernetes ingress authentication.

## How does it work

Most kubernetes ingress have a way of delegating the authentication to an external auth system.

To make this possible normally the ingress controller will forward the request to the external auth system (this auth app), and the auth app will return a 200 if its authenticated, and different than 200 if its not.

When it starts, this application will load a configuration file where it has all the tokens defined (and some other optional properties).

When the ingress-controller forwards the request, this app will check for `Authorization: Bearer <token>` header and validate against the tokens it has defined.

Examples of ingress controllers configurations for external auth:

- [ingress-nginx external authentication](https://kubernetes.github.io/ingress-nginx/user-guide/nginx-configuration/annotations/#external-authentication).
- [traefik forward authentication](https://doc.traefik.io/traefik/v2.0/middlewares/forwardauth/).

## Features

- Simple and easy to deploy (no complex setup, no databases...).
- Ability to rotate tokens (create a new token and add expiration date to the old one).
- Authenticate Kubernetes ingress easily.
- Fast and scalable (everything is in memory).
- Advanced token validation properties (expire date, disable...).
- Can be used with GRPC (e.g [ingress-nginx grpc](https://kubernetes.github.io/ingress-nginx/examples/grpc/))
- Different configuration formats (including env vars substitution support).

## Example

```bash
$  docker run --rm -it -p 8080:8080 -p 8081:8081 ghcr.io/slok/simple-ingress-external-auth --token-config-data='{"version": "v1","tokens": [{"value": "6kXEuNEWMYcd1yP16HsgrA=="}]}'
INFO[0000] Tokens loaded                                 addr=":8080" app=simple-ingress-external-auth svc=memory.TokenRepository tokens=1 version=dev
INFO[0000] http server listening for requests            addr=":8080" app=simple-ingress-external-auth version=dev
INFO[0000] http server listening for requests            addr=":8081" app=simple-ingress-external-auth health-check=/status metrics=/metrics pprof=/debug/pprof version=dev
```

```bash
$ curl -I -H "Authorization: Bearer 1234567890" http://127.0.0.1:8080/auth
HTTP/1.1 401 Unauthorized
Date: Mon, 20 Jun 2022 05:39:41 GMT
Content-Length: 13
Content-Type: text/plain; charset=utf-8

curl -I -H "Authorization: Bearer 6kXEuNEWMYcd1yP16HsgrA==" http://127.0.0.1:8080/auth
HTTP/1.1 200 OK
Date: Mon, 20 Jun 2022 05:39:50 GMT
```

## Token format

There is no restriction on the token format, for this application, it's just an string. You can use `1234567890` (please don't) or a JWT token.

An easy and portable way of generating tokens, would be using the old well known `openssl`, e.g:

```bash
$ openssl rand -base64 32
gmMCgSWCDzuBKxznnH7+vCajFnhRIK1+sTRvGJI2g1I=
```

## Advanced optional properties

Apart from regular token validation, we can use different optional properties:

- `client_id`: Not a security option, but used as metadata, for debugging/auditing purposes and token identification.
- `disable`: Will disable the token, handy when we want to disable temporally a token.
- `expires_at`: After the specified timestamp (RFC3339) the token will be invalid. Handy to rotate tokens.
- `allowed_url`: Regex that will validate the original URL being requested (Got from `X-Original-URL` header).
- `allowed_method`: Regex that will validate the original method being requested (Got from `X-Original-Method` header).

## Configuration

The tokens that the application will load will be provisioned with a configuration file (simple and portable). It has some features:

- JSON and YAML.
- Env vars substitution (`${X_Y_Z}` style).

### JSON example

```json
{
 "version": "v1",
 "tokens": [
  {
   "value": "9bOlMT/vGlWCq56D+Ycgp7eTNj9uQWInbGf4tjRr/P8=",
   "client_id": "test1"
  },
  {
   "value": "NOX11CM2EP9xlP0HsS8NRPNHMmsQKQis7egKGcI+tHQ=",
   "client_id": "test2",
   "disable": true,
   "expires_at": "2022-07-04T14:21:22.52Z",
   "allowed_url": "https://custom.host.slok.dev/.*",
   "allowed_method": "(GET|POST)"
  },
  {
   "value": "6yvOSWrLmjC+2Vz8QdwHCjYoHyqWkD+70krxDt5XzlY=",
   "client_id": "test3",
   "allowed_method": "PUT"
  }
  {
   "value": "${TOKEN_CLIENT_3}",
   "client_id": "test4"
  }
 ]
}
```

### YAML example

```yaml
version: v1
tokens:
- value: 9bOlMT/vGlWCq56D+Ycgp7eTNj9uQWInbGf4tjRr/P8=
  client_id: "test1"
- value: NOX11CM2EP9xlP0HsS8NRPNHMmsQKQis7egKGcI+tHQ=
  client_id: "test2"
  disable: true
  expires_at: 2022-07-04T14:21:22.52Z
  allowed_url: https://custom.host.slok.dev/.*
  allowed_method: (GET|POST)
- value: 6yvOSWrLmjC+2Vz8QdwHCjYoHyqWkD+70krxDt5XzlY=
  client_id: "test3"
  allowed_method: PUT
- value: ${TOKEN_CLIENT_3}
  client_id: "test4"
```

### As a Go API

You can access the configuration as a Go library so you can automate easily the creation of the configuration file.

Import as:

```go
package main

import (
 apiv1 "github.com/slok/simple-ingress-external-auth/pkg/api/v1"
)

func main() {
 //...
}
```

[Check this example of generating 1000k tokens](examples/config-generator).

## Why tokens and not basic auth

Although basic auth is simple and easy to use for web UIs, they are not secure. Tokens on the other side yes:

- Unique: The generated tokens can easily be unique (e.g: use `openssl` or [JWT](https://jwt.io/) to generate the tokens).
- Revocable: Tokens can have an expiration date, removed or temporally disabled. This gives us the ability to rotate them easily.
- Limited: Tokens can be narrowly scoped to specific URLs.
- Random: As with the `Unique` property, the tokens generated can be randomly generated easily with tools like `openssl`.
