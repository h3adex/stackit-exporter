package collector

import (
	"context"
	"log"
	"strconv"

	"github.com/h3adex/stackit-exporter/internal/metrics"
	"github.com/h3adex/stackit-exporter/internal/utils"
	"github.com/stackitcloud/stackit-sdk-go/services/ske"
)

type SkeClient interface {
	ListClusters(ctx context.Context, projectID string, region string) ske.ApiListClustersRequest
	ListProviderOptions(ctx context.Context, region string) ske.ApiListProviderOptionsRequest
}

func ScrapeSkeAPI(ctx context.Context, client SkeClient, projectID, region string, registry *metrics.SKERegistry) {
	resp, err := client.ListClusters(ctx, projectID, region).Execute()
	if err != nil {
		log.Printf("failed to fetch Kubernetes clusters: %v", err)
		return
	}

	clusterOptions, err := client.ListProviderOptions(ctx, region).Execute()
	if err != nil {
		log.Printf("failed to fetch Kubernetes clusters option: %v", err)
		return
	}

	for _, cluster := range *resp.Items {
		clusterVersionState := ""
		for _, option := range *clusterOptions.KubernetesVersions {
			if *cluster.Kubernetes.Version == *option.Version {
				clusterVersionState = *option.State
			}
		}
		clusterName := cluster.Name

		// Set Kubernetes version metric
		registry.K8sVersion.WithLabelValues(projectID, *clusterName, *cluster.Kubernetes.Version, clusterVersionState).Set(1)

		// Set Maintenance auto-update status
		registry.MaintenanceAutoUpdate.WithLabelValues(projectID, *clusterName).Set(utils.BoolToFloat(*cluster.Maintenance.AutoUpdate.KubernetesVersion))

		// Set Maintenance window metrics
		registry.MaintenanceWindowStart.WithLabelValues(projectID, *clusterName).Set(float64(cluster.Maintenance.TimeWindow.Start.UTC().Unix()))
		registry.MaintenanceWindowEnd.WithLabelValues(projectID, *clusterName).Set(float64(cluster.Maintenance.TimeWindow.End.UTC().Unix()))

		// Set Cluster status metrics
		registry.ClusterStatus.WithLabelValues(projectID, *clusterName, string(*cluster.Status.Aggregated)).Set(statusToGaugeValue(*cluster.Status.Aggregated))
		registry.ClusterCreationTimestamp.WithLabelValues(projectID, *clusterName).Set(float64(cluster.Status.CreationTime.UTC().Unix()))

		// Node pool metrics
		for _, nodepool := range *cluster.Nodepools {
			// parse the current state of the machine version
			machineVersionState := ""
			machineOsName := ""
			machineOsVersion := ""
			for _, image := range *clusterOptions.MachineImages {
				for _, version := range *image.Versions {
					if *nodepool.Machine.Image.Version == *version.Version {
						machineVersionState = *version.State
						machineOsName = *image.Name
						machineOsVersion = *version.Version
					}
				}
			}

			registry.NodePoolMachineVersion.WithLabelValues(projectID, *clusterName, *nodepool.Name, machineOsName, machineOsVersion, machineVersionState).Set(1)
			registry.NodePoolMachineTypes.WithLabelValues(projectID, *clusterName, *nodepool.Name, *nodepool.Machine.Type).Set(1)
			registry.NodePoolVolumeSizes.WithLabelValues(projectID, *clusterName, *nodepool.Name, strconv.Itoa(int(*nodepool.Volume.Size))).Set(float64(*nodepool.Volume.Size))

			for _, zone := range *nodepool.AvailabilityZones {
				registry.NodePoolAvailabilityZones.WithLabelValues(projectID, *clusterName, *nodepool.Name, zone).Set(1)
			}
		}

		for _, egressRange := range *cluster.Status.EgressAddressRanges {
			registry.EgressAddressRanges.WithLabelValues(projectID, *clusterName, egressRange).Set(1)
		}

		hasErrors := 0
		if cluster.Status.Errors != nil && len(*cluster.Status.Errors) > 0 {
			hasErrors = 1
		}
		registry.ClusterErrorStatus.WithLabelValues(projectID, *clusterName).Set(float64(hasErrors))
	}
}

func statusToGaugeValue(status ske.ClusterStatusState) float64 {
	switch status {
	case ske.CLUSTERSTATUSSTATE_HEALTHY:
		return 1
	// if not healthy status return 0
	default:
		return 0
	}
}
