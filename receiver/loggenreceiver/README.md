# Loggen Receiver

|## Configuration

| Field      | Default          | Description                                                  |
| ---------- | ---------------- | ------------------------------------------------------------ |
| `logs_per_sec`      | `10_000`               | Number of logs to generate per second  |
| `log_line`      | A 1KB sized log line                | Text to print as the log line  |
| `emit_interval`    | `100ms`         | The interval at which logs are emitted to the next component in the pipeline |


## Example Configurations

```yaml
receivers:
  loggen:
    logs_per_sec: 1000000
```

