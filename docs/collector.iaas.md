# IaaS Collector

The IaaS collector exposes metrics about STACKIT IaaS servers.

|                     |                  |
|---------------------|------------------|
| Metric name prefix  | `stackit_server` |
| Enabled by default? | Yes              |

## Metrics

| Name                                         | Description                                                 | Type  | Labels                                                              |
|----------------------------------------------|-------------------------------------------------------------|-------|---------------------------------------------------------------------|
| `stackit_server_maintenance_start_timestamp` | Scheduled maintenance window start time (Unix timestamp)    | Gauge | `project_id`, `server_id`, `name`, `zone`, `machine_type`           |
| `stackit_server_maintenance_end_timestamp`   | Scheduled maintenance window end time (Unix timestamp)      | Gauge | `project_id`, `server_id`, `name`, `zone`, `machine_type`           |
| `stackit_server_maintenance_status`          | Status of the maintenance window (1 for PLANNED or ONGOING) | Gauge | `project_id`, `server_id`, `name`, `zone`, `machine_type`, `status` |
| `stackit_server_status`                      | Current server lifecycle status (label: status; always 1)   | Gauge | `project_id`, `server_id`, `name`, `zone`, `status`                 |
| `stackit_server_power_status`                | Current server power status (label: power_status; always 1) | Gauge | `project_id`, `server_id`, `name`, `zone`, `power_status`           |

### Example Metric

```
stackit_server_power_status{project_id="xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx",server_id="srv-1",name="web-1",zone="eu01-01",power_status="RUNNING"} 1
stackit_server_status{project_id="xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx",server_id="srv-1",name="web-1",zone="eu01-01",status="ACTIVE"} 1
stackit_server_maintenance_status{project_id="xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx",server_id="srv-1",name="web-1",zone="eu01-01",machine_type="c2i.8",status="PLANNED"} 1
```

## Useful Queries

- **Servers in Maintenance:**
  ```promql
  stackit_server_maintenance_status{status=~"PLANNED|ONGOING"} == 1
  ```

- **Servers with Planned Maintenance:**
  ```promql
  stackit_server_maintenance_status{status="PLANNED"} == 1
  ```

- **Servers with Ongoing Maintenance:**
  ```promql
  stackit_server_maintenance_status{status="ONGOING"} == 1
  ```

- **Upcoming Maintenance Start Times:**
  ```promql
  stackit_server_maintenance_start_timestamp > time()
  ```

- **Current Maintenance Occurring Now:**
  ```promql
  stackit_server_maintenance_start_timestamp <= time() and time() <= stackit_server_maintenance_end_timestamp
  ```

- **Non-Running Servers:**
  ```promql
  stackit_server_power_status{power_status!="RUNNING"} == 1
  ```

- **Servers by Power Status:**
  ```promql
  count by (power_status) (stackit_server_power_status == 1)
  ```

- **Servers by Lifecycle Status:**
  ```promql
  count by (status) (stackit_server_status == 1)
  ```

- **Servers Not ACTIVE:**
  ```promql
  stackit_server_status{status!="ACTIVE"} == 1
  ```

- **Stopped, Crashed or Errored Servers:**
  ```promql
  stackit_server_power_status{power_status=~"STOPPED|CRASHED|ERROR"} == 1
  ```

## Alerting Examples

```yaml
- alert: ServerInMaintenance
  expr: stackit_server_maintenance_status{status=~"PLANNED|ONGOING"} == 1
  for: 10m
  labels:
    severity: warning
  annotations:
    summary: "Server {{ $labels.name }} in zone {{ $labels.zone }} is currently in maintenance ({{ $labels.status }})."
```

```yaml
- alert: MaintenanceWindowScheduled
  expr: stackit_server_maintenance_status{status="PLANNED"} == 1
  for: 1m
  labels:
    severity: info
  annotations:
    summary: "Maintenance has been scheduled for server {{ $labels.name }} in zone {{ $labels.zone }}."
```

```yaml
- alert: MaintenanceWindowStarted
  expr: stackit_server_maintenance_status{status="ONGOING"} == 1
  for: 1m
  labels:
    severity: info
  annotations:
    summary: "Maintenance has started for server {{ $labels.name }} in zone {{ $labels.zone }}."
```

```yaml
- alert: ServerNotRunning
  expr: stackit_server_power_status{power_status!="RUNNING"} == 1
  for: 5m
  labels:
    severity: critical
  annotations:
    summary: "Server {{ $labels.name }} in zone {{ $labels.zone }} is not running ({{ $labels.power_status }})."
```

```yaml
- alert: ServerNotActive
  expr: stackit_server_status{status!="ACTIVE"} == 1
  for: 10m
  labels:
    severity: warning
  annotations:
    summary: "Server {{ $labels.name }} in zone {{ $labels.zone }} is not in ACTIVE state ({{ $labels.status }})."
```