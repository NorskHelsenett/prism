```
docker run --rm -it --net=host -v ./api/config.yaml:/config/ -v ./api/.tmp:/app/.tmp -e GO_ENV=dev prism
```

.well-known/config.json
```
{
  "apiEndpoint": "http://localhost:8080",
  "providers": [
    {
      "name": "Helsegitlab",
      "type": "gitlab"
    },
    {
      "name": "Microsoft AD",
      "type": "azure"
    },
    {
      "name": "Mocc IdP",
      "type": "mocc"
    }
  ]
}
```

./.tmp/config.yaml
```
oidc:
  mocc:
    clientID: "prism-local-client"
    clientSecret: "prism-local-secret"
    redirectUri: "http://localhost:8080/api/callback"
    providerUri: "http://localhost:9999"
cors:
  origin: "http://localhost:8080"
admins:
  - alice.admin@test.local
database:
  path: "./.tmp"
events:
  interval: 10
slack:
  token: ""
  webhookUrl: ""
secrets:
  HMAC_SECRET_KEY: "DO_NOT_LET_ME_CATCH_YOU_USING_THIS_IN_PROD"
```