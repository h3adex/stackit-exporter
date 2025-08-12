package collector

import (
	"context"
	"log"
	"time"

	"github.com/h3adex/stackit-exporter/internal/metrics"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/stackitcloud/stackit-sdk-go/services/iaas"
)

// IaasClient defines the ability to list IaaS servers for a project.
type IaasClient interface {
	ListServers(ctx context.Context, projectID string) iaas.ApiListServersRequest
}

// ScrapeIaasAPI collects VM metrics from a project and exports them via Prometheus.
func ScrapeIaasAPI(ctx context.Context, client IaasClient, projectID string, registry *metrics.IaasRegistry) {
	resp, err := client.ListServers(ctx, projectID).Execute()
	if err != nil || resp == nil || resp.Items == nil {
		log.Printf("failed to fetch IaaS servers for project %s: %v", projectID, err)
		return
	}

	for _, srv := range *resp.Items {
		if srv.Id == nil || srv.Name == nil || srv.AvailabilityZone == nil || srv.MachineType == nil {
			log.Printf("skipping server with missing required metadata: %+v", srv)
			continue
		}

		basicLabel := basicLabels(projectID, srv)
		serverInfoLabels := enrichLabels(basicLabel, srv)

		registry.ServerInfo.With(serverInfoLabels).Set(1)
		setStatusMetric(registry.ServerPowerStatus, basicLabel, serverInfoLabels[metrics.LabelPowerStatus])
		setStatusMetric(registry.ServerStatus, basicLabel, serverInfoLabels[metrics.LabelServerStatus])
		setStatusMetric(registry.MaintenanceStatus, basicLabel, serverInfoLabels[metrics.LabelMaintenanceStatus])

		// Last seen timestamp.
		registry.ServerLastSeen.With(basicLabel).Set(float64(time.Now().Unix()))

		// Maintenance window metrics.
		if mw := srv.MaintenanceWindow; mw != nil {
			if start := SafeTimeUnix(mw.StartsAt); start > 0 {
				registry.MaintenanceStart.With(basicLabel).Set(start)
			}
			if end := SafeTimeUnix(mw.EndsAt); end > 0 {
				registry.MaintenanceEnd.With(basicLabel).Set(end)
			}
		}
	}
}

// basicLabels returns the minimal label set for a VM.
func basicLabels(projectID string, s iaas.Server) map[string]string {
	return map[string]string{
		metrics.LabelProjectID:   projectID,
		metrics.LabelServerID:    *s.Id,
		metrics.LabelName:        *s.Name,
		metrics.LabelZone:        *s.AvailabilityZone,
		metrics.LabelMachineType: *s.MachineType,
	}
}

// enrichLabels adds extended static fields as labels for server_info.
func enrichLabels(base map[string]string, s iaas.Server) map[string]string {
	labels := CopyMap(base)

	labels[metrics.LabelPowerStatus] = SafeString(s.PowerStatus)
	labels[metrics.LabelServerStatus] = SafeString(s.Status)
	labels[metrics.LabelImageID] = SafeString(s.ImageId)
	labels[metrics.LabelKeypairName] = SafeString(s.KeypairName)
	labels[metrics.LabelAffinityGroup] = SafeString(s.AffinityGroup)
	labels[metrics.LabelCreatedAt] = SafeTime(s.CreatedAt)
	labels[metrics.LabelLaunchedAt] = SafeTime(s.LaunchedAt)

	// Maintenance
	if mw := s.MaintenanceWindow; mw != nil {
		labels[metrics.LabelMaintenanceStatus] = SafeString(mw.Status)
		labels[metrics.LabelMaintenanceInfo] = SafeString(mw.Details)
	} else {
		labels[metrics.LabelMaintenanceStatus] = ""
		labels[metrics.LabelMaintenanceInfo] = ""
	}

	return labels
}

// setStatusMetric sets a label-based state metric.
func setStatusMetric(metric *prometheus.GaugeVec, base map[string]string, state string) {
	labels := CopyMap(base)
	labels[metrics.LabelStatus] = state
	metric.With(labels).Set(1)
}
