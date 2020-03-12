package main

import (
	"fmt"
	"testing"
)

func TestWeatherResponse_String(t *testing.T) {
	cases := []struct {
		wr   WeatherResponse
		want string
	}{
		{
			WeatherResponse{
				Name: "Berlin",
				Main: struct {
					Temp      float64 `json:"temp"`
					FeelsLike float64 `json:"feels_like"`
				}{
					9.07, 6.24,
				},
			},
			"Berlin, +9°",
		},
		{
			WeatherResponse{
				Name: "Berlin",
				Main: struct {
					Temp      float64 `json:"temp"`
					FeelsLike float64 `json:"feels_like"`
				}{
					12.94, 5.81,
				},
			},
			"Berlin, +13°",
		},
		{
			WeatherResponse{
				Name: "Berlin",
				Main: struct {
					Temp      float64 `json:"temp"`
					FeelsLike float64 `json:"feels_like"`
				}{
					-5.55, -12.2,
				},
			},
			"Berlin, -6°",
		},
	}

	for n, tc := range cases {
		t.Run(fmt.Sprintf("case=%d", n), func(t *testing.T) {
			if got := tc.wr.String(); got != tc.want {
				t.Errorf("WeatherResponse.String: want %q, got %q", tc.want, got)
			}
		})
	}
}
