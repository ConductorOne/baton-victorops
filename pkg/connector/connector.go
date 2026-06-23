package connector

import (
	"context"
	"io"

	"github.com/conductorone/baton-victorops/pkg/connector/client"

	v2 "github.com/conductorone/baton-sdk/pb/c1/connector/v2"
	"github.com/conductorone/baton-sdk/pkg/annotations"
	"github.com/conductorone/baton-sdk/pkg/cli"
	"github.com/conductorone/baton-sdk/pkg/connectorbuilder"
	cfg "github.com/conductorone/baton-victorops/pkg/config"
)

type Connector struct {
	client *client.VictorOpsClient
}

// ResourceSyncers returns a ResourceSyncerV2 for each resource type that should be synced from the upstream service.
func (d *Connector) ResourceSyncers(ctx context.Context) []connectorbuilder.ResourceSyncerV2 {
	return []connectorbuilder.ResourceSyncerV2{
		newUserBuilder(d.client),
		newTeamBuilder(d.client),
		newScheduleBuilder(d.client),
	}
}

// Asset takes an input AssetRef and attempts to fetch it using the connector's authenticated http client
// It streams a response, always starting with a metadata object, following by chunked payloads for the asset.
func (d *Connector) Asset(ctx context.Context, asset *v2.AssetRef) (string, io.ReadCloser, error) {
	return "", nil, nil
}

// Metadata returns metadata about the connector.
func (d *Connector) Metadata(ctx context.Context) (*v2.ConnectorMetadata, error) {
	return &v2.ConnectorMetadata{
		DisplayName: "VictorOps",
		Description: "Baton VictorOps Connector",
	}, nil
}

// Validate is called to ensure that the connector is properly configured. It should exercise any API credentials
// to be sure that they are valid.
func (d *Connector) Validate(ctx context.Context) (annotations.Annotations, error) {
	return nil, nil
}

// New returns a new instance of the connector.
func New(ctx context.Context, clientId, apiKey, baseURL string) (*Connector, error) {
	opsClient, err := client.NewVictorOpsClient(ctx, clientId, apiKey, baseURL)
	if err != nil {
		return nil, err
	}

	return &Connector{
		client: opsClient,
	}, nil
}

// NewLambdaConnector returns a new ConnectorBuilderV2 for use in Lambda/containerized deployments.
func NewLambdaConnector(ctx context.Context, ac *cfg.Victorops, _ *cli.ConnectorOpts) (connectorbuilder.ConnectorBuilderV2, []connectorbuilder.Opt, error) {
	c, err := New(ctx, ac.VictoropsApiId, ac.VictoropsApiKey, ac.BaseUrl)
	if err != nil {
		return nil, nil, err
	}
	return c, nil, nil
}
