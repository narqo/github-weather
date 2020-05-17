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

	"gopkg.in/yaml.v2"

	"github.com/machinebox/graphql"
)

type Config struct {
	Github struct {
		ClientID string `yaml:"clientID"`
		Endpoint string `yaml:"endpoint"`
		Token    string `yaml:"token"`
	} `yaml:"github"`
	Owm struct {
		ApiKey   string `yaml:"api_key"`
		Endpoint string `yaml:"endpoint"`
		Query    string `yaml:"query"`
	} `yaml:"owm"`
}

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
	var debug bool

	flags := flag.NewFlagSet("", flag.ExitOnError)
	flags.BoolVar(&debug, "debug", false, "Enable debug logging")

	if err := flags.Parse(args); err != nil {
		return err
	}

	f, err := os.Open("config.yml")
	if err != nil {
		return fmt.Errorf("error opening configuration file: %v", err)
	}
	defer f.Close()

	var cfg Config
	dec := yaml.NewDecoder(f)
	err = dec.Decode(&cfg)
	if err != nil {
		return fmt.Errorf("error parsing configuration file: %v", err)
	}

	if cfg.Owm.ApiKey == "" || cfg.Github.Token == "" {
		return fmt.Errorf("no API credentials passed: OpenWeather %q, GitHub %q", cfg.Owm.ApiKey, cfg.Github.Token)
	}

	owm := NewOWMClient(cfg.Owm.Endpoint, cfg.Owm.ApiKey)
	gh := NewGitHubClient(cfg.Github.Endpoint, cfg.Github.Token)
	if debug {
		gh.client.Log = func(s string) {
			log.Println(s)
		}
	}

	wr, err := owm.Weather(ctx, cfg.Owm.Query)
	if err != nil {
		return err
	}

	log.Printf("got owm response: %+v\n", wr)

	status := ChangeUserStatusInput{
		ClientMutationID: cfg.Github.ClientID,
		Emoji:            wr.Emoji(),
		Message:          wr.ShortString(),
		ExpiresAt:        time.Now().UTC().Add(30 * time.Minute), // XXX(narqo) status's expiration time is hardcoded
	}
	sr, err := gh.ChangeUserStatus(ctx, status)
	if err != nil {
		return err
	}

	log.Printf("set gh status: %+v\n", sr)

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

func (wr WeatherResponse) ShortString() string {
	var s strings.Builder

	s.WriteString(wr.Name)
	s.WriteByte(',')
	s.WriteByte(' ')

	if wr.Main.Temp > 0 {
		s.WriteByte('+')
	}
	s.WriteString(strconv.FormatFloat(wr.Main.Temp, 'f', 0, 64))
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
			return ":cloud:"
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
		return WeatherResponse{}, fmt.Errorf("weather API bad response, for %q: %+v", query, wr)
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
	ClientMutationID    string    `json:"clientMutationId,omitempty"`
	Emoji               string    `json:"emoji,omitempty"`
	ExpiresAt           time.Time `json:"expiresAt,omitempty"`
	LimitedAvailability bool      `json:"limitedAvailability,omitempty"`
	Message             string    `json:"message,omitempty"`
	OrganizationID      string    `json:"organizationId,omitempty"`
}

type ChangeUserStatusResponse struct {
	ID        string    `json:"id"`
	UpdatedAt time.Time `json:"updatedAt"`
	ExpiresAt time.Time `json:"expiresAt"`
}

const mutationChangeUserStatus = `
	mutation ($status: ChangeUserStatusInput!) {
	  changeUserStatus(input: $status) {
		status {
		  id
		  updatedAt
		  expiresAt
		}
	  }
	}
`

func (c *GitHubClient) ChangeUserStatus(ctx context.Context, input ChangeUserStatusInput) (ChangeUserStatusResponse, error) {
	req := graphql.NewRequest(mutationChangeUserStatus)
	req.Var("status", input)

	resp := struct {
		ChangeUserStatus struct {
			Status ChangeUserStatusResponse `json:"status"`
		} `json:"changeUserStatus"`
	}{}
	if err := c.run(ctx, req, &resp); err != nil {
		return ChangeUserStatusResponse{}, fmt.Errorf("github API request failed: %w", err)
	}

	status := resp.ChangeUserStatus.Status
	if status.UpdatedAt.Before(time.Now().UTC().Add(-time.Minute)) {
		return ChangeUserStatusResponse{}, fmt.Errorf("status not updated, github API response: %v", resp)
	}

	return status, nil
}

func (c *GitHubClient) run(ctx context.Context, req *graphql.Request, resp interface{}) error {
	if c.token != "" {
		req.Header.Add("Authorization", "bearer "+c.token)
	}
	return c.client.Run(ctx, req, resp)
}
