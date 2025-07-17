package utils

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"time"

	"github.com/go-resty/resty/v2"
)

func Post[T any](ctx context.Context, url string, headers map[string]string, data interface{}) (*T, error) {
	var result T
	client := resty.New().SetTimeout(60 * time.Second).
		SetTLSClientConfig(&tls.Config{InsecureSkipVerify: true}).R().SetDebug(true)

	if headers != nil {
		client.SetHeaders(headers)
	}

	resp, err := client.SetContext(ctx).
		SetBody(data).
		SetResult(&result).
		Post(url)

	if err != nil {
		return nil, err
	}

	if resp.StatusCode() != http.StatusOK {
		return nil, fmt.Errorf("request failed, status code: %d", resp.StatusCode())
	}

	return &result, nil
}
