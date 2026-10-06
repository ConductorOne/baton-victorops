package client

import (
	"context"
	"errors"
	"net/http"
	"net/url"

	"github.com/conductorone/baton-sdk/pkg/uhttp"
	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
)

var (
	DefaultBaseURL = "https://api.victorops.com"

	UsersEndpoint = "/api-public/v1/user"

	TeamsEndpoint       = "/api-public/v1/team"
	TeamMembersEndpoint = "/api-public/v1/team/%s/members"
	TeamAdminsEndpoint  = "/api-public/v1/team/%s/admins"

	AddTeamMemberEndpoint    = "/api-public/v1/team/%s/members"
	RemoveTeamMemberEndpoint = "/api-public/v1/team/%s/members/%s"
	OnCallCurrentEndpoint    = "/api-public/v1/oncall/current"
)

var ErrInvalidReplacement = errors.New("replacement user was not found or is not valid")

type VictorOpsClient struct {
	httpClient *uhttp.BaseHttpClient
	apiKey     string
	clientId   string
	baseUrl    *url.URL
}

func NewVictorOpsClient(ctx context.Context, clientId, apiKey, baseURL string) (*VictorOpsClient, error) {
	httpClient, err := uhttp.NewClient(ctx, uhttp.WithLogger(true, ctxzap.Extract(ctx)))
	if err != nil {
		return nil, err
	}

	if baseURL == "" {
		baseURL = DefaultBaseURL
	}

	parsedURL, err := url.Parse(baseURL)
	if err != nil {
		return nil, err
	}

	return &VictorOpsClient{
		httpClient: uhttp.NewBaseHttpClient(httpClient),
		clientId:   clientId,
		apiKey:     apiKey,
		baseUrl:    parsedURL,
	}, nil
}

func (c *VictorOpsClient) getUrl(endPoint string) *url.URL {
	return c.baseUrl.JoinPath(endPoint)
}

func (c *VictorOpsClient) request(
	ctx context.Context,
	method string,
	urlAddress *url.URL,
	res interface{},
	body interface{},
) error {
	_, err := c.doRequest(ctx, method, urlAddress, res, body)
	return err
}

// doRequest behaves like request and also returns the HTTP status code, or 0 if no response was received.
func (c *VictorOpsClient) doRequest(
	ctx context.Context,
	method string,
	urlAddress *url.URL,
	res interface{},
	body interface{},
) (int, error) {
	var resp *http.Response

	options := []uhttp.RequestOption{
		uhttp.WithHeader("X-VO-Api-Id", c.clientId),
		uhttp.WithHeader("X-VO-Api-Key", c.apiKey),
	}

	if body != nil {
		options = append(options, uhttp.WithJSONBody(body))
	}

	req, err := c.httpClient.NewRequest(
		ctx,
		method,
		urlAddress,
		options...,
	)
	if err != nil {
		return 0, err
	}

	switch method {
	case http.MethodGet:
		resp, err = c.httpClient.Do(req, uhttp.WithResponse(&res))
		if resp != nil {
			defer resp.Body.Close()
		}
	case http.MethodPost, http.MethodPatch, http.MethodDelete:
		resp, err = c.httpClient.Do(req)
		if resp != nil {
			defer resp.Body.Close()
		}
	}

	statusCode := 0
	if resp != nil {
		statusCode = resp.StatusCode
	}

	return statusCode, err
}
