# IaaS Collector

The IaaS collector exposes metrics about STACKIT IaaS servers, including maintenance windows and current server state using binary per-state metrics for easy PromQL usage.

|                     |                   |
|---------------------|-------------------|
| Metric name prefix  | `stackit_server_` |
| Enabled by default? | Yes               |

---

## Metrics

| Name                                       | Description                                                                                                               | Type  | Labels                                                                                                                                                                                       |
|--------------------------------------------|---------------------------------------------------------------------------------------------------------------------------|-------|----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| stackit_server_info                        | Static metadata for each observed server (value is always 1). Useful for enrichment and dashboards.                       | Gauge | `project_id`, `server_id`, `name`, `zone`, `machine_type`, `image_id`, `keypair_name`, `boot_volume_id`, `affinity_group`, `created_at`, `launched_at`, `maintenance`, `maintenance_details` |
| stackit_server_status_`state`              | Server lifecycle state (0/1 binary metric per state, `active`, `inactive`, `creating`, `deleting`, `rebuilding`, `error`) | Gauge | `project_id`, `server_id`, `name`, `zone`, `machine_type`                                                                                                                                    |
| stackit_server_power_`state`               | Server power state (0/1 binary metric per state, `running`, `stopped`, `crashed`, `rebooting`, `error`)                   | Gauge | `project_id`, `server_id`, `name`, `zone`, `machine_type`                                                                                                                                    |
| stackit_server_maintenance_start_timestamp | Scheduled maintenance window start time (Unix timestamp)                                                                  | Gauge | `project_id`, `server_id`, `name`, `zone`, `machine_type`                                                                                                                                    |
| stackit_server_maintenance_end_timestamp   | Scheduled maintenance window end time (Unix timestamp)                                                                    | Gauge | `project_id`, `server_id`, `name`, `zone`, `machine_type`                                                                                                                                    |
| stackit_server_maintenance_planned         | Maintenance has been scheduled but is not yet started (1 if true, 0 otherwise)                                            | Gauge | `project_id`, `server_id`, `name`, `zone`, `machine_type`                                                                                                                                    |
| stackit_server_maintenance_ongoing         | Server is currently under active maintenance (1 if true, 0 otherwise)                                                     | Gauge | `project_id`, `server_id`, `name`, `zone`, `machine_type`                                                                                                                                    |
| stackit_server_last_seen_timestamp         | Last time the server was observed by the exporter (Unix timestamp)                                                        | Gauge | `project_id`, `server_id`, `name`, `zone`, `machine_type`                                                                                                                                    |

---

## Example Metrics

```promql
stackit_server_info{project_id="abc", server_id="srv-1", name="web-1", zone="eu01-1", machine_type="c2.3", image_id="ubuntu-22.04", created_at="2024-05-01T11:14:00Z"} 1
stackit_server_power_running{project_id="abc", server_id="srv-1", name="web-1", zone="eu01-1", machine_type="c2.3"} 1
stackit_server_status_active{project_id="abc", server_id="srv-1", name="web-1", zone="eu01-1", machine_type="c2.3"} 1
stackit_server_power_stopped{project_id="abc", server_id="srv-2", name="batch", zone="eu01-1", machine_type="c2.3"} 1
stackit_server_maintenance_planned{project_id="abc", server_id="srv-1", name="web-1", zone="eu01-1", machine_type="c2.3"} 1
stackit_server_maintenance_ongoing{project_id="abc", server_id="srv-3", name="db-1", zone="eu01-1", machine_type="c2.6"} 1
stackit_server_last_seen_timestamp{project_id="abc", server_id="srv-1", name="web-1", zone="eu01-1", machine_type="c2.3"} 1.75449502e+09
```

---

## Useful Queries

- Servers that are currently running:
  ```promql
  stackit_server_power_running == 1
  ```

- Stopped, Crashed, or in Error:
  ```promql
  stackit_server_power_stopped == 1
  or stackit_server_power_crashed == 1
  or stackit_server_power_error == 1
  ```

- Servers not ACTIVE:
  ```promql
  stackit_server_status_active == 0
  ```

- All servers by current lifecycle state:
  ```promql
  count by(machine_type, zone) (
    stackit_server_status_active
    or stackit_server_status_inactive
    or stackit_server_status_error
  )
  ```

- All servers by power state:
  ```promql
  count by(zone) (
    stackit_server_power_running
    or stackit_server_power_stopped
    or stackit_server_power_crashed
    or stackit_server_power_error
  )
  ```

- Upcoming server maintenance (raw timestamp):
  ```promql
  stackit_server_maintenance_start_timestamp > time()
  ```

- Current maintenance (timestamp-based):
  ```promql
  stackit_server_maintenance_start_timestamp <= time()
  and stackit_server_maintenance_end_timestamp >= time()
  ```

- Maintenance currently ongoing (boolean-based):
  ```promql
  stackit_server_maintenance_ongoing == 1
  ```

- Servers with upcoming maintenance scheduled:
  ```promql
  stackit_server_maintenance_planned == 1
  ```

---

## Alerting Examples

```yaml
- alert: ServerNotRunning
  expr: (stackit_server_power_running == 0 and time() - stackit_server_last_seen_timestamp < 600)
  for: 10m
  labels:
    severity: critical
  annotations:
    summary: "Server {{ $labels.name }} is not running."
    description: "Server {{ $labels.name }} in zone {{ $labels.zone }} has not been in the RUNNING state for more than 10 minutes."

- alert: ServerNotActive
  expr: (stackit_server_status_active == 0 and time() - stackit_server_last_seen_timestamp < 600)
  for: 10m
  labels:
    severity: warning
  annotations:
    summary: "Server {{ $labels.name }} is not ACTIVE."
    description: "Server {{ $labels.name }} in zone {{ $labels.zone }} is not in ACTIVE lifecycle state."

- alert: MaintenanceScheduledSoon
  expr: (stackit_server_maintenance_start_timestamp - time()) < 86400
    and stackit_server_maintenance_planned == 1
  for: 1m
  labels:
    severity: info
  annotations:
    summary: "Server {{ $labels.name }} has maintenance scheduled within 24h."
    description: "Server {{ $labels.name }} maintenance begins at {{ $value | humanizeTimestamp }}."

- alert: MaintenanceOngoingNow
  expr: stackit_server_maintenance_ongoing == 1
  for: 5m
  labels:
    severity: info
  annotations:
    summary: "Server {{ $labels.name }} is under maintenance."
    description: "Current maintenance is active for server {{ $labels.name }} in zone {{ $labels.zone }}."

- alert: ServerMaintenanceChanged
  expr: changes(stackit_server_maintenance_start_timestamp[15m]) > 0
  for: 0m
  labels:
    severity: info
  annotations:
    summary: "Maintenance start time changed for server {{ $labels.name }}."
    description: "The scheduled maintenance for server {{ $labels.name }} in {{ $labels.zone }} has changed."
```