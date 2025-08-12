package collector

import (
	"context"
	"log"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/h3adex/stackit-exporter/internal/metrics"
	"github.com/stackitcloud/stackit-sdk-go/services/ske"
)

type SkeClient interface {
	ListClusters(ctx context.Context, projectID string, region string) ske.ApiListClustersRequest
	ListProviderOptions(ctx context.Context, region string) ske.ApiListProviderOptionsRequest
}

func ScrapeSkeAPI(ctx context.Context, client SkeClient, projectID, region string, registry *metrics.SKERegistry) {
	resp, err := client.ListClusters(ctx, projectID, region).Execute()
	if err != nil {
		log.Printf("error fetching clusters: %v", err)
		return
	}
	if resp == nil || resp.Items == nil {
		log.Println("cluster list is empty or nil")
		return
	}

	options, err := client.ListProviderOptions(ctx, region).Execute()
	if err != nil {
		log.Printf("error fetching provider options: %v", err)
		return
	}

	nowUnix := float64(time.Now().Unix())

	for _, cluster := range SafeSlice(resp.Items) {
		if cluster.Name == nil {
			continue
		}

		clusterName := *cluster.Name
		sharedLabels := prometheus.Labels{
			"project_id":   projectID,
			"cluster_name": clusterName,
		}

		infoLabels := prometheus.Labels{
			"creation_time": "",
			"credentials_rotation_last_completion_time": "",
			"credentials_rotation_last_initiation_time": "",
			"credentials_rotation_phase":                "",
			"egress_address_ranges":                     "",
			"hibernated":                                "false",
			"kubernetes_version":                        "",
			"maintenance_machine_image_enabled":         "false",
			"maintenance_machine_kubernetes_enabled":    "false",
			"maintenance_window_end":                    "",
			"maintenance_window_start":                  "",
			"name":                                      clusterName,
			"network_id":                                "",
			"nodepool_length":                           strconv.Itoa(len(SafeSlice(cluster.Nodepools))),
			"observability_enabled":                     "false",
			"observability_instance_id":                 "",
			"pod_address_ranges":                        "",
			"status":                                    "",
		}

		if cluster.Kubernetes != nil {
			infoLabels["kubernetes_version"] = SafeString(cluster.Kubernetes.Version)
		}

		if cluster.Network != nil {
			infoLabels["network_id"] = SafeString(cluster.Network.Id)
		}

		if cluster.Maintenance != nil {
			if cluster.Maintenance.AutoUpdate != nil {
				infoLabels["maintenance_machine_image_enabled"] = strconv.FormatBool(SafeBool(cluster.Maintenance.AutoUpdate.MachineImageVersion))
				infoLabels["maintenance_machine_kubernetes_enabled"] = strconv.FormatBool(SafeBool(cluster.Maintenance.AutoUpdate.KubernetesVersion))
			}
			if cluster.Maintenance.TimeWindow != nil {
				infoLabels["maintenance_window_start"] = SafeTime(cluster.Maintenance.TimeWindow.Start)
				infoLabels["maintenance_window_end"] = SafeTime(cluster.Maintenance.TimeWindow.End)
			}
		}

		if cluster.Status != nil {
			infoLabels["creation_time"] = SafeTime(cluster.Status.CreationTime)
			infoLabels["pod_address_ranges"] = SafeJoin(cluster.Status.PodAddressRanges)
			infoLabels["hibernated"] = strconv.FormatBool(SafeBool(cluster.Status.Hibernated))
			infoLabels["status"] = SafeString((*string)(cluster.Status.Aggregated))
			infoLabels["egress_address_ranges"] = SafeJoin(cluster.Status.EgressAddressRanges)

			if cluster.Status.CredentialsRotation != nil {
				infoLabels["credentials_rotation_last_completion_time"] = SafeTime(cluster.Status.CredentialsRotation.LastCompletionTime)
				infoLabels["credentials_rotation_last_initiation_time"] = SafeTime(cluster.Status.CredentialsRotation.LastInitiationTime)
				infoLabels["credentials_rotation_phase"] = SafeString((*string)(cluster.Status.CredentialsRotation.Phase))
			}
		}

		if cluster.Extensions != nil && cluster.Extensions.Observability != nil {
			infoLabels["observability_enabled"] = strconv.FormatBool(SafeBool(cluster.Extensions.Observability.Enabled))
			infoLabels["observability_instance_id"] = SafeString(cluster.Extensions.Observability.InstanceId)
		}

		registry.ClusterInfo.With(infoLabels).Set(1)

		if cluster.Status != nil && cluster.Status.Aggregated != nil {
			statusLabels := prometheus.Labels{
				"project_id":   projectID,
				"cluster_name": clusterName,
				"status":       string(*cluster.Status.Aggregated),
			}
			registry.ClusterStatus.With(statusLabels).Set(1)
		}

		errorValue := 0.0
		if cluster.Status != nil && cluster.Status.Errors != nil && len(*cluster.Status.Errors) > 0 {
			errorValue = 1.0
		}
		registry.ClusterErrorStatus.With(sharedLabels).Set(errorValue)

		if cluster.Status != nil && cluster.Status.CreationTime != nil {
			registry.ClusterCreationTime.With(sharedLabels).Set(float64(cluster.Status.CreationTime.UTC().Unix()))
		}

		registry.ClusterLastSeen.With(sharedLabels).Set(nowUnix)

		if cluster.Maintenance != nil {
			if cluster.Maintenance.AutoUpdate != nil && cluster.Maintenance.AutoUpdate.KubernetesVersion != nil {
				registry.MaintenanceAutoUpdate.With(sharedLabels).Set(BoolToFloat(*cluster.Maintenance.AutoUpdate.KubernetesVersion))
			}
			if cluster.Maintenance.TimeWindow != nil {
				if cluster.Maintenance.TimeWindow.Start != nil {
					registry.MaintenanceWindowStart.With(sharedLabels).Set(float64(cluster.Maintenance.TimeWindow.Start.UTC().Unix()))
				}
				if cluster.Maintenance.TimeWindow.End != nil {
					registry.MaintenanceWindowEnd.With(sharedLabels).Set(float64(cluster.Maintenance.TimeWindow.End.UTC().Unix()))
				}
			}
		}

		k8sVersion := SafeString(cluster.Kubernetes.Version)
		k8sStatus := ""
		for _, kv := range SafeSlice(options.KubernetesVersions) {
			if SafeString(kv.Version) == k8sVersion {
				k8sStatus = SafeString(kv.State)
				break
			}
		}
		k8sLabels := prometheus.Labels{
			"project_id":   projectID,
			"cluster_name": clusterName,
			"k8s_version":  k8sVersion,
			"status":       k8sStatus,
		}
		registry.K8sVersion.With(k8sLabels).Set(1)

		for _, cidr := range SafeSlice(cluster.Status.EgressAddressRanges) {
			registry.EgressAddressRanges.WithLabelValues(projectID, clusterName, cidr).Set(1)
		}

		for _, np := range SafeSlice(cluster.Nodepools) {
			if np.Name == nil {
				continue
			}
			nodepoolName := *np.Name

			imgName, imgVersion, imageStatus := "", "", ""
			if np.Machine != nil && np.Machine.Image != nil {
				imgName = SafeString(np.Machine.Image.Name)
				imgVersion = SafeString(np.Machine.Image.Version)

				for _, img := range SafeSlice(options.MachineImages) {
					for _, ver := range SafeSlice(img.Versions) {
						if SafeString(ver.Version) == imgVersion {
							imageStatus = SafeString(ver.State)
							break
						}
					}
				}
			}

			imgLabels := prometheus.Labels{
				"project_id":    projectID,
				"cluster_name":  clusterName,
				"nodepool_name": nodepoolName,
				"image":         imgName,
				"version":       imgVersion,
				"status":        imageStatus,
			}
			registry.NodePoolMachineVersion.With(imgLabels).Set(1)

			if np.Machine != nil && np.Machine.Type != nil {
				registry.NodePoolMachineTypes.WithLabelValues(projectID, clusterName, nodepoolName, *np.Machine.Type).Set(1)
			}

			if np.Volume != nil && np.Volume.Size != nil {
				size := *np.Volume.Size
				registry.NodePoolVolumeSizes.WithLabelValues(projectID, clusterName, nodepoolName, strconv.Itoa(int(size))).Set(float64(size))
			}

			for _, zone := range SafeSlice(np.AvailabilityZones) {
				registry.NodePoolAvailabilityZones.WithLabelValues(projectID, clusterName, nodepoolName, zone).Set(1)
			}

			registry.NodePoolLastSeen.WithLabelValues(projectID, clusterName, nodepoolName).Set(nowUnix)
		}
	}
}
