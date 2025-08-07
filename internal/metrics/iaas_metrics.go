package metrics

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

var (
	sharedLabels = []string{"project_id", "server_id", "name", "zone", "machine_type"}
)

type IaasRegistry struct {
	ServerStatus      map[string]*prometheus.GaugeVec
	PowerStatus       map[string]*prometheus.GaugeVec
	MaintenanceStatus map[string]*prometheus.GaugeVec
	LastSeen          *prometheus.GaugeVec
	MaintenanceStart  *prometheus.GaugeVec
	MaintenanceEnd    *prometheus.GaugeVec
}

func NewIaasRegistry() *IaasRegistry {
	r := &IaasRegistry{
		ServerStatus:      make(map[string]*prometheus.GaugeVec),
		PowerStatus:       make(map[string]*prometheus.GaugeVec),
		MaintenanceStatus: make(map[string]*prometheus.GaugeVec),

		LastSeen: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "stackit_server_last_seen_timestamp",
			Help: "Unix time when the server was last seen by the exporter",
		}, sharedLabels),

		MaintenanceStart: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "stackit_server_maintenance_start_timestamp",
			Help: "Unix time when the maintenance window starts",
		}, sharedLabels),

		MaintenanceEnd: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "stackit_server_maintenance_end_timestamp",
			Help: "Unix time when the maintenance window ends",
		}, sharedLabels),
	}

	prometheus.MustRegister(
		r.LastSeen,
		r.MaintenanceStart,
		r.MaintenanceEnd,
	)

	// Server Lifecycle States
	for _, s := range []string{"ACTIVE", "INACTIVE", "CREATING", "DELETING", "REBUILDING", "ERROR"} {
		name := "stackit_server_status_" + normalize(s)
		vec := prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: name,
			Help: "Binary state of server status: 1 if " + s + ", else 0",
		}, sharedLabels)
		r.ServerStatus[s] = vec
		prometheus.MustRegister(vec)
	}

	// Server Power States
	for _, s := range []string{"RUNNING", "STOPPED", "CRASHED", "REBOOTING", "ERROR"} {
		name := "stackit_server_power_" + normalize(s)
		vec := prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: name,
			Help: "Binary state of power status: 1 if " + s + ", else 0",
		}, sharedLabels)
		r.PowerStatus[s] = vec
		prometheus.MustRegister(vec)
	}

	// Server Maintenance States
	for _, s := range []string{"PLANNED", "ONGOING"} {
		name := "stackit_server_maintenance_" + normalize(s)
		vec := prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: name,
			Help: "Binary state of maintenance status: 1 if " + s + ", else 0",
		}, sharedLabels)
		r.MaintenanceStatus[s] = vec
		prometheus.MustRegister(vec)
	}

	return r
}

// SetServerState updates all statuses for a server
func (r *IaasRegistry) SetServerState(
	status, powerStatus, maintenanceStatus string,
	labels prometheus.Labels,
	maintenanceStart, maintenanceEnd time.Time,
) {
	SetOneHotStatus(r.ServerStatus, status, labels)
	SetOneHotStatus(r.PowerStatus, powerStatus, labels)
	SetOneHotStatus(r.MaintenanceStatus, maintenanceStatus, labels)

	r.LastSeen.With(labels).SetToCurrentTime()

	if !maintenanceStart.IsZero() {
		r.MaintenanceStart.With(labels).Set(float64(maintenanceStart.Unix()))
	}
	if !maintenanceEnd.IsZero() {
		r.MaintenanceEnd.With(labels).Set(float64(maintenanceEnd.Unix()))
	}
}
