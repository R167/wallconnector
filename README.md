# wallconnector

A super simple client which can query data from Tesla's Wall Connector v3.

This library also exposes a prometheus collector. The types are defined under
the `metrics.proto` annotations. To run it, use `go run ./cmd/prom -target <wall_connector_ip>`.
