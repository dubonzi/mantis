# Configuration

You can configure Mantis either by envinroment variables or by passing arguments when running the executable.

| Env                      | Arg                       | Default          |                                                                      |
| ------------------------ | ------------------------- | ---------------- | -------------------------------------------------------------------- |
| `SERVER_PORT`            | `-server.port`            | `8080`           | Port Mantis runs on                                                  |
| `HEALTH_PORT`            | `-health.port`            | `8081`           | Health check port                                                    |
| `LOADER_PATH_MAPPING`    | `-loader.path.mapping`    | `files/mapping`  | Path to mapping files                                                |
| `LOADER_PATH_RESPONSE`   | `-loader.path.response`   | `files/response` | Path to response files                                               |
| `OTEL_ENABLED`           | `-otel.enabled`           | `true`           | Enable/disable Opentelemetry support                                 |
| `OTEL_EXPORTER_PROTOCOL` | `-otel.exporter.protocol` | `http`           | Protocol used to export Opentelemetry traces and metrics (http/grpc) |
| `OTEL_EXPORTER_ENDPOINT` | `-otel.exporter.endpoint` | `localhost:4318` | Endpoint of the collecter for Opentelemetry traces and metrics       |
| `OTEL_EXPORTER_INSECURE` | `-otel.exporter.insecure` | `true`           | Use insecure connection for Opentelemetry exporter                   |
| `LOG_LEVEL`              | `-log.level`              | `INFO`           | Log level                                                            |
| `LOG_FORMAT`             | `-log.format`             | `TEXT`           | Log format (TEXT/JSON)                                               |