# simple-ingress-external-auth

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
   "client_id": "client0"
  },
  {
   "value": "NOX11CM2EP9xlP0HsS8NRPNHMmsQKQis7egKGcI+tHQ=",
   "client_id": "client1",
   "disable": true,
   "expires_at": "2022-07-04T14:21:22.52Z",
   "allowed_url": "https://custom.host.slok.dev/.*",
   "allowed_method": "(GET|POST)"
  },
  {
   "value": "6yvOSWrLmjC+2Vz8QdwHCjYoHyqWkD+70krxDt5XzlY=",
   "client_id": "client2",
   "allowed_method": "PUT"
  }
  {
   "value": "${TOKEN_CLIENT_3}",
   "client_id": "client3",
  }
 ]
}
```

### YAML example

```yaml
version: v1
tokens:
- value: 9bOlMT/vGlWCq56D+Ycgp7eTNj9uQWInbGf4tjRr/P8=
  client_id: client0
- value: NOX11CM2EP9xlP0HsS8NRPNHMmsQKQis7egKGcI+tHQ=
  client_id: client1
  disable: true
  expires_at: 2022-07-04T14:21:22.52Z
  allowed_url: https://custom.host.slok.dev/.*
  allowed_method: (GET|POST)
- value: 6yvOSWrLmjC+2Vz8QdwHCjYoHyqWkD+70krxDt5XzlY=
  client_id: client2
  allowed_method: PUT
- value: ${TOKEN_CLIENT_3}
  client_id: client3
```
