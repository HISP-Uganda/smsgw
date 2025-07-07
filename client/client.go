package client

import (
	"github.com/go-resty/resty/v2"
	"log"
)

type Client struct {
	RestClient *resty.Client
	BaseURL    string
	AuthToken  string
}

func NewClient(baseUrl, username, password, authToken string) *Client {
	client := resty.New()
	client.SetBaseURL(baseUrl)

	if authToken != "" {
		client.SetHeader("Authorization", "Token "+authToken)
	} else if username != "" && password != "" {
		client.SetBasicAuth(username, password)
	}

	client.SetHeaders(map[string]string{
		"Accept":       "application/json",
		"Content-Type": "application/json",
	})
	return &Client{
		RestClient: client,
		BaseURL:    baseUrl,
		AuthToken:  authToken,
	}
}

func (c *Client) GetResource(resourcePath string, params map[string]string) (*resty.Response, error) {
	request := c.RestClient.R()

	if params != nil {
		request.SetQueryParams(params)
	}

	resp, err := request.Get(resourcePath)
	if err != nil {
		log.Fatalf("Error when calling `GetResource`: %v", err)
	}
	return resp, err
}

func (c *Client) PostResource(resourcePath string, data interface{}) (*resty.Response, error) {
	resp, err := c.RestClient.R().
		SetHeader("Content-Type", "application/json").
		SetBody(data).
		Post(resourcePath)
	if err != nil {
		log.Fatalf("Error when calling `PostResource`: %v", err)
	}
	return resp, err
}
