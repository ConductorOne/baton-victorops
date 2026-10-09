package client

import (
	"context"
	"net/http"
)

func (c *VictorOpsClient) ListUsers(ctx context.Context) ([]User, error) {
	var response listUsersResponse

	endPoint := c.getUrl(UsersEndpoint)

	err := c.request(ctx, http.MethodGet, endPoint, &response, nil)
	if err != nil {
		return nil, err
	}

	var userResponse []User

	for _, users := range response.Users {
		userResponse = append(userResponse, users...)
	}

	return userResponse, nil
}
