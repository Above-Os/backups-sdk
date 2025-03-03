package net

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"time"

	"github.com/go-resty/resty/v2"
)

func Post[T any](url string, headers map[string]string, data interface{}, debug bool, insecureSkipVerify bool) (*T, error) {
	var result T
	client := resty.New().SetTimeout(10 * time.Second).
		SetTLSClientConfig(&tls.Config{InsecureSkipVerify: insecureSkipVerify}).R().SetDebug(debug)

	if headers != nil {
		client.SetHeaders(headers)
	}

	resp, err := client.
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
