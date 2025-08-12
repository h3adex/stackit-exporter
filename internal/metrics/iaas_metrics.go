package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
)

const (
	LabelAffinityGroup     = "affinity_group"
	LabelCreatedAt         = "created_at"
	LabelImageID           = "image_id"
	LabelKeypairName       = "keypair_name"
	LabelLaunchedAt        = "launched_at"
	LabelMachineType       = "machine_type"
	LabelMaintenanceInfo   = "maintenance_details"
	LabelMaintenanceStatus = "maintenance_status"
	LabelName              = "name"
	LabelPowerStatus       = "power_status"
	LabelProjectID         = "project_id"
	LabelServerID          = "server_id"
	LabelServerStatus      = "server_status"
	LabelStatus            = "status"
	LabelZone              = "zone"
)

// sharedIaasLabels are the base identifying labels for most server-related metrics.
var sharedIaasLabels = []string{
	LabelProjectID,
	LabelServerID,
	LabelName,
	LabelZone,
	LabelMachineType,
}

// infoLabels are all static/info labels exposed in stackit_server_info.
var infoLabels = []string{
	LabelProjectID,
	LabelServerID,
	LabelName,
	LabelZone,
	LabelMachineType,
	LabelPowerStatus,
	LabelServerStatus,
	LabelMaintenanceStatus,
	LabelImageID,
	LabelKeypairName,
	LabelAffinityGroup,
	LabelCreatedAt,
	LabelLaunchedAt,
	LabelMaintenanceInfo,
}

// IaasRegistry holds all Prometheus metrics related to IaaS servers.
type IaasRegistry struct {
	ServerInfo        *prometheus.GaugeVec
	ServerStatus      *prometheus.GaugeVec
	ServerPowerStatus *prometheus.GaugeVec
	MaintenanceStatus *prometheus.GaugeVec

	ServerLastSeen   *prometheus.GaugeVec
	MaintenanceStart *prometheus.GaugeVec
	MaintenanceEnd   *prometheus.GaugeVec
}

// NewIaasRegistry creates and registers all server metrics.
func NewIaasRegistry() *IaasRegistry {
	r := &IaasRegistry{
		ServerInfo: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "stackit_server_info",
			Help: "Descriptive labels of the server at scrape time. Value is always 1.",
		}, infoLabels),

		ServerStatus: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "stackit_server_status",
			Help: "Current status of the server. Possible status can be found here: https://docs.api.eu01.stackit.cloud/documentation/iaas/version/v1#tag/Servers/operation/v1ListServersInProject. Value is always 1.",
		}, append(sharedIaasLabels, LabelStatus)),

		ServerPowerStatus: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "stackit_server_power_status",
			Help: "Current power state of the server. Possible status can be found here: https://docs.api.eu01.stackit.cloud/documentation/iaas/version/v1#tag/Servers/operation/v1ListServersInProject. Value is always 1.",
		}, append(sharedIaasLabels, LabelStatus)),

		MaintenanceStatus: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "stackit_server_maintenance_status",
			Help: "Maintenance state of the server. Possible status can be found here: https://docs.api.eu01.stackit.cloud/documentation/iaas/version/v1#tag/Servers/operation/v1ListServersInProject. Value is always 1.",
		}, append(sharedIaasLabels, LabelStatus)),

		ServerLastSeen: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "stackit_server_last_seen_timestamp",
			Help: "Unix timestamp when the server was last seen during a scrape.",
		}, sharedIaasLabels),

		MaintenanceStart: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "stackit_server_maintenance_start_timestamp",
			Help: "Unix timestamp for the start of a scheduled maintenance window.",
		}, sharedIaasLabels),

		MaintenanceEnd: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "stackit_server_maintenance_end_timestamp",
			Help: "Unix timestamp for the end of a scheduled maintenance window.",
		}, sharedIaasLabels),
	}

	// Register all metrics with Prometheus's default registry.
	prometheus.MustRegister(
		r.ServerInfo,
		r.ServerStatus,
		r.ServerPowerStatus,
		r.MaintenanceStatus,
		r.ServerLastSeen,
		r.MaintenanceStart,
		r.MaintenanceEnd,
	)

	return r
}
