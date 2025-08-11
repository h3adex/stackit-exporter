package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
)

var sharedIaasLabels = []string{"project_id", "server_id", "name", "zone", "machine_type"}

type IaasRegistry struct {
	ServerInfo        *prometheus.GaugeVec
	ServerStatus      map[string]*prometheus.GaugeVec
	PowerStatus       map[string]*prometheus.GaugeVec
	MaintenanceStatus map[string]*prometheus.GaugeVec
	LastSeen          *prometheus.GaugeVec
	MaintenanceStart  *prometheus.GaugeVec
	MaintenanceEnd    *prometheus.GaugeVec
}

func NewIaasRegistry() *IaasRegistry {
	r := &IaasRegistry{
		ServerInfo: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "stackit_server_info",
			Help: "Descriptive info about the server at export time. Set to 1 if present.",
		}, []string{
			"project_id", "server_id", "name", "zone", "machine_type",
			"power_status", "server_status", "maintenance_status",
			"image_id", "keypair_name", "boot_volume_id", "affinity_group",
			"maintenance", "maintenance_details", "created_at", "launched_at",
		}),
		ServerStatus:      make(map[string]*prometheus.GaugeVec),
		PowerStatus:       make(map[string]*prometheus.GaugeVec),
		MaintenanceStatus: make(map[string]*prometheus.GaugeVec),
		LastSeen: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "stackit_server_last_seen_timestamp",
			Help: "Unix timestamp when the server was last scraped",
		}, sharedIaasLabels),
		MaintenanceStart: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "stackit_server_maintenance_start_timestamp",
			Help: "Unix timestamp of maintenance start time",
		}, sharedIaasLabels),
		MaintenanceEnd: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "stackit_server_maintenance_end_timestamp",
			Help: "Unix timestamp of maintenance end time",
		}, sharedIaasLabels),
	}

	// Register default metrics
	prometheus.MustRegister(
		r.ServerInfo,
		r.LastSeen,
		r.MaintenanceStart,
		r.MaintenanceEnd,
	)

	// Server Lifecycle States
	for _, s := range []string{"ACTIVE", "INACTIVE", "CREATING", "DELETING", "REBUILDING", "ERROR"} {
		name := "stackit_server_status_" + normalize(s)
		vec := prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: name,
			Help: "Binary server status metric. 1 if current state is " + s,
		}, sharedIaasLabels)
		r.ServerStatus[s] = vec
		prometheus.MustRegister(vec)
	}

	// Power States
	for _, s := range []string{"RUNNING", "STOPPED", "CRASHED", "REBOOTING", "ERROR"} {
		name := "stackit_server_power_" + normalize(s)
		vec := prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: name,
			Help: "Binary power status metric. 1 if current power state is " + s,
		}, sharedIaasLabels)
		r.PowerStatus[s] = vec
		prometheus.MustRegister(vec)
	}

	// Maintenance States
	for _, s := range []string{"PLANNED", "ONGOING"} {
		name := "stackit_server_maintenance_" + normalize(s)
		vec := prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: name,
			Help: "Binary maintenance status. 1 if current state is " + s,
		}, sharedIaasLabels)
		r.MaintenanceStatus[s] = vec
		prometheus.MustRegister(vec)
	}

	return r
}

// SetServerState sets the binary one-hot server status, power status, and maintenance status
func (r *IaasRegistry) SetServerState(status, powerStatus, maintenanceStatus string, labels prometheus.Labels) {
	SetOneHotStatus(r.ServerStatus, status, labels)
	SetOneHotStatus(r.PowerStatus, powerStatus, labels)
	SetOneHotStatus(r.MaintenanceStatus, maintenanceStatus, labels)
}
