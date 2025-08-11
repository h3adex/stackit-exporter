# IaaS Collector

The IaaS collector exposes metrics about STACKIT IaaS servers, including maintenance windows and current server state using binary per-state metrics for easy PromQL usage.

|                     |                   |
|---------------------|-------------------|
| Metric name prefix  | `stackit_server_` |
| Enabled by default? | Yes               |

---

## Metrics

| Name                                       | Description                                                                                                               | Type  | Labels                                                                                                                                                                                                                                              |
|--------------------------------------------|---------------------------------------------------------------------------------------------------------------------------|-------|-----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| stackit_server_info                        | Static metadata about a server. Always `1` if the server is seen.                                                         | Gauge | `project_id`, `server_id`, `name`, `zone`, `machine_type`, `power_status`, `server_status`, `maintenance_status`, `image_id`, `keypair_name`, `boot_volume_id`, `affinity_group`, `maintenance`, `maintenance_details`, `created_at`, `launched_at` |
| stackit_server_status_`state`              | Server lifecycle state (0/1 binary metric per state, `active`, `inactive`, `creating`, `deleting`, `rebuilding`, `error`) | Gauge | `project_id`, `server_id`, `name`, `zone`, `machine_type`                                                                                                                                                                                           |
| stackit_server_power_`state`               | Server power state (0/1 binary metric per state, `running`, `stopped`, `crashed`, `rebooting`, `error`)                   | Gauge | `project_id`, `server_id`, `name`, `zone`, `machine_type`                                                                                                                                                                                           |
| stackit_server_maintenance_`status`        | Binary `Gauge` indicating maintenance status (`planned`, `ongoing`)                                                       | Gauge | `project_id`, `server_id`, `name`, `zone`, `machine_type`                                                                                                                                                                                           |
| stackit_server_last_seen_timestamp         | Timestamp when the server was last observed (Unix epoch)                                                                  | Gauge | `project_id`, `server_id`, `name`, `zone`, `machine_type`                                                                                                                                                                                           |
| stackit_server_maintenance_start_timestamp | Start time of planned maintenance (Unix epoch)                                                                            | Gauge | `project_id`, `server_id`, `name`, `zone`, `machine_type`                                                                                                                                                                                           |
| stackit_server_maintenance_end_timestamp   | End time of planned maintenance (Unix epoch)                                                                              | Gauge | `project_id`, `server_id`, `name`, `zone`, `machine_type`                                                                                                                                                                                           |

---

## Example Metrics

```promql
stackit_server_status_error{machine_type="c2i.8",name="server1",project_id="xxx-xxx",server_id="xxx-xxx",zone="eu01-1"} 0
stackit_server_power_running{machine_type="c2i.8",name="server2",project_id="xxx-xxx",server_id="xxx-xxx",zone="eu01-1"} 1
stackit_server_last_seen_timestamp{machine_type="c2i.8",name="server1",project_id="xxx-xxx",server_id="xxx-xxx",zone="eu01-1"} 1.754917008e+09
stackit_server_maintenance_planned{machine_type="n3.14d.g1",name="server3",project_id="xxx-xxx",server_id="xxx-xxx",zone="eu01-2"} 1
stackit_server_info{affinity_group="",boot_volume_id="",created_at="2025-08-11T12:41:37Z",image_id="c8e9c49b-9c09-4426-a5dc-d35131f4cd6b",keypair_name="",launched_at="2025-08-11T12:41:47Z",machine_type="c2i.8",maintenance="",maintenance_details="",maintenance_status="",name="server-1",power_status="RUNNING",project_id="xxx-xxx",server_id="xxx-xxx",server_status="ACTIVE",zone="eu01-1"} 1
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