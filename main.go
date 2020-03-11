package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/machinebox/graphql"
)

const (
	owmAPIEndpoint    = "https://api.openweathermap.org/data/2.5/weather?appid={api-key}&units=metric"
	githubAPIEndpoint = "https://api.github.com/graphql"
	githubClientID    = "github/weather"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigc := make(chan os.Signal, 2)
	signal.Notify(sigc, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-sigc
		cancel()
	}()

	if err := run(ctx, os.Args[1:]); err != nil {
		log.Fatal(err)
	}
}

func run(ctx context.Context, args []string) error {
	flags := flag.NewFlagSet("", flag.ExitOnError)

	var (
		owmAPIKey string
		owmQuery  string

		githubAPIToken string
	)
	flags.StringVar(&owmAPIKey, "owm.api-key", "", "OpenWeather API key")
	flags.StringVar(&owmQuery, "owm.query", "Berlin,de", "OpenWeather API query, city name, state and country code divided by comma")

	flags.StringVar(&githubAPIToken, "github.token", "", "GitHub API token")

	if err := flags.Parse(args); err != nil {
		return err
	}

	if owmAPIKey == "" || githubAPIToken == "" {
		return fmt.Errorf("no API credentials passed: OpenWeather %q, GitHub %q", owmAPIKey, githubAPIToken)
	}

	owm := NewOWMClient(owmAPIEndpoint, owmAPIKey)
	gh := NewGitHubClient(githubAPIEndpoint, githubAPIToken)

	wr, err := owm.Weather(ctx, owmQuery)
	if err != nil {
		return err
	}

	log.Printf("got owm response: %v\n", wr)

	status := ChangeUserStatusInput{
		ClientMutationID: githubClientID,
		Emoji:            wr.Emoji(),
		Message:          wr.String(),
		ExpiresAt:        time.Now().Add(time.Hour).Add(5 * time.Minute),
	}
	sr, err := gh.ChangeUserStatus(ctx, status)
	if err != nil {
		return err
	}

	log.Printf("set gh status: %v\n", sr)

	return nil
}

type OWMClient struct {
	apiURL string
	client *http.Client
}

func NewOWMClient(apiURL, apiKey string) *OWMClient {
	return &OWMClient{
		apiURL: strings.Replace(apiURL, "{api-key}", apiKey, 1),
		client: &http.Client{},
	}
}

type WeatherResponse struct {
	Cod     int    `json:"cod"`
	ID      int    `json:"id"`
	Name    string `json:"name"`
	Weather []struct {
		ID          int    `json:"id"`
		Main        string `json:"main"`
		Description string `json:"description"`
		Icon        string `json:"icon"`
	} `json:"weather"`
	Main struct {
		Temp      float64 `json:"temp"`
		FeelsLike float64 `json:"feels_like"`
	} `json:"main"`
}

func (wr WeatherResponse) String() string {
	var s strings.Builder

	s.WriteString(wr.Name)
	s.WriteByte(',')
	s.WriteByte(' ')

	if wr.Main.Temp > 0 {
		s.WriteByte('+')
	}
	s.WriteString(strconv.FormatFloat(wr.Main.Temp, 'g', 0, 64))
	s.WriteString("°") // WriteString as "degree" is not from ASCII

	return s.String()
}

// Emoji maps OpenWeather weather status to emojis.
// See https://openweathermap.org/weather-conditions
func (wr WeatherResponse) Emoji() string {
	if len(wr.Weather) == 0 {
		return ":zap:"
	}

	w := wr.Weather[0]
	if w.ID == 800 {
		if w.Icon == "01n" {
			return ":full_moon:"
		}
		return ":sunny:"
	}
	if w.ID > 800 {
		switch w.ID {
		case 801:
			return "🌤️"
		case 802:
			return ":cloudy:"
		default:
			return ":partly_sunny:"
		}
	} else if w.ID >= 700 {
		return ":foggy:"
	} else if w.ID >= 600 {
		return ":snowflake:"
	} else if w.ID >= 500 {
		if w.ID == 500 {
			return "🌦️"
		}
		if w.ID >= 511 {
			return "🌨️"
		}
		return "☔"
	} else if w.ID >= 300 {
		return "🌦️"
	} else if w.ID >= 200 {
		return "⛈️"
	}

	return ":zap:"
}

func (c *OWMClient) Weather(ctx context.Context, query string) (WeatherResponse, error) {
	u := c.apiURL + "&q=" + query
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return WeatherResponse{}, err
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return WeatherResponse{}, fmt.Errorf("weather API request failed, query %q: %w", query, err)
	}
	defer resp.Body.Close()

	var wr WeatherResponse
	if err := json.NewDecoder(resp.Body).Decode(&wr); err != nil {
		return WeatherResponse{}, err
	}

	if wr.Cod != 200 {
		return WeatherResponse{}, fmt.Errorf("wearther API bad response, for %q: %+v", query, wr)
	}

	return wr, nil
}

type GitHubClient struct {
	apiURL string
	token  string
	client *graphql.Client
}

func NewGitHubClient(apiURL, token string, opts ...graphql.ClientOption) *GitHubClient {
	return &GitHubClient{
		apiURL: apiURL,
		token:  token,
		client: graphql.NewClient(apiURL, opts...),
	}
}

type ChangeUserStatusInput struct {
	ClientMutationID    string    `json:"clientMutationId"`
	Emoji               string    `json:"emoji"`
	ExpiresAt           time.Time `json:"expiresAt"`
	LimitedAvailability bool      `json:"limitedAvailability"`
	Message             string    `json:"message"`
	OrganizationID      string    `json:"organizationId"`
}

type ChangeUserStatusResponse struct {
	UpdatedAt time.Time `json:"updatedAt"`
	ExpiresAt time.Time `json:"expiresAt"`
}

const mutationChangeUserStatus = `
	mutation ChangeUserStatus($clientMutationId: String!, $emoji: String!, $message: String!, $expiresAt: DateTime) {
	  changeUserStatus(input: {clientMutationId: $clientMutationId, emoji: $emoji, message: $message, expiresAt: $expiresAt}) {
		status {
		  updatedAt
		  expiresAt
		}
	  }
	}
`

func (c *GitHubClient) ChangeUserStatus(ctx context.Context, input ChangeUserStatusInput) (ChangeUserStatusResponse, error) {
	req := graphql.NewRequest(mutationChangeUserStatus)
	req.Var("clientMutationId", input.ClientMutationID)
	req.Var("emoji", input.Emoji)
	req.Var("message", input.Message)
	req.Var("expiresAt", input.ExpiresAt)

	resp := struct {
		ChangeUserStatus struct {
			Status ChangeUserStatusResponse
		} `json:"changeUserStatus"`
	}{}
	if err := c.run(ctx, req, &resp); err != nil {
		return ChangeUserStatusResponse{}, err
	}
	return resp.ChangeUserStatus.Status, nil
}

func (c *GitHubClient) run(ctx context.Context, req *graphql.Request, resp interface{}) error {
	if c.token != "" {
		req.Header.Add("Authorization", "bearer "+c.token)
	}
	return c.client.Run(ctx, req, resp)
}
