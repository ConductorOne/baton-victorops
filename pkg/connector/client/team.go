package client

import (
	"context"
	"net/http"
)

func (c *VictorOpsClient) ListTeams(ctx context.Context) ([]Team, error) {
	var response []Team

	endPoint := c.getUrl(TeamsEndpoint)

	err := c.request(ctx, http.MethodGet, endPoint, &response, nil)
	if err != nil {
		return nil, err
	}

	return response, nil
}

func (c *VictorOpsClient) ListTeamMembers(ctx context.Context, teamId string) ([]TeamMember, error) {
	type Response struct {
		TeamMembers []TeamMember `json:"members"`
	}

	var response Response

	endPoint := c.getUrl(TeamMembersEndpoint)

	err := c.request(ctx, http.MethodGet, endPoint, &response, nil)
	if err != nil {
		return nil, err
	}

	return response.TeamMembers, nil
}

func (c *VictorOpsClient) ListTeamAdmins(ctx context.Context, teamId string) ([]TeamMemberAdmin, error) {
	type Response struct {
		TeamAdmins []TeamMemberAdmin `json:"teamAdmins"`
	}

	var response Response

	endPoint := c.getUrl(TeamAdminsEndpoint)

	err := c.request(ctx, http.MethodGet, endPoint, &response, nil)
	if err != nil {
		return nil, err
	}

	return response.TeamAdmins, nil
}
