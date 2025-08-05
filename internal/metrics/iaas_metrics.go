package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
)

type IaasRegistry struct {
	MaintenanceStart  *prometheus.GaugeVec
	MaintenanceEnd    *prometheus.GaugeVec
	MaintenanceStatus *prometheus.GaugeVec
	ServerStatus      *prometheus.GaugeVec
	ServerPowerStatus *prometheus.GaugeVec
}

func NewIaasRegistry() *IaasRegistry {
	r := &IaasRegistry{
		MaintenanceStart: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "stackit_server_maintenance_start_timestamp",
			Help: "Scheduled maintenance window start time (Unix timestamp)",
		}, []string{"project_id", "server_id", "name", "zone", "machine_type"}),

		MaintenanceEnd: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "stackit_server_maintenance_end_timestamp",
			Help: "Scheduled maintenance window end time (Unix timestamp)",
		}, []string{"project_id", "server_id", "name", "zone", "machine_type"}),

		MaintenanceStatus: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "stackit_server_maintenance_status",
			Help: "Current maintenance status (label: status; always 1)",
		}, []string{"project_id", "server_id", "name", "zone", "machine_type", "status"}),

		ServerStatus: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "stackit_server_status",
			Help: "Current server status (label: status; always 1)",
		}, []string{"project_id", "server_id", "name", "zone", "status"}),

		ServerPowerStatus: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "stackit_server_power_status",
			Help: "Current server power status (label: power_status; always 1)",
		}, []string{"project_id", "server_id", "name", "zone", "power_status"}),
	}

	prometheus.MustRegister(
		r.MaintenanceStart,
		r.MaintenanceEnd,
		r.MaintenanceStatus,
		r.ServerStatus,
		r.ServerPowerStatus,
	)

	return r
}

func (r *IaasRegistry) Describe(ch chan<- *prometheus.Desc) {
	r.MaintenanceStart.Describe(ch)
	r.MaintenanceEnd.Describe(ch)
	r.MaintenanceStatus.Describe(ch)
	r.ServerStatus.Describe(ch)
	r.ServerPowerStatus.Describe(ch)
}

func (r *IaasRegistry) Collect(ch chan<- prometheus.Metric) {
	r.MaintenanceStart.Collect(ch)
	r.MaintenanceEnd.Collect(ch)
	r.MaintenanceStatus.Collect(ch)
	r.ServerStatus.Collect(ch)
	r.ServerPowerStatus.Collect(ch)
}
