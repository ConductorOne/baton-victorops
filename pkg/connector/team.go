package connector

import (
	"context"
	"errors"
	"fmt"
	"strings"

	rs "github.com/conductorone/baton-sdk/pkg/types/resource"
	"github.com/conductorone/baton-victorops/pkg/connector/client"

	v2 "github.com/conductorone/baton-sdk/pb/c1/connector/v2"
	"github.com/conductorone/baton-sdk/pkg/annotations"
	ent "github.com/conductorone/baton-sdk/pkg/types/entitlement"
	"github.com/conductorone/baton-sdk/pkg/types/grant"
	"github.com/conductorone/baton-sdk/pkg/uhttp"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var (
	teamMemberEntitlement = "member"
	teamAdminEntitlement  = "admin"
)

type teamBuilder struct {
	client                 *client.VictorOpsClient
	removalReplacementUser string
}

func (o *teamBuilder) ResourceType(ctx context.Context) *v2.ResourceType {
	return teamResourceType
}

// List returns all the teams from the database as resource objects.
func (o *teamBuilder) List(ctx context.Context, parentResourceID *v2.ResourceId, opts rs.SyncOpAttrs) ([]*v2.Resource, *rs.SyncOpResults, error) {
	teams, err := o.client.ListTeams(ctx)
	if err != nil {
		return nil, nil, fmt.Errorf("baton-victorops: failed to list teams: %w", err)
	}

	rv := make([]*v2.Resource, len(teams))
	for i, team := range teams {
		teamResourceP, err := teamResource(&team)
		if err != nil {
			return nil, nil, err
		}
		rv[i] = teamResourceP
	}

	return rv, nil, nil
}

func teamResource(team *client.Team) (*v2.Resource, error) {
	profile := map[string]interface{}{
		"name":            team.Name,
		"is_default_team": team.IsDefaultTeam,
		"slug":            team.Slug,
		"description":     team.Description,
		"member_count":    team.MemberCount,
		"version":         team.Version,
	}

	return rs.NewResource(
		team.Name,
		teamResourceType,
		team.Slug,
		rs.WithGroupTrait(),
		rs.WithResourceProfile(profile),
	)
}

// Entitlements returns entitlements for teams.
func (o *teamBuilder) Entitlements(_ context.Context, resource *v2.Resource, _ rs.SyncOpAttrs) ([]*v2.Entitlement, *rs.SyncOpResults, error) {
	var rv []*v2.Entitlement

	ents := []string{teamMemberEntitlement, teamAdminEntitlement}

	descriptions := map[string]string{
		teamMemberEntitlement: fmt.Sprintf("Member of %s team", resource.DisplayName),
		teamAdminEntitlement:  fmt.Sprintf("Admin of %s team", resource.DisplayName),
	}

	for _, value := range ents {
		assigmentOptions := []ent.EntitlementOption{
			ent.WithGrantableTo(userResourceType),
			ent.WithDisplayName(fmt.Sprintf("%s team %s", resource.DisplayName, value)),
			ent.WithDescription(descriptions[value]),
		}

		entitlement := ent.NewAssignmentEntitlement(resource, value, assigmentOptions...)
		rv = append(rv, entitlement)
	}

	return rv, nil, nil
}

// Grants returns grants for team members.
func (o *teamBuilder) Grants(ctx context.Context, resource *v2.Resource, opts rs.SyncOpAttrs) ([]*v2.Grant, *rs.SyncOpResults, error) {
	teamId := resource.Id.Resource

	listUsers, err := o.client.ListTeamMembers(ctx, teamId)
	if err != nil {
		return nil, nil, fmt.Errorf("baton-victorops: failed to list members of team %s: %w", teamId, err)
	}

	rv := make([]*v2.Grant, len(listUsers))
	for i, user := range listUsers {
		userId, err := rs.NewResourceID(userResourceType, user.Username)
		if err != nil {
			return nil, nil, err
		}

		userGrant := grant.NewGrant(resource, teamMemberEntitlement, userId)

		rv[i] = userGrant
	}

	adminUsers, err := o.client.ListTeamAdmins(ctx, teamId)
	if err != nil {
		return nil, nil, fmt.Errorf("baton-victorops: failed to list admins of team %s: %w", teamId, err)
	}

	for _, user := range adminUsers {
		userId, err := rs.NewResourceID(userResourceType, user.Username)
		if err != nil {
			return nil, nil, err
		}

		userGrant := grant.NewGrant(resource, teamAdminEntitlement, userId, grant.WithAnnotation(&v2.GrantImmutable{}))

		rv = append(rv, userGrant)
	}

	return rv, nil, nil
}

func (o *teamBuilder) Grant(ctx context.Context, principal *v2.Resource, entitlement *v2.Entitlement) ([]*v2.Grant, annotations.Annotations, error) {
	if entitlement.Slug != teamMemberEntitlement {
		return nil, nil, status.Errorf(codes.InvalidArgument, "baton-victorops: entitlement %s is not supported", entitlement.Slug)
	}

	teamId := entitlement.Resource.Id.Resource
	username := principal.Id.Resource

	newGrant := grant.NewGrant(entitlement.Resource, entitlement.Slug, principal.Id)

	err := o.client.AddUserTeam(ctx, teamId, username)
	if err != nil {
		if status.Code(err) == codes.AlreadyExists {
			return []*v2.Grant{newGrant}, annotations.New(&v2.GrantAlreadyExists{}), nil
		}
		return nil, nil, fmt.Errorf("baton-victorops: failed to add user %s to team %s: %w", username, teamId, err)
	}

	return []*v2.Grant{newGrant}, nil, nil
}

func (o *teamBuilder) Revoke(ctx context.Context, grant *v2.Grant) (annotations.Annotations, error) {
	if grant.Entitlement.Slug != teamMemberEntitlement {
		return nil, status.Errorf(codes.InvalidArgument, "baton-victorops: entitlement %s is not supported", grant.Entitlement.Slug)
	}

	teamId := grant.Entitlement.Resource.Id.Resource
	username := grant.Principal.Id.Resource

	onCall, err := o.client.IsOnCall(ctx, teamId, username)
	if err != nil {
		return nil, fmt.Errorf("baton-victorops: failed to check whether user %s is on call for team %s: %w", username, teamId, err)
	}

	// Not in the vendor docs, which tie the replacement to rotation membership: on a live tenant, removing a rotation
	// member who is not currently on call succeeds without a replacement, and a replacement sent for them is not
	// added to the rotation. So a replacement is sent, and required, only for users currently on call.
	// A user cannot replace themselves.
	replacement := ""
	if onCall {
		if o.removalReplacementUser == "" {
			return nil, status.Errorf(codes.FailedPrecondition,
				"baton-victorops: cannot remove user %s from team %s because they are currently on call and VictorOps requires a replacement; "+
					"set Removal replacement user (removal-replacement-user) in the connector configuration", username, teamId)
		}
		if strings.EqualFold(o.removalReplacementUser, username) {
			return nil, status.Errorf(codes.FailedPrecondition,
				"baton-victorops: cannot remove user %s from team %s because they are currently on call and VictorOps requires a replacement; "+
					"they are the configured Removal replacement user (removal-replacement-user), so set a different user in the connector configuration", username, teamId)
		}
		replacement = o.removalReplacementUser
	}

	// A 404 is documented only as team-not-found, so it is not treated as already revoked.
	err = o.client.RemoveUserTeam(ctx, teamId, username, replacement)
	if err != nil {
		if errors.Is(err, client.ErrInvalidReplacement) {
			// FailedPrecondition so the error is not retried as the vendor's 5xx.
			return nil, uhttp.WrapErrors(codes.FailedPrecondition,
				fmt.Sprintf("baton-victorops: failed to remove user %s from team %s, check that removal-replacement-user %s exists in VictorOps",
					username, teamId, o.removalReplacementUser), err)
		}
		return nil, fmt.Errorf("baton-victorops: failed to remove user %s from team %s: %w", username, teamId, err)
	}

	return nil, nil
}

func newTeamBuilder(client *client.VictorOpsClient, removalReplacementUser string) *teamBuilder {
	return &teamBuilder{
		client:                 client,
		removalReplacementUser: removalReplacementUser,
	}
}
