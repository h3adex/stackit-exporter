package mocks

import (
	"context"

	"github.com/stackitcloud/stackit-sdk-go/services/ske"
)

type SkeMockClient struct {
	ClustersResponse        *ske.ListClustersResponse
	ProviderOptionsResponse *ske.ProviderOptions
}

type mockClusterListRequest struct {
	response *ske.ListClustersResponse
}

func (m *SkeMockClient) ListClusters(_ context.Context, _, _ string) ske.ApiListClustersRequest {
	return &mockClusterListRequest{response: m.ClustersResponse}
}

func (r *mockClusterListRequest) Execute() (*ske.ListClustersResponse, error) {
	return r.response, nil
}

type mockProviderOptionsRequest struct {
	response *ske.ProviderOptions
}

func (m *SkeMockClient) ListProviderOptions(_ context.Context, _ string) ske.ApiListProviderOptionsRequest {
	return &mockProviderOptionsRequest{response: m.ProviderOptionsResponse}
}

func (r *mockProviderOptionsRequest) Execute() (*ske.ProviderOptions, error) {
	return r.response, nil
}
