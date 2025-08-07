# SKE Collector

The SKE collector exports metrics related to Kubernetes clusters, node pools, maintenance configuration, and runtime status in STACKIT.

|                     |                |
|---------------------|----------------|
| Metric name prefix  | `stackit_ske_` |
| Enabled by default? | Yes            |

---

## Metrics

| Name                                               | Description                                                                                          | Type  | Labels                                                            |
|----------------------------------------------------|------------------------------------------------------------------------------------------------------|-------|-------------------------------------------------------------------|
| stackit_ske_cluster_creation_timestamp             | Cluster creation time (Unix timestamp)                                                               | Gauge | `project_id`, `cluster_name`                                      |
| stackit_ske_cluster_maintenance_autoupdate_enabled | Indicates if auto-update is enabled for maintenance                                                  | Gauge | `project_id`, `cluster_name`                                      |
| stackit_ske_cluster_maintenance_start_timestamp    | Scheduled maintenance window start time (Unix timestamp)                                             | Gauge | `project_id`, `cluster_name`                                      |
| stackit_ske_cluster_maintenance_end_timestamp      | Scheduled maintenance window end time (Unix timestamp)                                               | Gauge | `project_id`, `cluster_name`                                      |
| stackit_ske_cluster_error_status                   | Indicates if a cluster has errors (`1` if error exists)                                              | Gauge | `project_id`, `cluster_name`                                      |
| stackit_ske_cluster_status_state_<state>           | Binary `Gauge` per cluster status (`healthy`, `unhealthy`, `hibernated`, `unspecified`, `deleting` ) | Gauge | `project_id`, `cluster_name`                                      |
| stackit_ske_cluster_last_seen_timestamp            | Time when the cluster was last seen by the exporter (Unix timestamp)                                 | Gauge | `project_id`, `cluster_name`                                      |
| stackit_ske_k8s_version_<state>                    | Kubernetes version state (`supported`, `deprecated`, `preview`) as binary gauge                      | Gauge | `project_id`, `cluster_name`, `k8s_version`                       |
| stackit_ske_nodepool_machine_type                  | Machine types used in node pools                                                                     | Gauge | `project_id`, `cluster_name`, `nodepool_name`, `machine_type`     |
| stackit_ske_nodepool_machine_version_<state>       | Nodepool machine image version state (`supported`, `deprecated`, `preview`) as binary gauge          | Gauge | `project_id`, `cluster_name`, `nodepool_name`, `image`, `version` |
| stackit_ske_nodepool_volume_size_gb                | Volume sizes in the node pools (in GB)                                                               | Gauge | `project_id`, `cluster_name`, `nodepool_name`, `size_gb`          |
| stackit_ske_nodepool_availability_zone             | Availability zones for node pools                                                                    | Gauge | `project_id`, `cluster_name`, `nodepool_name`, `zone`             |
| stackit_ske_nodepool_last_seen_timestamp           | Time when a node pool was last observed by the exporter (Unix timestamp)                             | Gauge | `project_id`, `cluster_name`, `nodepool_name`                     |
| stackit_ske_cluster_egress_address_range           | Egress CIDR address ranges used by the cluster. Always 1. Use `cidr` for value context               | Gauge | `project_id`, `cluster_name`, `cidr`                              |

---

## Example Metrics

```promql
stackit_ske_cluster_status_state_healthy{project_id="abc", cluster_name="f13-edu-dev"} 1
stackit_ske_nodepool_machine_version_deprecated{project_id="abc", cluster_name="f13-edu-dev", nodepool_name="gpu-pool-l40s", image="ubuntu", version="2204.20250620.0"} 0
stackit_ske_k8s_version_deprecated{project_id="abc", cluster_name="f13-edu-dev", k8s_version="1.32.5"} 1
stackit_ske_nodepool_last_seen_timestamp{project_id="abc", cluster_name="f13-edu-dev", nodepool_name="default"} 1.75457e+09
```

---

## Useful Queries

- **Healthy clusters**:
  ```promql
  stackit_ske_cluster_status_state_healthy == 1
  ```

- **Unhealthy clusters**:
  ```promql
  stackit_ske_cluster_status_state_unhealthy == 1
  ```

- **Clusters with errors**:
  ```promql
  stackit_ske_cluster_error_status == 1
  ```

- **Clusters with auto-update enabled**:
  ```promql
  stackit_ske_cluster_maintenance_autoupdate_enabled == 1
  ```

- **Upcoming maintenance**:
  ```promql
  stackit_ske_cluster_maintenance_start_timestamp > time()
  ```

- **Maintenance currently in progress**:
  ```promql
  stackit_ske_cluster_maintenance_start_timestamp <= time()
  and stackit_ske_cluster_maintenance_end_timestamp >= time()
  ```

- **Clusters by Kubernetes version**:
  ```promql
  count by(k8s_version) (stackit_ske_k8s_version_supported == 1)
  ```

- **Deprecated Kubernetes versions**:
  ```promql
  stackit_ske_k8s_version_deprecated == 1
  ```

- **Nodepools using deprecated machine images**:
  ```promql
  stackit_ske_nodepool_machine_version_deprecated == 1
  and ignoring(image, version)
    (time() - stackit_ske_nodepool_last_seen_timestamp < 600)
  ```

- **Nodepools by availability zone**:
  ```promql
  count by(zone) (stackit_ske_nodepool_availability_zone == 1)
  ```

- **Nodepools by volume size**:
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
    summary: "Cluster {{ $labels.cluster_name }} is in an unhealthy state."
    description: "Cluster {{ $labels.cluster_name }} in project {{ $labels.project_id }} has reported an unhealthy status."

- alert: SKEMaintenancePlanned
  expr: (stackit_ske_cluster_maintenance_start_timestamp - time()) < 86400
        and stackit_ske_cluster_maintenance_start_timestamp > time()
  for: 5m
  labels:
    severity: warning
  annotations:
    summary: "Scheduled maintenance within 24h for cluster {{ $labels.cluster_name }}"
    description: "Cluster {{ $labels.cluster_name }} will begin maintenance within the next 24 hours."

- alert: DeprecatedKubernetesVersion
  expr: stackit_ske_k8s_version_deprecated == 1
  for: 10m
  labels:
    severity: warning
  annotations:
    summary: "Deprecated Kubernetes version in use on cluster {{ $labels.cluster_name }}"
    description: "Cluster {{ $labels.cluster_name }} is running deprecated K8s version {{ $labels.k8s_version }}."

- alert: DeprecatedMachineImageVersion
  expr: (stackit_ske_nodepool_machine_version_deprecated == 1)
        and ignoring(image, version)
            (time() - stackit_ske_nodepool_last_seen_timestamp < 600)
  for: 10m
  labels:
    severity: warning
  annotations:
    summary: "Deprecated machine image on nodepool {{ $labels.nodepool_name }}"
    description: "Nodepool {{ $labels.nodepool_name }} in cluster {{ $labels.cluster_name }} is using a deprecated machine image '{{ $labels.version }}'."
```