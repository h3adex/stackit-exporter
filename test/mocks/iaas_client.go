package mocks

import (
	"context"

	"github.com/stackitcloud/stackit-sdk-go/services/iaas"
)

type IaasMockClient struct {
	Response *iaas.ServerListResponse
}

type mockServerListResponse struct {
	response *iaas.ServerListResponse
}

func (r *mockServerListResponse) Details(_ bool) iaas.ApiListServersRequest {
	panic("not implemented")
}

func (r *mockServerListResponse) LabelSelector(_ string) iaas.ApiListServersRequest {
	panic("not implemented")
}

func (f *IaasMockClient) ListServers(_ context.Context, _ string) iaas.ApiListServersRequest {
	return &mockServerListResponse{response: f.Response}
}

func (r *mockServerListResponse) Execute() (*iaas.ServerListResponse, error) {
	return r.response, nil
}
