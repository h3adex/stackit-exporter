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
	"github.com/stackitcloud/stackit-sdk-go/services/ske"
	"github.com/stretchr/testify/require"
)

func mockSkeCluster() *ske.ListClustersResponse {
	creationTime := time.Unix(1744629614, 0) // Example creation time

	start1 := time.Unix(1710000000, 0) // Example maintenance window start
	end1 := time.Unix(1710003600, 0)   // Example maintenance window end

	return &ske.ListClustersResponse{
		Items: &[]ske.Cluster{
			{
				Name: mocks.Ptr("ske-mock-01"),
				Kubernetes: &ske.Kubernetes{
					Version: mocks.Ptr("1.30.10"),
				},
				Maintenance: &ske.Maintenance{
					AutoUpdate: &ske.MaintenanceAutoUpdate{
						KubernetesVersion:   mocks.Ptr(false),
						MachineImageVersion: mocks.Ptr(true),
					},
					TimeWindow: &ske.TimeWindow{
						Start: &start1,
						End:   &end1,
					},
				},
				Nodepools: &[]ske.Nodepool{
					{
						Name: mocks.Ptr("default-pool"),
						Machine: &ske.Machine{
							Type: mocks.Ptr("c1.4"),
							Image: &ske.Image{
								Name:    mocks.Ptr("flatcar"),
								Version: mocks.Ptr("4152.2.3"),
							},
						},
						AvailabilityZones: &[]string{"eu01-01", "eu01-02"},
						Volume:            &ske.Volume{Size: mocks.Ptr(int64(50))},
					},
					{
						Name: mocks.Ptr("gpu-l40s"),
						Machine: &ske.Machine{
							Type: mocks.Ptr("n2.14d.g1"),
							Image: &ske.Image{
								Name:    mocks.Ptr("ubuntu"),
								Version: mocks.Ptr("2204.20250516.0"),
							},
						},
						AvailabilityZones: &[]string{"eu01-01"},
						Volume:            &ske.Volume{Size: mocks.Ptr(int64(100))},
					},
				},
				Status: &ske.ClusterStatus{
					Aggregated:          ske.ClusterStatusGetAggregatedAttributeType(mocks.Ptr("STATE_HEALTHY")),
					CreationTime:        &creationTime,
					EgressAddressRanges: &[]string{"10.0.0.1/32"},
					Errors:              &[]ske.ClusterError{},
				},
			},
		},
	}
}

func mockProviderOptions() *ske.ProviderOptions {
	return &ske.ProviderOptions{
		KubernetesVersions: &[]ske.KubernetesVersion{
			{
				State:   mocks.Ptr("supported"),
				Version: mocks.Ptr("1.30.10"),
			},
		},
		MachineImages: &[]ske.MachineImage{
			{
				Name: mocks.Ptr("flatcar"),
				Versions: &[]ske.MachineImageVersion{
					{
						State:   mocks.Ptr("supported"),
						Version: mocks.Ptr("4152.2.3"),
					},
				},
			},
			{
				Name: mocks.Ptr("ubuntu"),
				Versions: &[]ske.MachineImageVersion{
					{
						State:   mocks.Ptr("supported"),
						Version: mocks.Ptr("2204.20250516.0"),
					},
				},
			},
		},
	}
}

func TestScrapeSkeAPI_PopulatesMetrics(t *testing.T) {
	client := &mocks.SkeMockClient{
		ClustersResponse:        mockSkeCluster(),
		ProviderOptionsResponse: mockProviderOptions(),
	}

	reg := metrics.NewSKERegistry()
	registry := prometheus.NewRegistry()

	// Register all necessary metrics
	for _, vec := range reg.ClusterStatus {
		require.NoError(t, registry.Register(vec))
	}
	for _, vec := range reg.K8sVersion {
		require.NoError(t, registry.Register(vec))
	}
	for _, vec := range reg.NodePoolMachineVersion {
		require.NoError(t, registry.Register(vec))
	}
	require.NoError(t, registry.Register(reg.ClusterErrorStatus))
	require.NoError(t, registry.Register(reg.ClusterCreationTime))
	require.NoError(t, registry.Register(reg.EgressAddressRanges))
	require.NoError(t, registry.Register(reg.MaintenanceWindowStart))
	require.NoError(t, registry.Register(reg.MaintenanceWindowEnd))
	require.NoError(t, registry.Register(reg.NodePoolMachineTypes))
	require.NoError(t, registry.Register(reg.NodePoolVolumeSizes))
	require.NoError(t, registry.Register(reg.NodePoolAvailabilityZones))
	require.NoError(t, registry.Register(reg.MaintenanceAutoUpdate))

	ctx := context.Background()
	projectID := "" // Example without projectID
	collector.ScrapeSkeAPI(ctx, client, projectID, "eu01", reg)

	const expected = `
# HELP stackit_ske_cluster_creation_timestamp Unix timestamp when cluster was created
# TYPE stackit_ske_cluster_creation_timestamp gauge
stackit_ske_cluster_creation_timestamp{cluster_name="ske-mock-01",project_id=""} 1.744629614e+09
# HELP stackit_ske_cluster_egress_address_range Egress IP range used by cluster
# TYPE stackit_ske_cluster_egress_address_range gauge
stackit_ske_cluster_egress_address_range{cidr="10.0.0.1/32",cluster_name="ske-mock-01",project_id=""} 1
# HELP stackit_ske_cluster_error_status 1 if cluster has error
# TYPE stackit_ske_cluster_error_status gauge
stackit_ske_cluster_error_status{cluster_name="ske-mock-01",project_id=""} 0
# HELP stackit_ske_cluster_maintenance_autoupdate_enabled 1 if autoupdate is enabled
# TYPE stackit_ske_cluster_maintenance_autoupdate_enabled gauge
stackit_ske_cluster_maintenance_autoupdate_enabled{cluster_name="ske-mock-01",project_id=""} 0
# HELP stackit_ske_cluster_maintenance_end_timestamp End time of maintenance window
# TYPE stackit_ske_cluster_maintenance_end_timestamp gauge
stackit_ske_cluster_maintenance_end_timestamp{cluster_name="ske-mock-01",project_id=""} 1.7100036e+09
# HELP stackit_ske_cluster_maintenance_start_timestamp Start time of maintenance window
# TYPE stackit_ske_cluster_maintenance_start_timestamp gauge
stackit_ske_cluster_maintenance_start_timestamp{cluster_name="ske-mock-01",project_id=""} 1.71e+09
# HELP stackit_ske_cluster_status_state_healthy Cluster status: STATE_HEALTHY
# TYPE stackit_ske_cluster_status_state_healthy gauge
stackit_ske_cluster_status_state_healthy{cluster_name="ske-mock-01",project_id=""} 1
# HELP stackit_ske_k8s_version_supported Kubernetes version state: supported
# TYPE stackit_ske_k8s_version_supported gauge
stackit_ske_k8s_version_supported{cluster_name="ske-mock-01",k8s_version="1.30.10",project_id=""} 1
# HELP stackit_ske_nodepool_availability_zone Availability zones configured
# TYPE stackit_ske_nodepool_availability_zone gauge
stackit_ske_nodepool_availability_zone{cluster_name="ske-mock-01",nodepool_name="default-pool",project_id="",zone="eu01-01"} 1
stackit_ske_nodepool_availability_zone{cluster_name="ske-mock-01",nodepool_name="default-pool",project_id="",zone="eu01-02"} 1
stackit_ske_nodepool_availability_zone{cluster_name="ske-mock-01",nodepool_name="gpu-l40s",project_id="",zone="eu01-01"} 1
# HELP stackit_ske_nodepool_machine_type Type of machine used
# TYPE stackit_ske_nodepool_machine_type gauge
stackit_ske_nodepool_machine_type{cluster_name="ske-mock-01",machine_type="c1.4",nodepool_name="default-pool",project_id=""} 1
stackit_ske_nodepool_machine_type{cluster_name="ske-mock-01",machine_type="n2.14d.g1",nodepool_name="gpu-l40s",project_id=""} 1
# HELP stackit_ske_nodepool_machine_version_supported Machine image state: supported
# TYPE stackit_ske_nodepool_machine_version_supported gauge
stackit_ske_nodepool_machine_version_supported{cluster_name="ske-mock-01",image="flatcar",nodepool_name="default-pool",project_id="",version="4152.2.3"} 1
stackit_ske_nodepool_machine_version_supported{cluster_name="ske-mock-01",image="ubuntu",nodepool_name="gpu-l40s",project_id="",version="2204.20250516.0"} 1
# HELP stackit_ske_nodepool_volume_size_gb Volume size per node
# TYPE stackit_ske_nodepool_volume_size_gb gauge
stackit_ske_nodepool_volume_size_gb{cluster_name="ske-mock-01",nodepool_name="default-pool",project_id="",size_gb="50"} 50
stackit_ske_nodepool_volume_size_gb{cluster_name="ske-mock-01",nodepool_name="gpu-l40s",project_id="",size_gb="100"} 100
`

	err := testutil.GatherAndCompare(registry, strings.NewReader(expected),
		"stackit_ske_cluster_creation_timestamp",
		"stackit_ske_cluster_error_status",
		"stackit_ske_cluster_status_state_healthy",
		"stackit_ske_k8s_version_supported",
		"stackit_ske_cluster_maintenance_autoupdate_enabled",
		"stackit_ske_cluster_maintenance_end_timestamp",
		"stackit_ske_cluster_maintenance_start_timestamp",
		"stackit_ske_cluster_egress_address_range",
		"stackit_ske_nodepool_availability_zone",
		"stackit_ske_nodepool_machine_type",
		"stackit_ske_nodepool_machine_version_supported",
		"stackit_ske_nodepool_volume_size_gb",
	)

	require.NoError(t, err)
}
