package connector

import (
	"context"
	"fmt"

	rs "github.com/conductorone/baton-sdk/pkg/types/resource"
	"github.com/conductorone/baton-victorops/pkg/connector/client"

	v2 "github.com/conductorone/baton-sdk/pb/c1/connector/v2"
	"github.com/conductorone/baton-sdk/pkg/annotations"
	"github.com/conductorone/baton-sdk/pkg/pagination"
	ent "github.com/conductorone/baton-sdk/pkg/types/entitlement"
)

var (
	teamMemberEntitlement = "member"
	teamAdminEntitlement  = "admin"
)

type teamBuilder struct {
	client *client.VictorOpsClient
}

func (o *teamBuilder) ResourceType(ctx context.Context) *v2.ResourceType {
	return teamResourceType
}

// List returns all the users from the database as resource objects.
// Users include a UserTrait because they are the 'shape' of a standard user.
func (o *teamBuilder) List(ctx context.Context, parentResourceID *v2.ResourceId, pToken *pagination.Token) ([]*v2.Resource, string, annotations.Annotations, error) {
	teams, err := o.client.ListTeams(ctx)
	if err != nil {
		return nil, "", nil, err
	}

	rv := make([]*v2.Resource, len(teams))
	for i, team := range teams {
		teamResourceP, err := teamResource(&team)
		if err != nil {
			return nil, "", nil, err
		}
		rv[i] = teamResourceP
	}

	return rv, "", nil, nil
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

	teamTraitOptions := rs.WithGroupTrait(
		rs.WithGroupProfile(profile),
	)

	return rs.NewResource(
		team.Name,
		teamResourceType,
		team.Name,
		teamTraitOptions,
	)
}

// Entitlements always returns an empty slice for users.
func (o *teamBuilder) Entitlements(_ context.Context, resource *v2.Resource, _ *pagination.Token) ([]*v2.Entitlement, string, annotations.Annotations, error) {
	var rv []*v2.Entitlement

	ents := []string{teamMemberEntitlement, teamAdminEntitlement}

	for _, value := range ents {
		assigmentOptions := []ent.EntitlementOption{
			ent.WithGrantableTo(userResourceType),
			ent.WithDisplayName(fmt.Sprintf("%s team %s", resource.DisplayName, value)),
			ent.WithDescription(fmt.Sprintf("Member of %s team", resource.DisplayName)),
		}

		entitlement := ent.NewAssignmentEntitlement(resource, value, assigmentOptions...)
		rv = append(rv, entitlement)
	}

	return rv, "", nil, nil
}

// Grants always returns an empty slice for users since they don't have any entitlements.
func (o *teamBuilder) Grants(ctx context.Context, resource *v2.Resource, pToken *pagination.Token) ([]*v2.Grant, string, annotations.Annotations, error) {
	return nil, "", nil, nil
}

func newTeamBuilder(client *client.VictorOpsClient) *teamBuilder {
	return &teamBuilder{
		client: client,
	}
}
