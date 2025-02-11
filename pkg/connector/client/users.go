package client

import (
	"context"
	"net/http"
)

func (c *VictorOpsClient) ListUsers(ctx context.Context) ([]User, error) {
	var response []User

	endPoint := c.getUrl(UsersEndpoint)

	err := c.request(ctx, http.MethodGet, endPoint, &response, nil)
	if err != nil {
		return nil, err
	}

	return response, nil
}
