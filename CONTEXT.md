# Telemetry Vocabulary

| Term | Meaning |
|---|---|
| Device | An industrial or edge hardware unit that submits telemetry under one static API credential in MVP. |
| Telemetry event | One timestamped numeric measurement sent by a device. |
| Telemetry batch | A request containing 1–500 telemetry events for one authenticated device. |
| Measurement | One supported telemetry kind: temperature, voltage, battery, or pressure. |
| Aggregate | Derived five-minute processing-time statistics for one device and measurement. |
| Fleet | The collection of devices visible to internal engineering and operations teams. |
| Freshness | Whether a device has processed telemetry within the last 60 seconds. |
| Stale | A device with last processed telemetry older than 60 seconds. |
| Degraded mode | Redis is unavailable; raw-event persistence can continue, but live aggregates are paused. |
| Manual history | An explicitly user-triggered, bounded PostgreSQL query for a device’s raw events. It is not live dashboard data. |
