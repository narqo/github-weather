# :partly_sunny: Weather Status Widget for GitHub

![github-weather1](docs/github-weather-card.jpg)

[![Docker Pulls](https://img.shields.io/docker/pulls/varankinv/github-weather.svg)](https://hub.docker.com/r/varankinv/github-weather)
[![MIT licensed](https://img.shields.io/badge/license-MIT-blue.svg)](https://raw.githubusercontent.com/narqo/github-weather/master/LICENSE)

A silly program to update your GitHub user's [profile status](https://github.blog/changelog/2019-01-09-set-your-status/) with the current weather:

![github-weather2](docs/github-weather.jpg)

The program retrieves data from [OpenWeather API](https://openweathermap.org). It requires API key,
that can be obtained for free, following [OpenWeather API documentation][1].

To update user's status on GitHub, the program uses [GitHub's GraphQL API](https://developer.github.com/v4/) and requires API
token with `user` scope.

## Usage
- Add your desired values for following parameters:
  - Expiration Time of the new status in minutes (default 30 min, maximum 255 min)
  - GitHub Token
  - OpenWeather API Key
  - Query location (default Berlin,De)
- Compile and run the application with following commands:
```
$ go build -o github-weather .
$ ./github-weather
```

`github-weather -help` will print the list of all available options.

### Run the program as cronjob

```
*/10 * * * * github-weather [params] 2>> github-weather.log
```

See example of a crontab file in the project's `misc` directory.

### Run the program on Kubernetes

Refer to [deployments/README.md](./deployments/README.md).

## License

MIT

[1]: https://openweathermap.org/api
