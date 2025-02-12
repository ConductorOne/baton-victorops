package client

import (
	"context"
	"fmt"
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

	endPoint := c.getUrl(fmt.Sprintf(TeamMembersEndpoint, teamId))

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

	endPoint := c.getUrl(fmt.Sprintf(TeamMembersEndpoint, teamId))

	err := c.request(ctx, http.MethodGet, endPoint, &response, nil)
	if err != nil {
		return nil, err
	}

	return response.TeamAdmins, nil
}

func (c *VictorOpsClient) AddUserTeam(ctx context.Context, teamId, username string) error {
	type Body struct {
		Username string `json:"username"`
	}

	body := Body{
		Username: username,
	}

	endPoint := c.getUrl(fmt.Sprintf(AddTeamMemberEndpoint, teamId))

	err := c.request(ctx, http.MethodPost, endPoint, nil, body)
	if err != nil {
		return err
	}

	return nil
}

func (c *VictorOpsClient) RemoveUserTeam(ctx context.Context, teamId, username string) error {
	endPoint := c.getUrl(fmt.Sprintf(RemoveTeamMemberEndpoint, teamId, username))

	err := c.request(ctx, http.MethodPost, endPoint, nil, nil)
	if err != nil {
		return err
	}

	return nil
}
