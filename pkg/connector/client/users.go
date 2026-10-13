package client

import (
	"context"
	"fmt"
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

// UserExists reports whether username exists. The API answers 404 for an unknown user.
func (c *VictorOpsClient) UserExists(ctx context.Context, username string) (bool, error) {
	endPoint := c.getUrl(fmt.Sprintf(UserEndpoint, pathSegment(username)))

	statusCode, err := c.doRequest(ctx, http.MethodGet, endPoint, nil, nil)
	if err != nil {
		if statusCode == http.StatusNotFound {
			return false, nil
		}
		return false, err
	}

	return true, nil
}
