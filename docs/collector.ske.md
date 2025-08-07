# SKE Collector

The SKE collector exports metrics related to Kubernetes clusters, node pools, maintenance configuration, and runtime status.

|                     |                |
|---------------------|----------------|
| Metric name prefix  | `stackit_ske_` |
| Enabled by default? | Yes            |

## Metrics

| Name                                       | Description                                                                                          | Type  | Labels                                                              |
|--------------------------------------------|------------------------------------------------------------------------------------------------------|-------|---------------------------------------------------------------------|
| stackit_ske_k8s_version                    | Kubernetes version in use (value always 1). `state` = supported/deprecated/preview                   | Gauge | project_id, cluster_name, cluster_version, state                    |
| stackit_ske_cluster_status                 | Cluster status (1 if status is present). Use label 'status' to identify state such as STATE_HEALTHY. | Gauge | project_id, cluster_name, status                                    |
| stackit_ske_cluster_creation_timestamp     | Cluster creation time (Unix timestamp)                                                               | Gauge | project_id, cluster_name                                            |
| stackit_ske_maintenance_autoupdate_enabled | Indicates if auto-update is enabled for maintenance                                                  | Gauge | project_id, cluster_name                                            |
| stackit_ske_maintenance_window_start       | Scheduled maintenance window start time (Unix timestamp)                                             | Gauge | project_id, cluster_name                                            |
| stackit_ske_maintenance_window_end         | Scheduled maintenance window end time (Unix timestamp)                                               | Gauge | project_id, cluster_name                                            |
| stackit_ske_nodepool_machine_types         | Machine types used in node pools. Always 1; use labels for details.                                  | Gauge | project_id, cluster_name, nodepool_name, machine_type               |
| stackit_ske_nodepool_machine_version       | Machine image version in use (value always 1). `state` = supported/deprecated/preview                | Gauge | project_id, cluster_name, nodepool_name, os_name, os_version, state |
| stackit_ske_nodepool_volume_sizes_gb       | Volume sizes in the node pools (in GB)                                                               | Gauge | project_id, cluster_name, nodepool_name, volume_size                |
| stackit_ske_nodepool_availability_zones    | Availability zones for node pools. Always 1; use labels.                                             | Gauge | project_id, cluster_name, nodepool_name, zone                       |
| stackit_ske_nodepool_last_seen             | Last time the node pool was observed/updated (Unix timestamp)                                        | Gauge | project_id, cluster_name, nodepool_name                             |
| stackit_ske_egress_address_ranges          | Egress CIDR address ranges of the cluster. Always 1; use labels.                                     | Gauge | project_id, cluster_name, cidr                                      |
| stackit_ske_cluster_error_status           | Indicates if a cluster has errors (1 if error exists, otherwise 0)                                   | Gauge | project_id, cluster_name                                            |

## Example Metric

```
stackit_ske_k8s_version{project_id="xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx",cluster_name="c1",cluster_version="1.31.10",state="supported"} 1
stackit_ske_cluster_status{project_id="xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx",cluster_name="c1",status="STATE_HEALTHY"} 1
stackit_ske_cluster_error_status{project_id="xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx",cluster_name="c1"} 0
stackit_ske_nodepool_machine_types{project_id="xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx", cluster_name="c1", nodepool_name="np1", machine_type="c2i.8"} 1
stackit_ske_nodepool_last_seen{project_id="xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx",cluster_name="c1",nodepool_name="np1"} 1717093012
```

## Useful Queries

- **Healthy clusters**:
  ```promql
  stackit_ske_cluster_status{status="STATE_HEALTHY"} == 1
  ```

- **Unhealthy clusters**:
  ```promql
  stackit_ske_cluster_status{status!="STATE_HEALTHY"} == 1
  ```

- **Clusters with errors**:
  ```promql
  stackit_ske_cluster_error_status == 1
  ```

- **Clusters with auto-update enabled**:
  ```promql
  stackit_ske_maintenance_autoupdate_enabled == 1
  ```

- **Upcoming maintenance**:
  ```promql
  stackit_ske_maintenance_window_start > time()
  ```

- **Maintenance currently in progress**:
  ```promql
  stackit_ske_maintenance_window_start <= time()
  and stackit_ske_maintenance_window_end >= time()
  ```

- **Clusters by Kubernetes version**:
  ```promql
  count by(cluster_version, state) (stackit_ske_k8s_version == 1)
  ```

- **Node pools by machine image version**:
  ```promql
  count by(state) (stackit_ske_nodepool_machine_version == 1)
  ```

- **Node pools by volume size**:
  ```promql
  count by(volume_size) (stackit_ske_nodepool_volume_sizes_gb)
  ```

- **Node pools by availability zone**:
  ```promql
  count by(zone) (stackit_ske_nodepool_availability_zones == 1)
  ```

- **Clusters by CIDR**:
  ```promql
  count by(cidr) (stackit_ske_egress_address_ranges == 1)
  ```

## Alerting Examples

```yaml
- alert: SKEClusterUnhealthy
  expr: stackit_ske_cluster_status{status!="STATE_HEALTHY"} == 1
  for: 5m
  labels:
    severity: critical
  annotations:
    summary: "Cluster {{ $labels.cluster_name }} is in an unhealthy state ({{ $labels.status }})."
    description: "Cluster {{ $labels.cluster_name }} in project {{ $labels.project_id }} is not reporting a healthy status for at least 5 minutes."
```

```yaml
- alert: SKEClusterErrorDetected
  expr: stackit_ske_cluster_error_status == 1
  for: 5m
  labels:
    severity: critical
  annotations:
    summary: "Cluster {{ $labels.cluster_name }} reports an error condition."
    description: "Cluster {{ $labels.cluster_name }} in project {{ $labels.project_id }} is reporting one or more errors in its current status."
```

```yaml
- alert: SKEMaintenancePlannedIn24H
  expr: (stackit_ske_maintenance_window_start - time()) < 86400 and time() < stackit_ske_maintenance_window_start
  for: 5m
  labels:
    severity: warning
  annotations:
    summary: "Maintenance is planned for cluster {{ $labels.cluster_name }} within 24 hours."
    description: "Scheduled maintenance for cluster {{ $labels.cluster_name }} is beginning at {{ $labels.start_time }}."
```

```yaml
- alert: SKEMaintenanceInProgress
  expr: (stackit_ske_maintenance_window_start <= time()) and (time() <= stackit_ske_maintenance_window_end)
  for: 1m
  labels:
    severity: info
  annotations:
    summary: "Maintenance is in progress for cluster {{ $labels.cluster_name }}."
    description: "Cluster {{ $labels.cluster_name }} is currently within its scheduled maintenance window."
```

```yaml
- alert: DeprecatedKubernetesVersion
  expr: stackit_ske_k8s_version{state!="supported"} == 1
  for: 10m
  labels:
    severity: warning
  annotations:
    summary: "Cluster {{ $labels.cluster_name }} is using a {{ $labels.state }} Kubernetes version."
    description: "The version {{ $labels.cluster_version }} is marked as {{ $labels.state }}. Consider upgrading your cluster {{ $labels.cluster_name }}."
```

```yaml
- alert: DeprecatedMachineImageVersion
  expr: (stackit_ske_nodepool_machine_version{state!="supported"} == 1) and (stackit_ske_nodepool_last_seen{state!="supported"} > time() - 600)
  for: 10m
  labels:
    severity: warning
  annotations:
    summary: "Nodepool {{ $labels.nodepool_name }} is using a {{ $labels.state }} machine image."
    description: "Nodepool {{ $labels.nodepool_name }} in cluster {{ $labels.cluster_name }} is running a machine image marked as '{{ $labels.state }}'. It was updated in the last 10 minutes, so this alert should be considered valid."
```