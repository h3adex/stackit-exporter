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
	creationTime := time.Unix(1744629614, 0) // Mon May 13 2025 13:20:14 UTC

	start1 := time.Unix(1710000000, 0) // Example maintenance window
	end1 := time.Unix(1710003600, 0)

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

	require.NoError(t, registry.Register(reg.K8sVersion))
	require.NoError(t, registry.Register(reg.MaintenanceWindowStart))
	require.NoError(t, registry.Register(reg.MaintenanceWindowEnd))
	require.NoError(t, registry.Register(reg.ClusterStatus))
	require.NoError(t, registry.Register(reg.NodePoolMachineVersion))
	require.NoError(t, registry.Register(reg.NodePoolMachineTypes))
	require.NoError(t, registry.Register(reg.NodePoolVolumeSizes))
	require.NoError(t, registry.Register(reg.NodePoolAvailabilityZones))
	require.NoError(t, registry.Register(reg.MaintenanceAutoUpdate))
	require.NoError(t, registry.Register(reg.EgressAddressRanges))
	require.NoError(t, registry.Register(reg.ClusterCreationTimestamp))
	require.NoError(t, registry.Register(reg.ClusterErrorStatus))

	ctx := context.Background()
	projectID := "" // Empty as requested
	collector.ScrapeSkeAPI(ctx, client, projectID, "eu01", reg)

	const expected = `
# HELP stackit_ske_cluster_creation_timestamp Cluster creation time (Unix timestamp)
# TYPE stackit_ske_cluster_creation_timestamp gauge
stackit_ske_cluster_creation_timestamp{cluster_name="ske-mock-01",project_id=""} 1.744629614e+09
# HELP stackit_ske_cluster_error_status Indicates if a cluster has errors (1 if error exists, otherwise 0)
# TYPE stackit_ske_cluster_error_status gauge
stackit_ske_cluster_error_status{cluster_name="ske-mock-01",project_id=""} 0
# HELP stackit_ske_cluster_status Cluster status (1 if status is present). Use label 'status' to identify state such as STATE_HEALTHY.
# TYPE stackit_ske_cluster_status gauge
stackit_ske_cluster_status{cluster_name="ske-mock-01",project_id="",status="STATE_HEALTHY"} 1
# HELP stackit_ske_egress_address_ranges Egress CIDR address ranges of the cluster. Always 1; use labels.
# TYPE stackit_ske_egress_address_ranges gauge
stackit_ske_egress_address_ranges{cidr="10.0.0.1/32",cluster_name="ske-mock-01",project_id=""} 1
# HELP stackit_ske_k8s_version Kubernetes version in use (value always 1). ` + "`state`" + ` = supported/deprecated/preview
# TYPE stackit_ske_k8s_version gauge
stackit_ske_k8s_version{cluster_name="ske-mock-01",cluster_version="1.30.10",project_id="",state="supported"} 1
# HELP stackit_ske_maintenance_autoupdate_enabled Indicates if auto-update is enabled for maintenance
# TYPE stackit_ske_maintenance_autoupdate_enabled gauge
stackit_ske_maintenance_autoupdate_enabled{cluster_name="ske-mock-01",project_id=""} 0
# HELP stackit_ske_maintenance_window_end Scheduled maintenance window end time (Unix timestamp)
# TYPE stackit_ske_maintenance_window_end gauge
stackit_ske_maintenance_window_end{cluster_name="ske-mock-01",project_id=""} 1.7100036e+09
# HELP stackit_ske_maintenance_window_start Scheduled maintenance window start time (Unix timestamp)
# TYPE stackit_ske_maintenance_window_start gauge
stackit_ske_maintenance_window_start{cluster_name="ske-mock-01",project_id=""} 1.710000e+09
# HELP stackit_ske_nodepool_availability_zones Availability zones for node pools. Always 1; use labels.
# TYPE stackit_ske_nodepool_availability_zones gauge
stackit_ske_nodepool_availability_zones{cluster_name="ske-mock-01",nodepool_name="default-pool",project_id="",zone="eu01-01"} 1
stackit_ske_nodepool_availability_zones{cluster_name="ske-mock-01",nodepool_name="default-pool",project_id="",zone="eu01-02"} 1
stackit_ske_nodepool_availability_zones{cluster_name="ske-mock-01",nodepool_name="gpu-l40s",project_id="",zone="eu01-01"} 1
# HELP stackit_ske_nodepool_machine_types Machine types used in node pools. Always 1; use labels for details.
# TYPE stackit_ske_nodepool_machine_types gauge
stackit_ske_nodepool_machine_types{cluster_name="ske-mock-01",machine_type="c1.4",nodepool_name="default-pool",project_id=""} 1
stackit_ske_nodepool_machine_types{cluster_name="ske-mock-01",machine_type="n2.14d.g1",nodepool_name="gpu-l40s",project_id=""} 1
# HELP stackit_ske_nodepool_machine_version Machine image version in use (value always 1). ` + "`state`" + ` = supported/deprecated/preview
# TYPE stackit_ske_nodepool_machine_version gauge
stackit_ske_nodepool_machine_version{cluster_name="ske-mock-01",nodepool_name="default-pool",os_name="flatcar",os_version="4152.2.3",project_id="",state="supported"} 1
stackit_ske_nodepool_machine_version{cluster_name="ske-mock-01",nodepool_name="gpu-l40s",os_name="ubuntu",os_version="2204.20250516.0",project_id="",state="supported"} 1
# HELP stackit_ske_nodepool_volume_sizes_gb Volume sizes in the node pools (in GB)
# TYPE stackit_ske_nodepool_volume_sizes_gb gauge
stackit_ske_nodepool_volume_sizes_gb{cluster_name="ske-mock-01",nodepool_name="default-pool",project_id="",volume_size="50"} 50
stackit_ske_nodepool_volume_sizes_gb{cluster_name="ske-mock-01",nodepool_name="gpu-l40s",project_id="",volume_size="100"} 100
`

	err := testutil.GatherAndCompare(registry, strings.NewReader(expected),
		"stackit_ske_cluster_creation_timestamp",
		"stackit_ske_cluster_error_status",
		"stackit_ske_cluster_status",
		"stackit_ske_egress_address_ranges",
		"stackit_ske_k8s_version",
		"stackit_ske_maintenance_window_start",
		"stackit_ske_maintenance_window_end",
		"stackit_ske_maintenance_autoupdate_enabled",
		"stackit_ske_nodepool_availability_zones",
		"stackit_ske_nodepool_machine_types",
		"stackit_ske_nodepool_machine_version",
		"stackit_ske_nodepool_volume_sizes_gb",
	)

	require.NoError(t, err)
}
