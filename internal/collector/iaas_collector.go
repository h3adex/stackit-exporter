package collector

import (
	"context"
	"log"
	"time"

	"github.com/h3adex/stackit-exporter/internal/metrics"
	"github.com/h3adex/stackit-exporter/internal/utils"
	"github.com/stackitcloud/stackit-sdk-go/services/iaas"
)

type IaasClient interface {
	ListServers(ctx context.Context, projectID string) iaas.ApiListServersRequest
}

func ScrapeIaasAPI(ctx context.Context, client IaasClient, projectID string, registry *metrics.IaasRegistry) {
	resp, err := client.ListServers(ctx, projectID).Execute()
	if err != nil {
		log.Printf("failed to fetch servers: %v", err)
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

		var (
			status             = ""
			powerStatus        = ""
			maintenanceStatus  = ""
			maintenanceDetails = ""
			maintenanceEnd     time.Time
			maintenanceStart   time.Time
		)

		if srv.Status != nil {
			status = *srv.Status
		}

		if srv.PowerStatus != nil {
			powerStatus = *srv.PowerStatus
		}

		if mw := srv.MaintenanceWindow; mw != nil {
			maintenanceStatus = *mw.Status
			if mw.Details != nil {
				maintenanceDetails = *mw.Details
			}
			if mw.StartsAt != nil {
				maintenanceStart = mw.StartsAt.UTC()
			}
			if mw.EndsAt != nil {
				maintenanceEnd = mw.EndsAt.UTC()
			}
		}

		serverInfoLabels := make(map[string]string, len(labels))
		for k, v := range labels {
			serverInfoLabels[k] = v
		}

		serverInfoLabels["image_id"] = utils.SafeString(srv.ImageId)
		serverInfoLabels["keypair_name"] = utils.SafeString(srv.KeypairName)
		serverInfoLabels["boot_volume_id"] = ""
		if srv.BootVolume != nil && srv.BootVolume.Id != nil {
			serverInfoLabels["boot_volume_id"] = *srv.BootVolume.Id
		}
		serverInfoLabels["affinity_group"] = utils.SafeString(srv.AffinityGroup)
		serverInfoLabels["maintenance"] = maintenanceStatus
		serverInfoLabels["maintenance_details"] = maintenanceDetails
		serverInfoLabels["created_at"] = utils.SafeTime(srv.CreatedAt)
		serverInfoLabels["launched_at"] = utils.SafeTime(srv.LaunchedAt)

		registry.ServerInfo.With(serverInfoLabels).Set(1)
		registry.LastSeen.With(labels).SetToCurrentTime()
		if !maintenanceStart.IsZero() {
			registry.MaintenanceStart.With(labels).Set(float64(maintenanceStart.Unix()))
		}
		if !maintenanceEnd.IsZero() {
			registry.MaintenanceEnd.With(labels).Set(float64(maintenanceEnd.Unix()))
		}

		registry.SetServerState(status, powerStatus, maintenanceStatus, labels)
	}
}
