# Configuration

All configuration lives in a YAML file. Durations accept Go-style strings like `5s`, `10m`, `1h`.

## Full YAML template

```yaml
monitoring:
  lookback_duration: "1h"      # How far back to scan logs each cycle
  check_interval: "30s"        # Scan interval
  enable_real_time: true        # Reserved for future real-time tailing
  log_buffer_size: 1000         # Buffer size for log events

blocking:
  failure_threshold: 5          # Attempts per IP before blocking
  block_duration: "20h"         # Block duration (0 = permanent)
  max_concurrent_blocks: 1000   # Safety cap for active blocks
  whitelisted_ips:              # IPs or CIDR ranges to skip
    - "127.0.0.1"
    - "::1"
    - "192.168.1.0/24"
  auto_unblock: true            # Remove blocks after expiration
  cleanup_interval: "5m"        # Cleanup cadence
  rule_name_template: "Guardian - {ip} - {timestamp}"

logging:
  level: "info"                # debug | info | warn | error
  format: "text"               # text | json
  output: "stdout"             # stdout | stderr
  enable_file: true             # Write to file if file_path is set
  file_path: "C:\\ProgramData\\Guardian\\logs\\guardian.log"
  enable_contextual: true       # Include source context in logs
  log_event_lookups: false      # Event scan summaries
  log_firewall_actions: true    # Block/unblock logs
  log_attack_attempts: true     # Parsed attack attempts
  log_monitoring_events: false  # Monitoring start/stop
  log_cleanup_events: true      # Cleanup actions

storage:
  type: "sqlite"                # memory | sqlite (sqlite planned)
  file_path: "C:\\ProgramData\\Guardian\\data\\guardian.db"

services:
  - name: "RDP"
    log_path: "Security"        # Windows Event Log name
    log_pattern: "4625"         # Failed logon event ID
    custom_threshold: 0         # Override failure_threshold if > 0
    enabled: true
```

## Field reference

### monitoring
- `lookback_duration`: Sliding window size for log scans.
- `check_interval`: How often to scan.
- `enable_real_time`: Reserved for real-time tailing.
- `log_buffer_size`: Buffer size for log events (future use).

### blocking
- `failure_threshold`: Attempts per IP required to block.
- `block_duration`: How long to block (0 = permanent).
- `max_concurrent_blocks`: Safety cap on active blocks.
- `whitelisted_ips`: IPs/CIDR ranges to never block.
- `auto_unblock`: Whether to remove expired blocks automatically.
- `cleanup_interval`: Cleanup cadence for expired blocks.
- `rule_name_template`: Rule name template. Placeholders: `{app}`, `{ip}`, `{timestamp}`, `{service}`.

Note: Firewall rules created by Guardian include a description tag `GuardianTag=Guardian` to allow de-duplication and identification.

### logging
- `level`: Log level (debug/info/warn/error).
- `format`: `text` or `json`.
- `output`: `stdout` or `stderr`.
- `enable_file`: Write to file when `file_path` is set.
- `file_path`: Absolute log file path.
- `enable_contextual`: Adds file/function context.
- `log_event_lookups`: Event scan summaries.
- `log_firewall_actions`: Block/unblock events.
- `log_attack_attempts`: Parsed attack attempts.
- `log_monitoring_events`: Start/stop events.
- `log_cleanup_events`: Cleanup details.

### storage
- `type`: `memory` or `sqlite` (sqlite planned).
- `file_path`: Database path (used for sqlite).

### services
Each item defines a monitored service:
- `name`: Service name (e.g., RDP, SSH, IIS).
- `log_path`: Log path or Windows Event Log name.
- `log_pattern`: Pattern or event ID to match.
- `custom_threshold`: Overrides `blocking.failure_threshold` if > 0.
- `enabled`: Enable/disable monitoring for the service.
