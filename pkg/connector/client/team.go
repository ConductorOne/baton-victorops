package client

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/conductorone/baton-sdk/pkg/uhttp"
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

// RemoveUserTeam removes username from the team. An empty replacement sends none, which the API rejects with a 500
// when the user is on call. The live API also answers 500 for an unknown replacement (documented as 422),
// so a 500 is attributed to the replacement only after a lookup confirms it.
func (c *VictorOpsClient) RemoveUserTeam(ctx context.Context, teamId, username, replacement string) error {
	body := removeTeamMemberBody{
		Replacement: replacement,
	}

	endPoint := c.getUrl(fmt.Sprintf(RemoveTeamMemberEndpoint, pathSegment(teamId), pathSegment(username)))

	statusCode, err := c.doRequest(ctx, http.MethodDelete, endPoint, nil, body)
	if err != nil {
		switch statusCode {
		case http.StatusUnprocessableEntity:
			return fmt.Errorf("%w: %w", ErrInvalidReplacement, err)
		case http.StatusInternalServerError:
			if replacement == "" {
				return err
			}
			if exists, lookupErr := c.UserExists(ctx, replacement); lookupErr == nil && !exists {
				return fmt.Errorf("%w: %w", ErrInvalidReplacement, err)
			}
		}
		return err
	}

	return nil
}

// IsOnCall reports whether username is currently on call for the team.
func (c *VictorOpsClient) IsOnCall(ctx context.Context, teamId, username string) (bool, error) {
	var response TeamsOnCallResponse

	endPoint := c.getUrl(OnCallCurrentEndpoint)

	_, err := c.doRequest(ctx, http.MethodGet, endPoint, &response, nil, uhttp.WithNoCache())
	if err != nil {
		return false, err
	}

	for _, team := range response.TeamsOnCall {
		if team.Team.Slug != teamId {
			continue
		}
		for _, onCall := range team.OnCallNow {
			for _, user := range onCall.Users {
				if strings.EqualFold(user.OnCallUser.Username, username) {
					return true, nil
				}
			}
		}
	}

	return false, nil
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
