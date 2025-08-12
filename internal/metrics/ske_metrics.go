package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
)

var sharedClusterLabels = []string{"project_id", "cluster_name"}
var sharedNodePoolLabels = []string{"project_id", "cluster_name", "nodepool_name"}

type SKERegistry struct {
	ClusterInfo *prometheus.GaugeVec

	ClusterStatus       *prometheus.GaugeVec
	ClusterErrorStatus  *prometheus.GaugeVec
	ClusterCreationTime *prometheus.GaugeVec
	ClusterLastSeen     *prometheus.GaugeVec

	K8sVersion *prometheus.GaugeVec

	MaintenanceAutoUpdate  *prometheus.GaugeVec
	MaintenanceWindowStart *prometheus.GaugeVec
	MaintenanceWindowEnd   *prometheus.GaugeVec

	EgressAddressRanges *prometheus.GaugeVec

	NodePoolMachineVersion    *prometheus.GaugeVec
	NodePoolMachineTypes      *prometheus.GaugeVec
	NodePoolVolumeSizes       *prometheus.GaugeVec
	NodePoolAvailabilityZones *prometheus.GaugeVec
	NodePoolLastSeen          *prometheus.GaugeVec
}

func NewSKERegistry() *SKERegistry {
	reg := &SKERegistry{
		ClusterInfo: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "stackit_ske_info",
				Help: "General information about the Stackit SKE cluster. Set to 1 if the cluster exists, 0 otherwise.",
			},
			[]string{
				"creation_time", "credentials_rotation_last_completion_time", "credentials_rotation_last_initiation_time",
				"credentials_rotation_phase", "egress_address_ranges", "hibernated", "kubernetes_version",
				"maintenance_machine_image_enabled", "maintenance_machine_kubernetes_enabled", "maintenance_window_end",
				"maintenance_window_start", "name", "network_id", "nodepool_length", "observability_enabled",
				"observability_instance_id", "pod_address_ranges", "status",
			},
		),
		ClusterStatus: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "stackit_ske_cluster_status",
			Help: "Cluster status: ",
		}, append(sharedClusterLabels, "status")),
		ClusterErrorStatus: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "stackit_ske_cluster_error_status",
			Help: "1 if cluster has error",
		}, sharedClusterLabels),
		ClusterCreationTime: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "stackit_ske_cluster_creation_timestamp",
			Help: "Unix timestamp when cluster was created",
		}, sharedClusterLabels),
		ClusterLastSeen: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "stackit_ske_cluster_last_seen_timestamp",
			Help: "Last observed timestamp",
		}, sharedClusterLabels),
		K8sVersion: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "stackit_ske_k8s_version",
			Help: "Kubernetes version state",
		}, append(sharedClusterLabels, "k8s_version", "status")),
		MaintenanceAutoUpdate: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "stackit_ske_cluster_maintenance_autoupdate_enabled",
			Help: "1 if autoupdate is enabled",
		}, sharedClusterLabels),
		MaintenanceWindowStart: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "stackit_ske_cluster_maintenance_start_timestamp",
			Help: "Start time of maintenance window",
		}, sharedClusterLabels),
		MaintenanceWindowEnd: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "stackit_ske_cluster_maintenance_end_timestamp",
			Help: "End time of maintenance window",
		}, sharedClusterLabels),
		EgressAddressRanges: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "stackit_ske_cluster_egress_address_range",
			Help: "Egress IP range used by cluster",
		}, append(sharedClusterLabels, "cidr")),
		NodePoolMachineVersion: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "stackit_ske_nodepool_machine_version",
			Help: "Machine image state: ",
		}, append(sharedNodePoolLabels, "image", "version", "status")),
		NodePoolMachineTypes: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "stackit_ske_nodepool_machine_type",
			Help: "Type of machine used",
		}, append(sharedNodePoolLabels, "machine_type")),
		NodePoolVolumeSizes: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "stackit_ske_nodepool_volume_size_gb",
			Help: "Volume size per node",
		}, append(sharedNodePoolLabels, "size_gb")),
		NodePoolAvailabilityZones: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "stackit_ske_nodepool_availability_zone",
			Help: "Availability zones configured",
		}, append(sharedNodePoolLabels, "zone")),
		NodePoolLastSeen: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "stackit_ske_nodepool_last_seen_timestamp",
			Help: "Last time node pool was seen",
		}, sharedNodePoolLabels),
	}

	prometheus.MustRegister(
		reg.ClusterInfo,
		reg.ClusterStatus,
		reg.ClusterErrorStatus,
		reg.ClusterCreationTime,
		reg.ClusterLastSeen,
		reg.K8sVersion,
		reg.MaintenanceAutoUpdate,
		reg.MaintenanceWindowStart,
		reg.MaintenanceWindowEnd,
		reg.EgressAddressRanges,
		reg.NodePoolMachineVersion,
		reg.NodePoolMachineTypes,
		reg.NodePoolVolumeSizes,
		reg.NodePoolAvailabilityZones,
		reg.NodePoolLastSeen,
	)

	return reg
}
