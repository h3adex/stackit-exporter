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
			status            = ""
			powerStatus       = ""
			maintenanceStatus = ""
			maintenanceEnd    time.Time
			maintenanceStart  time.Time
		)

		if srv.Status != nil {
			status = *srv.Status
		}

		if srv.PowerStatus != nil {
			powerStatus = *srv.PowerStatus
		}

		if mw := srv.MaintenanceWindow; mw != nil {
			maintenanceStatus = *mw.Status
			if mw.StartsAt != nil {
				maintenanceStart = mw.StartsAt.UTC()
			}
			if mw.EndsAt != nil {
				maintenanceEnd = mw.EndsAt.UTC()
			}
		}

		registry.SetServerState(status, powerStatus, maintenanceStatus, labels, maintenanceStart, maintenanceEnd)
	}
}
