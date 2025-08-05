package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
)

type SKERegistry struct {
	K8sVersion                *prometheus.GaugeVec
	MaintenanceAutoUpdate     *prometheus.GaugeVec
	MaintenanceWindowStart    *prometheus.GaugeVec
	MaintenanceWindowEnd      *prometheus.GaugeVec
	ClusterStatus             *prometheus.GaugeVec
	ClusterCreationTimestamp  *prometheus.GaugeVec
	EgressAddressRanges       *prometheus.GaugeVec
	NodePoolMachineTypes      *prometheus.GaugeVec
	NodePoolMachineVersion    *prometheus.GaugeVec
	NodePoolVolumeSizes       *prometheus.GaugeVec
	NodePoolAvailabilityZones *prometheus.GaugeVec
	ClusterErrorStatus        *prometheus.GaugeVec
}

func NewSKERegistry() *SKERegistry {
	r := &SKERegistry{
		K8sVersion: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "stackit_ske_k8s_version",
			Help: "Kubernetes version in use (value always 1). `state` = supported/deprecated/preview",
		}, []string{"project_id", "cluster_name", "cluster_version", "state"}),

		MaintenanceAutoUpdate: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "stackit_ske_maintenance_autoupdate_enabled",
			Help: "Indicates if auto-update is enabled for maintenance",
		}, []string{"project_id", "cluster_name"}),

		MaintenanceWindowStart: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "stackit_ske_maintenance_window_start",
			Help: "Scheduled maintenance window start time (Unix timestamp)",
		}, []string{"project_id", "cluster_name"}),

		MaintenanceWindowEnd: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "stackit_ske_maintenance_window_end",
			Help: "Scheduled maintenance window end time (Unix timestamp)",
		}, []string{"project_id", "cluster_name"}),

		ClusterStatus: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "stackit_ske_cluster_status",
			Help: "Cluster status (1 if status is present). Use label 'status' to identify state such as STATE_HEALTHY.",
		}, []string{"project_id", "cluster_name", "status"}),

		ClusterCreationTimestamp: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "stackit_ske_cluster_creation_timestamp",
			Help: "Cluster creation time (Unix timestamp)",
		}, []string{"project_id", "cluster_name"}),

		NodePoolMachineTypes: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "stackit_ske_nodepool_machine_types",
			Help: "Machine types used in node pools. Always 1; use labels for details.",
		}, []string{"project_id", "cluster_name", "nodepool_name", "machine_type"}),

		NodePoolMachineVersion: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "stackit_ske_nodepool_machine_version",
			Help: "Machine image version in use (value always 1). `state` = supported/deprecated/preview",
		}, []string{"project_id", "cluster_name", "nodepool_name", "os_name", "os_version", "state"}),

		NodePoolVolumeSizes: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "stackit_ske_nodepool_volume_sizes_gb",
			Help: "Volume sizes in the node pools (in GB)",
		}, []string{"project_id", "cluster_name", "nodepool_name", "volume_size"}),

		NodePoolAvailabilityZones: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "stackit_ske_nodepool_availability_zones",
			Help: "Availability zones for node pools. Always 1; use labels.",
		}, []string{"project_id", "cluster_name", "nodepool_name", "zone"}),

		EgressAddressRanges: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "stackit_ske_egress_address_ranges",
			Help: "Egress CIDR address ranges of the cluster. Always 1; use labels.",
		}, []string{"project_id", "cluster_name", "cidr"}),

		ClusterErrorStatus: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "stackit_ske_cluster_error_status",
			Help: "Indicates if a cluster has errors (1 if error exists, otherwise 0)",
		}, []string{"project_id", "cluster_name"}),
	}

	prometheus.MustRegister(
		r.K8sVersion,
		r.MaintenanceWindowStart,
		r.MaintenanceWindowEnd,
		r.ClusterStatus,
		r.NodePoolMachineVersion,
		r.NodePoolMachineTypes,
		r.NodePoolVolumeSizes,
		r.NodePoolAvailabilityZones,
		r.MaintenanceAutoUpdate,
		r.EgressAddressRanges,
		r.ClusterCreationTimestamp,
		r.ClusterErrorStatus,
	)

	return r
}

func (r *SKERegistry) Describe(ch chan<- *prometheus.Desc) {
	r.K8sVersion.Describe(ch)
	r.MaintenanceAutoUpdate.Describe(ch)
	r.MaintenanceWindowStart.Describe(ch)
	r.MaintenanceWindowEnd.Describe(ch)
	r.ClusterStatus.Describe(ch)
	r.ClusterCreationTimestamp.Describe(ch)
	r.EgressAddressRanges.Describe(ch)
	r.NodePoolMachineVersion.Describe(ch)
	r.NodePoolMachineTypes.Describe(ch)
	r.NodePoolVolumeSizes.Describe(ch)
	r.NodePoolAvailabilityZones.Describe(ch)
	r.ClusterErrorStatus.Describe(ch)
}

func (r *SKERegistry) Collect(ch chan<- prometheus.Metric) {
	r.K8sVersion.Collect(ch)
	r.MaintenanceAutoUpdate.Collect(ch)
	r.MaintenanceWindowStart.Collect(ch)
	r.MaintenanceWindowEnd.Collect(ch)
	r.ClusterStatus.Collect(ch)
	r.ClusterCreationTimestamp.Collect(ch)
	r.EgressAddressRanges.Collect(ch)
	r.NodePoolMachineVersion.Collect(ch)
	r.NodePoolMachineTypes.Collect(ch)
	r.NodePoolVolumeSizes.Collect(ch)
	r.NodePoolAvailabilityZones.Collect(ch)
	r.ClusterErrorStatus.Collect(ch)
}
