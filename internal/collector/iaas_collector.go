package collector

import (
	"context"
	"log"
	"time"

	"github.com/h3adex/stackit-exporter/internal/metrics"
	"github.com/stackitcloud/stackit-sdk-go/services/iaas"
)

type IaasClient interface {
	ListServers(ctx context.Context, projectID string) iaas.ApiListServersRequest
}

func ScrapeIaasAPI(ctx context.Context, client IaasClient, projectID string, registry *metrics.IaasRegistry) {
	resp, err := client.ListServers(ctx, projectID).Execute()
	if err != nil {
		log.Printf("failed to fetch IaaS servers: %v", err)
		return
	}
	if resp == nil || resp.Items == nil {
		log.Println("received nil or empty servers response")
		return
	}

	for i := range *resp.Items {
		srv := &(*resp.Items)[i]

		if srv.Id == nil || srv.Name == nil || srv.AvailabilityZone == nil || srv.MachineType == nil {
			continue
		}

		labels := map[string]string{
			"project_id":   projectID,
			"server_id":    *srv.Id,
			"name":         *srv.Name,
			"zone":         *srv.AvailabilityZone,
			"machine_type": *srv.MachineType,
		}

		serverInfoLabels := make(map[string]string, len(labels))
		for k, v := range labels {
			serverInfoLabels[k] = v
		}

		// Basic safe values for status fields
		status := SafeString(srv.Status)
		powerStatus := SafeString(srv.PowerStatus)
		maintenanceStatus := ""
		maintenanceDetails := ""
		maintenanceStart := 0.0
		maintenanceEnd := 0.0

		serverInfoLabels["power_status"] = powerStatus
		serverInfoLabels["server_status"] = status
		serverInfoLabels["maintenance_status"] = ""
		serverInfoLabels["image_id"] = SafeString(srv.ImageId)
		serverInfoLabels["keypair_name"] = SafeString(srv.KeypairName)
		serverInfoLabels["boot_volume_id"] = ""
		serverInfoLabels["affinity_group"] = SafeString(srv.AffinityGroup)
		serverInfoLabels["created_at"] = SafeTime(srv.CreatedAt)
		serverInfoLabels["launched_at"] = SafeTime(srv.LaunchedAt)

		if srv.BootVolume != nil && srv.BootVolume.Id != nil {
			serverInfoLabels["boot_volume_id"] = *srv.BootVolume.Id
		}

		// Maintenance info
		if mw := srv.MaintenanceWindow; mw != nil {
			maintenanceStatus = SafeString(mw.Status)
			serverInfoLabels["maintenance_status"] = maintenanceStatus

			if mw.Details != nil {
				maintenanceDetails = *mw.Details
			}
			maintenanceStart = SafeTimeUnix(mw.StartsAt)
			maintenanceEnd = SafeTimeUnix(mw.EndsAt)
		}

		serverInfoLabels["maintenance"] = maintenanceStatus
		serverInfoLabels["maintenance_details"] = maintenanceDetails

		// Set static metadata
		registry.ServerInfo.With(serverInfoLabels).Set(1)

		// Time metrics
		registry.LastSeen.With(labels).Set(float64(time.Now().Unix()))
		if maintenanceStart > 0 {
			registry.MaintenanceStart.With(labels).Set(maintenanceStart)
		}
		if maintenanceEnd > 0 {
			registry.MaintenanceEnd.With(labels).Set(maintenanceEnd)
		}

		// Set binary status values
		registry.SetServerState(status, powerStatus, maintenanceStatus, labels)
	}
}
