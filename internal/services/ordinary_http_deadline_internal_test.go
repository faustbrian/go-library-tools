package services

import (
	"context"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"
)

type ordinaryHTTPRoundTrip func(*http.Request) (*http.Response, error)

func (trip ordinaryHTTPRoundTrip) RoundTrip(request *http.Request) (*http.Response, error) {
	return trip(request)
}

func TestOrdinaryFixtureHTTPDeadline(t *testing.T) {
	for _, earlier := range []bool{false, true} {
		t.Run(map[bool]string{false: "background", true: "earlier parent"}[earlier], func(t *testing.T) {
			ctx := context.Background()
			var expected time.Time
			if earlier {
				expected = time.Now().Add(time.Minute / 60)
				var cancel context.CancelFunc
				ctx, cancel = context.WithDeadline(ctx, expected)
				defer cancel()
			}
			client := fixtureHTTPClient()
			before := time.Now()
			client.Transport = ordinaryHTTPRoundTrip(func(request *http.Request) (*http.Response, error) {
				deadline, ok := request.Context().Deadline()
				switch {
				case !ok:
					t.Error("operational request has no deadline")
				case earlier && !deadline.Equal(expected):
					t.Error("earlier caller deadline changed")
				case !earlier && (deadline.Before(before.Add(5*time.Second)) || deadline.After(time.Now().Add(5*time.Second))):
					t.Error("operational deadline is not five seconds from request admission")
				}
				return &http.Response{StatusCode: http.StatusNoContent, Body: io.NopCloser(strings.NewReader("")), Header: make(http.Header)}, nil
			})
			request, err := http.NewRequestWithContext(ctx, http.MethodGet, "http://fixture.invalid/", nil)
			if err != nil {
				t.Fatal(err)
			}
			response, err := client.Do(request)
			if err != nil || response.StatusCode != http.StatusNoContent {
				t.Fatalf("immediate response = %v, %v", response, err)
			}
			response.Body.Close()
		})
	}
}

func TestOrdinaryFixtureHTTPCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	client := fixtureHTTPClient()
	client.Transport = ordinaryHTTPRoundTrip(func(request *http.Request) (*http.Response, error) {
		cancel()
		return nil, request.Context().Err()
	})
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, "http://fixture.invalid/", nil)
	if err != nil {
		t.Fatal(err)
	}
	response, err := client.Do(request)
	if response != nil || !errors.Is(err, context.Canceled) {
		t.Fatalf("canceled request = %v, %v", response, err)
	}
}
