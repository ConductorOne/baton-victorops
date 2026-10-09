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
	var response listTeamMembersResponse

	endPoint := c.getUrl(fmt.Sprintf(TeamMembersEndpoint, pathSegment(teamId)))

	err := c.request(ctx, http.MethodGet, endPoint, &response, nil)
	if err != nil {
		return nil, err
	}

	return response.TeamMembers, nil
}

func (c *VictorOpsClient) ListTeamAdmins(ctx context.Context, teamId string) ([]TeamMemberAdmin, error) {
	var response listTeamAdminsResponse

	endPoint := c.getUrl(fmt.Sprintf(TeamAdminsEndpoint, pathSegment(teamId)))

	err := c.request(ctx, http.MethodGet, endPoint, &response, nil)
	if err != nil {
		return nil, err
	}

	return response.TeamAdmins, nil
}

func (c *VictorOpsClient) AddUserTeam(ctx context.Context, teamId, username string) error {
	body := addTeamMemberBody{
		Username: username,
	}

	endPoint := c.getUrl(fmt.Sprintf(AddTeamMemberEndpoint, pathSegment(teamId)))

	err := c.request(ctx, http.MethodPost, endPoint, nil, body)
	if err != nil {
		return err
	}

	return nil
}

// RemoveUserTeam removes username from the team. The API requires a replacement when the user is on call
// (rotations or escalation policies), and answers 422 when the replacement is not valid.
func (c *VictorOpsClient) RemoveUserTeam(ctx context.Context, teamId, username, replacement string) error {
	body := removeTeamMemberBody{
		Replacement: replacement,
	}

	endPoint := c.getUrl(fmt.Sprintf(RemoveTeamMemberEndpoint, pathSegment(teamId), pathSegment(username)))

	statusCode, err := c.doRequest(ctx, http.MethodDelete, endPoint, nil, body)
	if err != nil {
		if statusCode == http.StatusUnprocessableEntity {
			return fmt.Errorf("%w: %w", ErrInvalidReplacement, err)
		}
		return err
	}

	return nil
}

func (c *VictorOpsClient) ListCurrentOnCallTeamsWithUsers(ctx context.Context) (*TeamsOnCallResponse, error) {
	response := &TeamsOnCallResponse{}

	endPoint := c.getUrl(OnCallCurrentEndpoint)

	err := c.request(ctx, http.MethodGet, endPoint, response, nil)
	if err != nil {
		return nil, err
	}

	return response, nil
}
