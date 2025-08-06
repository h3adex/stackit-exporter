package collector_test

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/h3adex/stackit-exporter/internal/collector"
	"github.com/h3adex/stackit-exporter/internal/metrics"
	"github.com/h3adex/stackit-exporter/test/mocks"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/stackitcloud/stackit-sdk-go/services/iaas"
	"github.com/stretchr/testify/require"
)

func mockMultipleServers() *[]iaas.Server {
	start1 := time.Unix(1710000000, 0)
	end1 := time.Unix(1710003600, 0)
	start2 := time.Unix(1710007200, 0)
	end2 := time.Unix(1710010800, 0)

	return &[]iaas.Server{
		{
			Id:               mocks.Ptr("server1-id"),
			Name:             mocks.Ptr("server-1"),
			AvailabilityZone: mocks.Ptr("eu01-1"),
			MachineType:      mocks.Ptr("c1.2"),
			Status:           mocks.Ptr("ACTIVE"),
			PowerStatus:      mocks.Ptr("RUNNING"),
			MaintenanceWindow: &iaas.ServerMaintenance{
				StartsAt: mocks.Ptr(start1),
				EndsAt:   mocks.Ptr(end1),
				Status:   mocks.Ptr("PLANNED"),
			},
		},
		{
			Id:               mocks.Ptr("server2-id"),
			Name:             mocks.Ptr("server-2"),
			AvailabilityZone: mocks.Ptr("eu01-2"),
			MachineType:      mocks.Ptr("c1.4"),
			Status:           mocks.Ptr("STOPPED"),
			PowerStatus:      mocks.Ptr("SHUTOFF"),
			MaintenanceWindow: &iaas.ServerMaintenance{
				StartsAt: mocks.Ptr(start2),
				EndsAt:   mocks.Ptr(end2),
				Status:   mocks.Ptr("ONGOING"),
			},
		},
	}
}

func TestScrapeIaasAPI_PopulatesMetrics(t *testing.T) {
	client := &mocks.IaasMockClient{
		Response: &iaas.ServerListResponse{
			Items: mockMultipleServers(),
		},
	}

	reg := metrics.NewIaasRegistry()
	testRegistry := prometheus.NewRegistry()

	require.NoError(t, testRegistry.Register(reg.MaintenanceStart))
	require.NoError(t, testRegistry.Register(reg.MaintenanceEnd))
	require.NoError(t, testRegistry.Register(reg.MaintenanceStatus))
	require.NoError(t, testRegistry.Register(reg.ServerStatus))
	require.NoError(t, testRegistry.Register(reg.ServerPowerStatus))
	require.NoError(t, testRegistry.Register(reg.ServerLastSeen))

	ctx := context.Background()
	collector.ScrapeIaasAPI(ctx, client, "", reg)

	const expected = `
# HELP stackit_server_maintenance_end_timestamp Scheduled maintenance window end time (Unix timestamp)
# TYPE stackit_server_maintenance_end_timestamp gauge
stackit_server_maintenance_end_timestamp{machine_type="c1.2",name="server-1",project_id="",server_id="server1-id",zone="eu01-1"} 1.7100036e+09
stackit_server_maintenance_end_timestamp{machine_type="c1.4",name="server-2",project_id="",server_id="server2-id",zone="eu01-2"} 1.7100108e+09
# HELP stackit_server_maintenance_start_timestamp Scheduled maintenance window start time (Unix timestamp)
# TYPE stackit_server_maintenance_start_timestamp gauge
stackit_server_maintenance_start_timestamp{machine_type="c1.2",name="server-1",project_id="",server_id="server1-id",zone="eu01-1"} 1.71e+09
stackit_server_maintenance_start_timestamp{machine_type="c1.4",name="server-2",project_id="",server_id="server2-id",zone="eu01-2"} 1.7100072e+09
# HELP stackit_server_maintenance_status Current maintenance status (label: status; always 1)
# TYPE stackit_server_maintenance_status gauge
stackit_server_maintenance_status{machine_type="c1.2",name="server-1",project_id="",server_id="server1-id",status="PLANNED",zone="eu01-1"} 1
stackit_server_maintenance_status{machine_type="c1.4",name="server-2",project_id="",server_id="server2-id",status="ONGOING",zone="eu01-2"} 1
# HELP stackit_server_power_status Current server power status (label: power_status; always 1)
# TYPE stackit_server_power_status gauge
stackit_server_power_status{name="server-1",power_status="RUNNING",project_id="",server_id="server1-id",zone="eu01-1"} 1
stackit_server_power_status{name="server-2",power_status="SHUTOFF",project_id="",server_id="server2-id",zone="eu01-2"} 1
# HELP stackit_server_status Current server status (label: status; always 1)
# TYPE stackit_server_status gauge
stackit_server_status{name="server-1",project_id="",server_id="server1-id",status="ACTIVE",zone="eu01-1"} 1
stackit_server_status{name="server-2",project_id="",server_id="server2-id",status="STOPPED",zone="eu01-2"} 1
`

	err := testutil.GatherAndCompare(testRegistry, strings.NewReader(expected),
		"stackit_server_status",
		"stackit_server_power_status",
		"stackit_server_maintenance_start_timestamp",
		"stackit_server_maintenance_end_timestamp",
		"stackit_server_maintenance_status",
	)

	require.NoError(t, err)
}
