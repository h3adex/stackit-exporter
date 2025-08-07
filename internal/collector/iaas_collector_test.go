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

// mockMultipleServers creates a slice of iaas.Server with predefined data for testing.
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
			Status:           mocks.Ptr("INACTIVE"),
			PowerStatus:      mocks.Ptr("STOPPED"),
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

	// Register all metrics in the registry
	for _, gaugeVec := range reg.MaintenanceStatus {
		require.NoError(t, testRegistry.Register(gaugeVec))
	}
	for _, gaugeVec := range reg.PowerStatus {
		require.NoError(t, testRegistry.Register(gaugeVec))
	}
	for _, gaugeVec := range reg.ServerStatus {
		require.NoError(t, testRegistry.Register(gaugeVec))
	}

	require.NoError(t, testRegistry.Register(reg.LastSeen))
	require.NoError(t, testRegistry.Register(reg.MaintenanceStart))
	require.NoError(t, testRegistry.Register(reg.MaintenanceEnd))

	ctx := context.Background()
	collector.ScrapeIaasAPI(ctx, client, "", reg)

	const expected = `
# HELP stackit_server_maintenance_end_timestamp Unix time when the maintenance window ends
# TYPE stackit_server_maintenance_end_timestamp gauge
stackit_server_maintenance_end_timestamp{machine_type="c1.2",name="server-1",project_id="",server_id="server1-id",zone="eu01-1"} 1.7100036e+09
stackit_server_maintenance_end_timestamp{machine_type="c1.4",name="server-2",project_id="",server_id="server2-id",zone="eu01-2"} 1.7100108e+09
# HELP stackit_server_maintenance_start_timestamp Unix time when the maintenance window starts
# TYPE stackit_server_maintenance_start_timestamp gauge
stackit_server_maintenance_start_timestamp{machine_type="c1.2",name="server-1",project_id="",server_id="server1-id",zone="eu01-1"} 1.71e+09
stackit_server_maintenance_start_timestamp{machine_type="c1.4",name="server-2",project_id="",server_id="server2-id",zone="eu01-2"} 1.7100072e+09
# HELP stackit_server_maintenance_planned Binary state of maintenance status: 1 if PLANNED, else 0
# TYPE stackit_server_maintenance_planned gauge
stackit_server_maintenance_planned{machine_type="c1.2",name="server-1",project_id="",server_id="server1-id",zone="eu01-1"} 1
stackit_server_maintenance_planned{machine_type="c1.4",name="server-2",project_id="",server_id="server2-id",zone="eu01-2"} 0
# HELP stackit_server_maintenance_ongoing Binary state of maintenance status: 1 if ONGOING, else 0
# TYPE stackit_server_maintenance_ongoing gauge
stackit_server_maintenance_ongoing{machine_type="c1.2",name="server-1",project_id="",server_id="server1-id",zone="eu01-1"} 0
stackit_server_maintenance_ongoing{machine_type="c1.4",name="server-2",project_id="",server_id="server2-id",zone="eu01-2"} 1
# HELP stackit_server_power_running Binary state of power status: 1 if RUNNING, else 0
# TYPE stackit_server_power_running gauge
stackit_server_power_running{machine_type="c1.2",name="server-1",project_id="",server_id="server1-id",zone="eu01-1"} 1
stackit_server_power_running{machine_type="c1.4",name="server-2",project_id="",server_id="server2-id",zone="eu01-2"} 0
# HELP stackit_server_power_stopped Binary state of power status: 1 if STOPPED, else 0
# TYPE stackit_server_power_stopped gauge
stackit_server_power_stopped{machine_type="c1.4",name="server-2",project_id="",server_id="server2-id",zone="eu01-2"} 1
stackit_server_power_stopped{machine_type="c1.2",name="server-1",project_id="",server_id="server1-id",zone="eu01-1"} 0
# HELP stackit_server_status_active Binary state of server status: 1 if ACTIVE, else 0
# TYPE stackit_server_status_active gauge
stackit_server_status_active{machine_type="c1.2",name="server-1",project_id="",server_id="server1-id",zone="eu01-1"} 1
stackit_server_status_active{machine_type="c1.4",name="server-2",project_id="",server_id="server2-id",zone="eu01-2"} 0
# HELP stackit_server_status_inactive Binary state of server status: 1 if INACTIVE, else 0
# TYPE stackit_server_status_inactive gauge
stackit_server_status_inactive{machine_type="c1.2",name="server-1",project_id="",server_id="server1-id",zone="eu01-1"} 0
stackit_server_status_inactive{machine_type="c1.4",name="server-2",project_id="",server_id="server2-id",zone="eu01-2"} 1
`

	err := testutil.GatherAndCompare(testRegistry, strings.NewReader(expected),
		"stackit_server_status_active",
		"stackit_server_status_inactive",
		"stackit_server_power_running",
		"stackit_server_power_stopped",
		"stackit_server_maintenance_start_timestamp",
		"stackit_server_maintenance_end_timestamp",
		"stackit_server_maintenance_planned",
		"stackit_server_maintenance_ongoing",
	)

	require.NoError(t, err)
}
