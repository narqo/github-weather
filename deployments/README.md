# Deploy on Kubernetes

## Requirements

You have to create two API tokens:

- Github API token with `user` scope.
- OpenWeather API token.

Having these tokens, create a Kubernetes secret as follows:

```bash
$ kubectl create secret generic github-weather \
    --from-literal=github_token=123_YOUR_GITHUB_TOKEN_HERE_321 \
    --from-literal=openweather_api_token=123_YOUR_OPEN_WEATHER_TOKEN_HERE_321
secret/github-weather created
```

*Replace with the correct values*

## Customize your city

By default, this cronjob gathers `Madrid` weather information. If you want to use another different city, just change
it in the [`cronjob.yaml`](cronjob.yaml) file.

```yaml
containers:
  - name: github-weather
    image: docker.io/varankinv/github-weather:v1.0.0
    command: ["/bin/sh", "-c"]
    args: ["/bin/github-weather --debug --github.token $(GITHUB_TOKEN) --owm.api-key $(OPENWEATHER_API_TOKEN) --owm.query Madrid"]
```

## Deploy it

Once the secret is placed in the cluster, deploy the cronjob:

```bash
$ kubectl apply -f cronjob.yaml
```

It is executed every ten minutes. Wait for it or...

## Test it

Just run 

```bash
$ kubectl create job test  --from cronjob/github-weather
job.batch/test created
$ kubectl logs -f job/test
2020/04/23 12:28:34 got owm response: {Cod:200 ID:3117735 Name:Madrid Weather:[{ID:801 Main:Clouds Description:few clouds Icon:02d}] Main:{Temp:17.4 FeelsLike:16.31}}
2020/04/23 12:28:34 >> variables: map[status:{github/weather 🌤️ 2020-04-23 12:58:34.365587279 +0000 UTC false Madrid, +17° }]
2020/04/23 12:28:34 >> query: 
        mutation ($status: ChangeUserStatusInput!) {
          changeUserStatus(input: $status) {
                status {
                  id
                  updatedAt
                  expiresAt
                }
          }
        }

2020/04/23 12:28:34 >> headers: map[Accept:[application/json; charset=utf-8] Authorization:[bearer ] Content-Type:[application/json; charset=utf-8]]
2020/04/23 12:28:34 << {"data":{"changeUserStatus":{"status":{"id":"1246618","updatedAt":"2020-04-23T12:28:34Z","expiresAt":"2020-04-23T12:58:34Z"}}}}
2020/04/23 12:28:34 set gh status: {ID:1246618 UpdatedAt:2020-04-23 12:28:34 +0000 UTC ExpiresAt:2020-04-23 12:58:34 +0000 UTC}
```

