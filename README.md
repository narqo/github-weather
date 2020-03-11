# :partly_sunny: Weather Status Widget for GitHub

A silly program to update a Github user's status with current weather.

Note, the program retrieves data from [OpenWeather API](https://openweathermap.org/current). It requires OpenWeather API key,
that can be obtained for free, following [OpenWeather API documentation][1].

To update user's status, the program uses [GitHub's GraphQL API](https://developer.github.com/v4/) and requires API
token with `user` scope.

## Usage

```
$ go build -o github-weather ./
$ github-weather -github.token <gh-api-token> -owm.api-key <open-weather-api-token> -owm.query <City>
```

`github-weather -help` will print the list of all available options.

### Run the program as cronjob

```
0 * * * * /usr/local/bin/github-weather [params]
```

[1]: https://openweathermap.org/api
