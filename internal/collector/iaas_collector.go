package collector

import (
	"context"
	"log"

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

		serverID := *srv.Id
		name := *srv.Name
		zone := *srv.AvailabilityZone
		machineType := *srv.MachineType

		var unixStart, unixEnd float64

		if mw := srv.MaintenanceWindow; mw != nil {
			if mw.StartsAt != nil {
				unixStart = float64(mw.StartsAt.UTC().Unix())
			}
			if mw.EndsAt != nil {
				unixEnd = float64(mw.EndsAt.UTC().Unix())
			}

			baseLabels := []string{projectID, serverID, name, zone, machineType}
			registry.MaintenanceStart.WithLabelValues(baseLabels...).Set(unixStart)
			registry.MaintenanceEnd.WithLabelValues(baseLabels...).Set(unixEnd)

			if mw.Status != nil {
				registry.MaintenanceStatus.WithLabelValues(
					projectID, serverID, name, zone, machineType, *mw.Status,
				).Set(1)
			}
		}

		if srv.Status != nil {
			registry.ServerStatus.WithLabelValues(
				projectID, serverID, name, zone, *srv.Status,
			).Set(1)
		}

		if srv.PowerStatus != nil {
			registry.ServerPowerStatus.WithLabelValues(
				projectID, serverID, name, zone, *srv.PowerStatus,
			).Set(1)
		}
	}
}
