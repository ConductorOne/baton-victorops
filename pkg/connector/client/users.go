package client

import (
	"context"
	"net/http"
)

func (c *VictorOpsClient) ListUsers(ctx context.Context) ([]User, error) {
	type Response struct {
		Users [][]User `json:"users"`
	}

	var response Response

	endPoint := c.getUrl(UsersEndpoint)

	err := c.request(ctx, http.MethodGet, endPoint, &response, nil)
	if err != nil {
		return nil, err
	}

	var userResponse []User

	for _, users := range response.Users {
		for _, user := range users {
			userResponse = append(userResponse, user)
		}
	}

	return userResponse, nil
}
