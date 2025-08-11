# SKE Collector

The SKE collector exports metrics related to Kubernetes clusters, node pools, maintenance configuration, and runtime status in STACKIT.

|                     |                |
|---------------------|----------------|
| Metric name prefix  | `stackit_ske_` |
| Enabled by default? | Yes            |

---

## Metrics

| Name                                               | Description                                                                                        | Type  | Labels                                                                                                                                                                                                                                                                                                                                                                                                                                                                   |
|----------------------------------------------------|----------------------------------------------------------------------------------------------------|-------|--------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| stackit_ske_info                                   | Cluster metadata record. Always `1`. Cluster fields are exposed as labels.                         | Gauge | `name`, `status`, `kubernetes_version`, `network_id`, `egress_address_ranges`, `pod_address_ranges`, `observability_enabled`, `observability_instance_id`, `maintenance_window_start`, `maintenance_window_end`, `maintenance_machine_image_enabled`, `maintenance_machine_kubernetes_enabled`, `credentials_rotation_last_completion_time`, `credentials_rotation_last_initiation_time`, `credentials_rotation_phase`, `hibernated`, `creation_time`, `nodepool_length` |
| stackit_ske_cluster_status_state_`state`           | One-hot cluster lifecycle states (`healthy`, `unhealthy`, `hibernated`, `unspecified`, `deleting`) | Gauge | `project_id`, `cluster_name`                                                                                                                                                                                                                                                                                                                                                                                                                                             |
| stackit_ske_cluster_error_status                   | Value `1` if cluster is in an error state                                                          | Gauge | `project_id`, `cluster_name`                                                                                                                                                                                                                                                                                                                                                                                                                                             |
| stackit_ske_cluster_creation_timestamp             | Cluster creation time (Unix timestamp)                                                             | Gauge | `project_id`, `cluster_name`                                                                                                                                                                                                                                                                                                                                                                                                                                             |
| stackit_ske_cluster_last_seen_timestamp            | Last time the exporter observed this cluster (Unix timestamp)                                      | Gauge | `project_id`, `cluster_name`                                                                                                                                                                                                                                                                                                                                                                                                                                             |
| stackit_ske_cluster_maintenance_autoupdate_enabled | `1` if auto-update is enabled for maintenance                                                      | Gauge | `project_id`, `cluster_name`                                                                                                                                                                                                                                                                                                                                                                                                                                             |
| stackit_ske_cluster_maintenance_start_timestamp    | Maintenance window start time (Unix timestamp)                                                     | Gauge | `project_id`, `cluster_name`                                                                                                                                                                                                                                                                                                                                                                                                                                             |
| stackit_ske_cluster_maintenance_end_timestamp      | Maintenance window end time (Unix timestamp)                                                       | Gauge | `project_id`, `cluster_name`                                                                                                                                                                                                                                                                                                                                                                                                                                             |
| stackit_ske_cluster_egress_address_range           | Egress CIDR used by the cluster. Value always `1`, `cidr` is context                               | Gauge | `project_id`, `cluster_name`, `cidr`                                                                                                                                                                                                                                                                                                                                                                                                                                     |
| stackit_ske_k8s_version_`state`                    | Kubernetes version state as one-hot binary (`supported`, `deprecated`, `preview`)                  | Gauge | `project_id`, `cluster_name`, `k8s_version`                                                                                                                                                                                                                                                                                                                                                                                                                              |
| stackit_ske_nodepool_machine_version_`state`       | Node pool image version status (`supported`, `deprecated`, `preview`)                              | Gauge | `project_id`, `cluster_name`, `nodepool_name`, `image`, `version`                                                                                                                                                                                                                                                                                                                                                                                                        |
| stackit_ske_nodepool_machine_type                  | Machine type for a node pool. Value always 1                                                       | Gauge | `project_id`, `cluster_name`, `nodepool_name`, `machine_type`                                                                                                                                                                                                                                                                                                                                                                                                            |
| stackit_ske_nodepool_availability_zone             | Availability zone for node pool. Value always 1                                                    | Gauge | `project_id`, `cluster_name`, `nodepool_name`, `zone`                                                                                                                                                                                                                                                                                                                                                                                                                    |
| stackit_ske_nodepool_volume_size_gb                | Volume size in node pool (in GB). Value represents size                                            | Gauge | `project_id`, `cluster_name`, `nodepool_name`, `size_gb`                                                                                                                                                                                                                                                                                                                                                                                                                 |
| stackit_ske_nodepool_last_seen_timestamp           | Last time node pool was observed (Unix timestamp)                                                  | Gauge | `project_id`, `cluster_name`, `nodepool_name`                                                                                                                                                                                                                                                                                                                                                                                                                            |

---

## Example Metrics

```promql
stackit_ske_cluster_status_state_healthy{project_id="abc", cluster_name="test-ske-dev"} 1
stackit_ske_nodepool_machine_version_deprecated{project_id="abc", cluster_name="test-ske-dev", nodepool_name="gpu-pool-l40s", image="ubuntu", version="2204.20250620.0"} 0
stackit_ske_k8s_version_deprecated{project_id="abc", cluster_name="test-ske-dev", k8s_version="1.32.5"} 1
stackit_ske_nodepool_last_seen_timestamp{project_id="abc", cluster_name="test-ske-dev", nodepool_name="default"} 1.75457e+09
```

---

## Useful Queries

-  Healthy clusters
  ```promql
  stackit_ske_cluster_status_state_healthy == 1
  ```

- Unhealthy clusters
  ```promql
  stackit_ske_cluster_status_state_unhealthy == 1
  ```

- Clusters with errors
  ```promql
  stackit_ske_cluster_error_status == 1
  ```

- Clusters with auto-update enabled
  ```promql
  stackit_ske_cluster_maintenance_autoupdate_enabled == 1
  ```

- Upcoming maintenance
  ```promql
  stackit_ske_cluster_maintenance_start_timestamp > time()
  ```

- Maintenance ongoing
  ```promql
  stackit_ske_cluster_maintenance_start_timestamp <= time()
  and stackit_ske_cluster_maintenance_end_timestamp >= time()
  ```

- Supported Kubernetes versions
  ```promql
  count by(k8s_version) (stackit_ske_k8s_version_supported == 1)
  ```

- Deprecated Kubernetes versions
  ```promql
  stackit_ske_k8s_version_deprecated == 1
  ```

- Node pools using deprecated machine images (seen in last 10 min)
  ```promql
  stackit_ske_nodepool_machine_version_deprecated == 1
  and ignoring(image, version)
      (time() - stackit_ske_nodepool_last_seen_timestamp < 600)
  ```

- Node pools by availability zone
  ```promql
  count by(zone) (stackit_ske_nodepool_availability_zone == 1)
  ```

- Node pools by volume size
  ```promql
  count by(size_gb) (stackit_ske_nodepool_volume_size_gb)
  ```

---

## Alerting Examples

```yaml
- alert: SKEClusterUnhealthy
  expr: stackit_ske_cluster_status_state_unhealthy == 1
  for: 5m
  labels:
    severity: critical
  annotations:
    summary: "Cluster {{ $labels.cluster_name }} is unhealthy"
    description: "Cluster {{ $labels.cluster_name }} in project {{ $labels.project_id }} has an unhealthy status."

- alert: SKEMaintenancePlanned
  expr: (stackit_ske_cluster_maintenance_start_timestamp - time()) < 86400
    and stackit_ske_cluster_maintenance_start_timestamp > time()
  for: 5m
  labels:
    severity: warning
  annotations:
    summary: "Cluster maintenance scheduled within 24h"
    description: "Cluster {{ $labels.cluster_name }} will undergo maintenance within the next 24 hours."

- alert: DeprecatedKubernetesVersion
  expr: stackit_ske_k8s_version_deprecated == 1
  for: 10m
  labels:
    severity: warning
  annotations:
    summary: "Deprecated Kubernetes version"
    description: "Cluster {{ $labels.cluster_name }} is running deprecated Kubernetes version {{ $labels.k8s_version }}."

- alert: DeprecatedMachineImageVersion
  expr: (stackit_ske_nodepool_machine_version_deprecated == 1)
    and ignoring(image, version)
    (time() - stackit_ske_nodepool_last_seen_timestamp < 600)
  for: 10m
  labels:
    severity: warning
  annotations:
    summary: "Deprecated machine image used in nodepool"
    description: "Nodepool {{ $labels.nodepool_name }} in cluster {{ $labels.cluster_name }} is using deprecated version {{ $labels.version }}."

- alert: SKEClusterErrorDetected
  expr: (stackit_ske_cluster_error_status == 1)
    and (time() - stackit_ske_cluster_last_seen_timestamp < 300)
  for: 5m
  labels:
    severity: critical
  annotations:
    summary: "Cluster in error state"
    description: "Cluster {{ $labels.cluster_name }} is in an ERROR state and observed in the last 5 minutes."
```