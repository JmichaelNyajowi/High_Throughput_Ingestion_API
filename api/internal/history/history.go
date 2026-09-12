package history

import (
	"context"
	"errors"
	"github.com/jackc/pgx/v5/pgxpool"
	"time"
)

const maxRange = 24 * time.Hour

type Event struct {
	EventID, DeviceID, MeasurementType, Unit string
	Value                                    float64
	EventTimestamp, ReceivedAt               time.Time
}

func Query(ctx context.Context, pool *pgxpool.Pool, device string, from, to time.Time, limit int, measurement string) ([]Event, error) {
	if device == "" || !from.Before(to) || to.Sub(from) > maxRange || limit < 1 || limit > 1000 {
		return nil, errors.New("invalid history bounds")
	}
	c, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()
	rows, err := pool.Query(c, `SELECT e.event_id,d.external_id,e.measurement_type,e.value,e.unit,e.event_timestamp,e.received_at FROM telemetry_events e JOIN devices d ON d.id=e.device_id WHERE d.external_id=$1 AND e.event_timestamp >= $2 AND e.event_timestamp < $3 AND ($4 = '' OR e.measurement_type = $4) ORDER BY e.event_timestamp DESC,e.id DESC LIMIT $5`, device, from, to, measurement, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []Event
	for rows.Next() {
		var e Event
		if err := rows.Scan(&e.EventID, &e.DeviceID, &e.MeasurementType, &e.Value, &e.Unit, &e.EventTimestamp, &e.ReceivedAt); err != nil {
			return nil, err
		}
		out = append(out, e)
	}
	return out, rows.Err()
}
