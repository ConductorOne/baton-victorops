package client

import (
	"context"
	"net/http"
	"net/url"

	"github.com/conductorone/baton-sdk/pkg/uhttp"
	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
)

var (
	BaseUrl = "https://api.victorops.com"

	UsersEndpoint = "/api-public/v1/user"

	TeamsEndpoint       = "/api-public/v1/team"
	TeamMembersEndpoint = "/api-public/v1/team/%s/members"
	TeamAdminsEndpoint  = "/api-public/v1/team/%s/admins"

	AddTeamMemberEndpoint    = "/api-public/v1/team/%s/members"
	RemoveTeamMemberEndpoint = "/api-public/v1/team/%s/members/%s"
)

type VictorOpsClient struct {
	httpClient *uhttp.BaseHttpClient
	apiKey     string
	clientId   string
	baseUrl    *url.URL
}

func NewVictorOpsClient(ctx context.Context, clientId, apiKey string) (*VictorOpsClient, error) {
	httpClient, err := uhttp.NewClient(ctx, uhttp.WithLogger(true, ctxzap.Extract(ctx)))
	if err != nil {
		return nil, err
	}

	baseUrl, err := url.Parse(BaseUrl)
	if err != nil {
		return nil, err
	}

	return &VictorOpsClient{
		httpClient: uhttp.NewBaseHttpClient(httpClient),
		clientId:   clientId,
		apiKey:     apiKey,
		baseUrl:    baseUrl,
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
	var (
		resp *http.Response
		err  error
	)

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
		return err
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

	if err != nil {
		return err
	}

	return nil
}
