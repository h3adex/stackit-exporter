package collector

import (
	"context"
	"log"
	"strconv"
	"time"

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

	options, err := client.ListProviderOptions(ctx, region).Execute()
	if err != nil {
		log.Printf("failed to fetch Kubernetes options: %v", err)
		return
	}

	for _, cluster := range *resp.Items {
		labels := map[string]string{
			"project_id":   projectID,
			"cluster_name": *cluster.Name,
		}

		// Cluster status
		metrics.SetOneHotStatus(registry.ClusterStatus, string(*cluster.Status.Aggregated), labels)

		// Error / health
		errorValue := 0.0
		if cluster.Status.Errors != nil && len(*cluster.Status.Errors) > 0 {
			errorValue = 1.0
		}
		registry.ClusterErrorStatus.With(labels).Set(errorValue)

		// Creation time & last seen
		registry.ClusterCreationTime.With(labels).Set(float64(cluster.Status.CreationTime.UTC().Unix()))
		registry.ClusterLastSeen.With(labels).Set(float64(time.Now().Unix()))

		// Maintenance info
		registry.MaintenanceAutoUpdate.With(labels).Set(utils.BoolToFloat(*cluster.Maintenance.AutoUpdate.KubernetesVersion))
		registry.MaintenanceWindowStart.With(labels).Set(float64(cluster.Maintenance.TimeWindow.Start.UTC().Unix()))
		registry.MaintenanceWindowEnd.With(labels).Set(float64(cluster.Maintenance.TimeWindow.End.UTC().Unix()))

		// Kubernetes version support
		k8sVersion := *cluster.Kubernetes.Version
		k8sState := ""
		for _, kv := range *options.KubernetesVersions {
			if *kv.Version == k8sVersion {
				k8sState = *kv.State
			}
		}
		versionLabels := map[string]string{
			"project_id":   projectID,
			"cluster_name": *cluster.Name,
			"k8s_version":  k8sVersion,
		}
		metrics.SetOneHotStatus(registry.K8sVersion, k8sState, versionLabels)

		// Egress Ranges
		for _, r := range *cluster.Status.EgressAddressRanges {
			registry.EgressAddressRanges.WithLabelValues(projectID, *cluster.Name, r).Set(1)
		}

		// Node pools
		for _, np := range *cluster.Nodepools {
			// Machine image info
			image := *np.Machine.Image
			imageState := ""
			osName, osVersion := "", ""
			for _, img := range *options.MachineImages {
				for _, ver := range *img.Versions {
					if *ver.Version == *image.Version {
						imageState = *ver.State
						osName = *img.Name
						osVersion = *ver.Version
					}
				}
			}
			imgLabels := map[string]string{
				"project_id":    projectID,
				"cluster_name":  *cluster.Name,
				"nodepool_name": *np.Name,
				"image":         osName,
				"version":       osVersion,
			}
			metrics.SetOneHotStatus(registry.NodePoolMachineVersion, imageState, imgLabels)

			// Machine type & volume
			registry.NodePoolMachineTypes.WithLabelValues(projectID, *cluster.Name, *np.Name, *np.Machine.Type).Set(1)
			registry.NodePoolVolumeSizes.WithLabelValues(projectID, *cluster.Name, *np.Name, strconv.Itoa(int(*np.Volume.Size))).Set(float64(*np.Volume.Size))

			// Zones
			for _, zone := range *np.AvailabilityZones {
				registry.NodePoolAvailabilityZones.WithLabelValues(projectID, *cluster.Name, *np.Name, zone).Set(1)
			}

			// Last seen
			registry.NodePoolLastSeen.WithLabelValues(projectID, *cluster.Name, *np.Name).Set(float64(time.Now().Unix()))
		}
	}
}
