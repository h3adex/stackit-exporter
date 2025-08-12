package collector_test

/*import (
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
			Id:            mocks.Ptr("server1-id"),
			Name:          mocks.Ptr("server-1"),
			AffinityGroup: mocks.Ptr("group1"),
			BootVolume: &iaas.ServerGetBootVolumeRetType{
				Id: mocks.Ptr("123"),
			},
			AvailabilityZone: mocks.Ptr("eu01-1"),
			KeypairName:      mocks.Ptr("keypair1"),
			MachineType:      mocks.Ptr("c1.2"),
			ImageId:          mocks.Ptr("123"),
			Status:           mocks.Ptr("ACTIVE"),
			PowerStatus:      mocks.Ptr("RUNNING"),
			MaintenanceWindow: &iaas.ServerMaintenance{
				StartsAt: mocks.Ptr(start1),
				EndsAt:   mocks.Ptr(end1),
				Status:   mocks.Ptr("PLANNED"),
				Details:  mocks.Ptr("cve-123"),
			},
		},
		{
			Id:               mocks.Ptr("server2-id"),
			Name:             mocks.Ptr("server-2"),
			AvailabilityZone: mocks.Ptr("eu01-2"),
			AffinityGroup:    mocks.Ptr("group2"),
			ImageId:          mocks.Ptr("123"),
			BootVolume: &iaas.ServerGetBootVolumeRetType{
				Id: mocks.Ptr("123"),
			},
			KeypairName: mocks.Ptr("keypair2"),
			MachineType: mocks.Ptr("c1.4"),
			Status:      mocks.Ptr("INACTIVE"),
			PowerStatus: mocks.Ptr("STOPPED"),
			MaintenanceWindow: &iaas.ServerMaintenance{
				StartsAt: mocks.Ptr(start2),
				EndsAt:   mocks.Ptr(end2),
				Status:   mocks.Ptr("ONGOING"),
				Details:  mocks.Ptr("cve-123"),
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

	require.NoError(t, testRegistry.Register(reg.ServerInfo))

	// Register all metrics in the registry
	for _, gaugeVec := range reg.MaintenanceStatus {
		require.NoError(t, testRegistry.Register(gaugeVec))
	}
	for _, gaugeVec := range reg.ServerPowerStatus {
		require.NoError(t, testRegistry.Register(gaugeVec))
	}
	for _, gaugeVec := range reg.ServerStatus {
		require.NoError(t, testRegistry.Register(gaugeVec))
	}

	require.NoError(t, testRegistry.Register(reg.ServerLastSeen))
	require.NoError(t, testRegistry.Register(reg.MaintenanceStart))
	require.NoError(t, testRegistry.Register(reg.MaintenanceEnd))

	ctx := context.Background()
	collector.ScrapeIaasAPI(ctx, client, "", reg)

	const expected = `
# HELP stackit_server_info Descriptive info about the server at export time. Set to 1 if present.
# TYPE stackit_server_info gauge
stackit_server_info{affinity_group="group1",boot_volume_id="123",created_at="",image_id="123",keypair_name="keypair1",launched_at="",machine_type="c1.2",maintenance="PLANNED",maintenance_details="cve-123",maintenance_status="PLANNED",name="server-1",power_status="RUNNING",project_id="",server_id="server1-id",server_status="ACTIVE",zone="eu01-1"} 1
stackit_server_info{affinity_group="group2",boot_volume_id="123",created_at="",image_id="123",keypair_name="keypair2",launched_at="",machine_type="c1.4",maintenance="ONGOING",maintenance_details="cve-123",maintenance_status="ONGOING",name="server-2",power_status="STOPPED",project_id="",server_id="server2-id",server_status="INACTIVE",zone="eu01-2"} 1
# HELP stackit_server_maintenance_end_timestamp Unix timestamp of maintenance end time
# TYPE stackit_server_maintenance_end_timestamp gauge
stackit_server_maintenance_end_timestamp{machine_type="c1.2",name="server-1",project_id="",server_id="server1-id",zone="eu01-1"} 1.7100036e+09
stackit_server_maintenance_end_timestamp{machine_type="c1.4",name="server-2",project_id="",server_id="server2-id",zone="eu01-2"} 1.7100108e+09
# HELP stackit_server_maintenance_ongoing Binary maintenance status. 1 if current state is ONGOING
# TYPE stackit_server_maintenance_ongoing gauge
stackit_server_maintenance_ongoing{machine_type="c1.2",name="server-1",project_id="",server_id="server1-id",zone="eu01-1"} 0
stackit_server_maintenance_ongoing{machine_type="c1.4",name="server-2",project_id="",server_id="server2-id",zone="eu01-2"} 1
# HELP stackit_server_maintenance_planned Binary maintenance status. 1 if current state is PLANNED
# TYPE stackit_server_maintenance_planned gauge
stackit_server_maintenance_planned{machine_type="c1.2",name="server-1",project_id="",server_id="server1-id",zone="eu01-1"} 1
stackit_server_maintenance_planned{machine_type="c1.4",name="server-2",project_id="",server_id="server2-id",zone="eu01-2"} 0
# HELP stackit_server_maintenance_start_timestamp Unix timestamp of maintenance start time
# TYPE stackit_server_maintenance_start_timestamp gauge
stackit_server_maintenance_start_timestamp{machine_type="c1.2",name="server-1",project_id="",server_id="server1-id",zone="eu01-1"} 1.71e+09
stackit_server_maintenance_start_timestamp{machine_type="c1.4",name="server-2",project_id="",server_id="server2-id",zone="eu01-2"} 1.7100072e+09
# HELP stackit_server_power_running Binary power status metric. 1 if current power state is RUNNING
# TYPE stackit_server_power_running gauge
stackit_server_power_running{machine_type="c1.2",name="server-1",project_id="",server_id="server1-id",zone="eu01-1"} 1
stackit_server_power_running{machine_type="c1.4",name="server-2",project_id="",server_id="server2-id",zone="eu01-2"} 0
# HELP stackit_server_power_stopped Binary power status metric. 1 if current power state is STOPPED
# TYPE stackit_server_power_stopped gauge
stackit_server_power_stopped{machine_type="c1.2",name="server-1",project_id="",server_id="server1-id",zone="eu01-1"} 0
stackit_server_power_stopped{machine_type="c1.4",name="server-2",project_id="",server_id="server2-id",zone="eu01-2"} 1
# HELP stackit_server_status_active Binary server status metric. 1 if current state is ACTIVE
# TYPE stackit_server_status_active gauge
stackit_server_status_active{machine_type="c1.2",name="server-1",project_id="",server_id="server1-id",zone="eu01-1"} 1
stackit_server_status_active{machine_type="c1.4",name="server-2",project_id="",server_id="server2-id",zone="eu01-2"} 0
# HELP stackit_server_status_inactive Binary server status metric. 1 if current state is INACTIVE
# TYPE stackit_server_status_inactive gauge
stackit_server_status_inactive{machine_type="c1.2",name="server-1",project_id="",server_id="server1-id",zone="eu01-1"} 0
stackit_server_status_inactive{machine_type="c1.4",name="server-2",project_id="",server_id="server2-id",zone="eu01-2"} 1
`

	err := testutil.GatherAndCompare(testRegistry, strings.NewReader(expected),
		"stackit_server_info",
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
*/
