package movie

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/masaya-nishimura-09/movie-log-api/internal/domain/exception"
)

type TMDBClient struct {
	client   *http.Client
	endpoint *url.URL
	token    string
}

func NewTMDBClient(endpoint *url.URL, token string) *TMDBClient {
	return &TMDBClient{
		client:   &http.Client{Timeout: 5 * time.Second},
		endpoint: endpoint,
		token:    token,
	}
}

func (c *TMDBClient) Get(
	ctx context.Context,
	path string,
	query url.Values,
) ([]byte, error) {
	u := c.endpoint.JoinPath(path)
	u.RawQuery = query.Encode()

	req, err := http.NewRequestWithContext(
		ctx, http.MethodGet, u.String(), nil,
	)
	if err != nil {
		return nil, fmt.Errorf("create TMDB request: %w", err)
	}
	req.Header.Set("accept", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.token)

	res, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request TMDB: %w", err)
	}
	defer res.Body.Close()

	if res.StatusCode == http.StatusNotFound {
		return nil, exception.ErrNotFound
	}
	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("TMDB returned status %d", res.StatusCode)
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("read TMDB response: %w", err)
	}

	return body, nil
}
