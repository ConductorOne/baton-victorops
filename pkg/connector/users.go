package connector

import (
	"context"

	"github.com/conductorone/baton-victorops/pkg/connector/client"

	v2 "github.com/conductorone/baton-sdk/pb/c1/connector/v2"
	"github.com/conductorone/baton-sdk/pkg/annotations"
	"github.com/conductorone/baton-sdk/pkg/pagination"
	rs "github.com/conductorone/baton-sdk/pkg/types/resource"
)

type userBuilder struct {
	client *client.VictorOpsClient
}

func (o *userBuilder) ResourceType(ctx context.Context) *v2.ResourceType {
	return userResourceType
}

// List returns all the users from the database as resource objects.
// Users include a UserTrait because they are the 'shape' of a standard user.
func (o *userBuilder) List(ctx context.Context, parentResourceID *v2.ResourceId, pToken *pagination.Token) ([]*v2.Resource, string, annotations.Annotations, error) {
	users, err := o.client.ListUsers(ctx)
	if err != nil {
		return nil, "", nil, err
	}

	rv := make([]*v2.Resource, len(users))

	for i, user := range users {
		us, err := userResource(ctx, &user)
		if err != nil {
			return nil, "", nil, err
		}
		rv[i] = us
	}

	return rv, "", nil, nil
}

func userResource(ctx context.Context, user *client.User) (*v2.Resource, error) {
	status := v2.UserTrait_Status_STATUS_ENABLED
	profile := map[string]interface{}{
		"first_name":            user.FirstName,
		"last_name":             user.LastName,
		"username":              user.Username,
		"email":                 user.Email,
		"created_at":            user.CreatedAt,
		"verified":              user.Verified,
		"password_last_updated": user.PasswordLastUpdated,
	}

	userTraitOptions := rs.WithUserTrait(
		rs.WithUserProfile(profile),
		rs.WithStatus(status),
		rs.WithEmail(user.Email, true),
	)

	resource, err := rs.NewResource(
		user.Username,
		userResourceType,
		user.Username,
		userTraitOptions,
	)

	if err != nil {
		return nil, err
	}

	return resource, nil
}

// Entitlements always returns an empty slice for users.
func (o *userBuilder) Entitlements(_ context.Context, resource *v2.Resource, _ *pagination.Token) ([]*v2.Entitlement, string, annotations.Annotations, error) {
	return nil, "", nil, nil
}

// Grants always returns an empty slice for users since they don't have any entitlements.
func (o *userBuilder) Grants(ctx context.Context, resource *v2.Resource, pToken *pagination.Token) ([]*v2.Grant, string, annotations.Annotations, error) {
	return nil, "", nil, nil
}

func newUserBuilder(client *client.VictorOpsClient) *userBuilder {
	return &userBuilder{
		client: client,
	}
}
