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
| stackit_ske_egress_address_ranges          | Egress CIDR address ranges of the cluster. Always 1; use labels.                                     | Gauge | project_id, cluster_name, cidr                                      |
| stackit_ske_cluster_error_status           | Indicates if a cluster has errors (1 if error exists, otherwise 0)                                   | Gauge | project_id, cluster_name                                            |

## Example Metric

```
stackit_ske_k8s_version{project_id="xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx",cluster_name="c1",cluster_version="1.31.10",state="supported"} 1
stackit_ske_cluster_status{project_id="xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx",cluster_name="c1",status="STATE_HEALTHY"} 1
stackit_ske_cluster_error_status{project_id="xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx",cluster_name="c1"} 0
stackit_ske_nodepool_machine_types{project_id="...", cluster_name="...", nodepool_name="np1", machine_type="c2i.8"} 1
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
    summary: "Cluster {{ $labels.cluster_name }} is in an unhealthy state: {{ $labels.status }}."
```

```yaml
- alert: SKEClusterErrorDetected
  expr: stackit_ske_cluster_error_status == 1
  for: 5m
  labels:
    severity: critical
  annotations:
    summary: "Cluster {{ $labels.cluster_name }} reports an error condition."
```

```yaml
- alert: SKEMaintenancePlannedIn24H
  expr: stackit_ske_maintenance_window_start - time() < 86400 and time() < stackit_ske_maintenance_window_start
  for: 5m
  labels:
    severity: warning
  annotations:
    summary: "Maintenance is scheduled to start in 24H for cluster {{ $labels.cluster_name }}."
```

```yaml
- alert: SKEMaintenanceInProgress
  expr: stackit_ske_maintenance_window_start <= time() and time() <= stackit_ske_maintenance_window_end
  for: 1m
  labels:
    severity: info
  annotations:
    summary: "Maintenance is currently in progress for cluster {{ $labels.cluster_name }}."
```

```yaml
- alert: DeprecatedKubernetesVersion
  expr: stackit_ske_k8s_version{state="deprecated"} == 1
  for: 10m
  labels:
    severity: warning
  annotations:
    summary: "Cluster {{ $labels.cluster_name }} is using a deprecated Kubernetes version: {{ $labels.cluster_version }}."
```

```yaml
- alert: DeprecatedMachineImageVersion
  expr: stackit_ske_nodepool_machine_version{state="deprecated"} == 1
  for: 10m
  labels:
    severity: warning
  annotations:
    summary: "Nodepool {{ $labels.nodepool_name }} is using a deprecated machine image."
```