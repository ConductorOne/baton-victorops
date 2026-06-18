package connector

import (
	"context"
	"fmt"

	v2 "github.com/conductorone/baton-sdk/pb/c1/connector/v2"
	ent "github.com/conductorone/baton-sdk/pkg/types/entitlement"
	"github.com/conductorone/baton-sdk/pkg/types/grant"
	rs "github.com/conductorone/baton-sdk/pkg/types/resource"
	"github.com/conductorone/baton-victorops/pkg/connector/client"
	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
)

const (
	scheduleMember = "member"
	scheduleOnCall = "on-call"
)

type scheduleBuilder struct {
	resourceType *v2.ResourceType
	client       *client.VictorOpsClient
}

func (s *scheduleBuilder) ResourceType(_ context.Context) *v2.ResourceType {
	return scheduleResourceType
}

func scheduleResource(onCallInfo *client.OnCallInfo, team *client.Team) (*v2.Resource, error) {
	displayName := titleCase(fmt.Sprintf("schedule for '%s' policy", onCallInfo.EscalationPolicy.Name))
	scheduleSlug := fmt.Sprintf("schedule_%s", onCallInfo.EscalationPolicy.Slug)
	profile := map[string]interface{}{
		"schedule_name":      displayName,
		"schedule_team_slug": team.Slug,
	}

	if len(onCallInfo.Users) > 0 {
		profile["schedule_oncall_users"] = scheduleMembersToInterfaceSlice(onCallInfo.Users)
	}

	resource, err := rs.NewGroupResource(
		displayName,
		scheduleResourceType,
		scheduleSlug,
		[]rs.GroupTraitOption{rs.WithGroupProfile(profile)},
		rs.WithParentResourceID(&v2.ResourceId{
			ResourceType: teamResourceType.Id,
			Resource:     team.Slug,
		}),
	)
	if err != nil {
		return nil, err
	}

	return resource, nil
}

func (s *scheduleBuilder) List(ctx context.Context, _ *v2.ResourceId, _ rs.SyncOpAttrs) ([]*v2.Resource, *rs.SyncOpResults, error) {
	listOnCallResponse, err := s.client.ListCurrentOnCallTeamsWithUsers(ctx)
	if err != nil {
		return nil, nil, fmt.Errorf("victorops-connector: failed to list schedules: %w", err)
	}

	var rv []*v2.Resource
	for _, teamsOnCall := range listOnCallResponse.TeamsOnCall {
		for _, onCallInfo := range teamsOnCall.OnCallNow {
			sr, err := scheduleResource(&onCallInfo, &teamsOnCall.Team)
			if err != nil {
				return nil, nil, fmt.Errorf("failed to create schedule resource: %w", err)
			}

			rv = append(rv, sr)
		}
	}

	return rv, nil, nil
}

func (s *scheduleBuilder) Entitlements(_ context.Context, resource *v2.Resource, _ rs.SyncOpAttrs) ([]*v2.Entitlement, *rs.SyncOpResults, error) {
	var rv []*v2.Entitlement

	oncallEntitlementOptions := []ent.EntitlementOption{
		ent.WithGrantableTo(userResourceType),
		ent.WithDisplayName(fmt.Sprintf("%s %s", scheduleOnCall, resource.DisplayName)),
		ent.WithDescription(fmt.Sprintf("%s VictorOps %s schedule entitlement", resource.DisplayName, scheduleOnCall)),
	}

	rv = append(
		rv,
		ent.NewAssignmentEntitlement(resource, scheduleOnCall, oncallEntitlementOptions...),
	)

	return rv, nil, nil
}

func (s *scheduleBuilder) Grants(ctx context.Context, resource *v2.Resource, opts rs.SyncOpAttrs) ([]*v2.Grant, *rs.SyncOpResults, error) {
	l := ctxzap.Extract(ctx)

	// parse resource profile to get schedule oncall users and grant them the oncall entitlement
	groupTrait, err := rs.GetGroupTrait(resource)
	if err != nil {
		return nil, nil, err
	}

	onCallUsers, ok := getProfileStringArray(groupTrait.Profile, "schedule_oncall_users")
	if !ok {
		l.Info("victorops-connector: no on-call users found for schedule resource")
	}

	teamSlug := groupTrait.Profile.Fields["schedule_team_slug"].GetStringValue()
	if teamSlug == "" {
		l.Info("victorops-connector: no team found for schedule resource")
	}

	var rv []*v2.Grant
	for _, onCallUser := range onCallUsers {
		rv = append(rv, grant.NewGrant(
			resource,
			scheduleOnCall,
			&v2.ResourceId{
				ResourceType: userResourceType.Id,
				Resource:     onCallUser,
			},
		))
	}

	return rv, nil, nil
}

func newScheduleBuilder(client *client.VictorOpsClient) *scheduleBuilder {
	return &scheduleBuilder{
		resourceType: scheduleResourceType,
		client:       client,
	}
}
