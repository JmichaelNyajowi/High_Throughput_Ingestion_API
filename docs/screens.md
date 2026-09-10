# Screen Inventory

## Operations

### Fleet overview

- **Screen:** Fleet telemetry
  **Purpose:** Triage current device freshness, threshold state, and headline telemetry in a dense Redis-backed table.
  **Status:** planned

### Device detail

- **Screen:** Device telemetry detail
  **Purpose:** Investigate a device’s current readings, five-minute aggregate charts, and data freshness.
  **Status:** planned

### Device history

- **Screen:** Manual device history
  **Purpose:** Run a deliberate bounded PostgreSQL query for a device’s raw telemetry.
  **Status:** planned

### Recovery states

- **Screen:** Live data degraded
  **Purpose:** Keep last known live data visible while clearly stating that Redis-backed aggregate updates are paused.
  **Status:** planned

- **Screen:** Access denied and not found
  **Purpose:** Safely recover when an operator lacks access or a destination/device cannot be found.
  **Status:** planned

All screen composition, visual tokens, responsive behavior, and state requirements are in `docs/DESIGN.md`.
