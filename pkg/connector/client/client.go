package client

import (
	"context"

	"github.com/conductorone/baton-sdk/pkg/uhttp"
	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
)

type VictorOpsClient struct {
	client   *uhttp.BaseHttpClient
	apiKey   string
	clientId string
}

func NewVictorOpsClient(ctx context.Context, clientId, apiKey string) (*VictorOpsClient, error) {
	httpClient, err := uhttp.NewClient(ctx, uhttp.WithLogger(true, ctxzap.Extract(ctx)))
	if err != nil {
		return nil, err
	}

	return &VictorOpsClient{
		client:   uhttp.NewBaseHttpClient(httpClient),
		clientId: clientId,
		apiKey:   apiKey,
	}, nil
}
