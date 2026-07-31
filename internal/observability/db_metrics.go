package observability

import (
	"context"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

const databaseStatsQueryTimeout = 5 * time.Second

// DatabaseStats reads current database and pool state for gauge export.
type DatabaseStats struct {
	PoolConnections func() (active, idle int64)
	CountAvatars    func(ctx context.Context) (total int64, byUser map[string]int64, err error)
}

// RegisterDatabaseMetrics registers OTel observable gauges for pool and avatar counts.
// The returned function unregisters callbacks and should run before closing the pool.
func RegisterDatabaseMetrics(stats DatabaseStats) (unregister func(), err error) {
	if !metricsEnabled || stats.PoolConnections == nil || stats.CountAvatars == nil {
		return func() {}, nil
	}

	meter := otel.Meter(TracerName)

	activeConnections, err := meter.Float64ObservableGauge("db_connections_active",
		metric.WithDescription("Number of active database connections"))
	if err != nil {
		return nil, err
	}

	idleConnections, err := meter.Float64ObservableGauge("db_connections_idle",
		metric.WithDescription("Number of idle database connections"))
	if err != nil {
		return nil, err
	}

	activeAvatarsTotal, err := meter.Float64ObservableGauge("avatars_active_total",
		metric.WithDescription("Number of active avatars"))
	if err != nil {
		return nil, err
	}

	activeAvatarsByUser, err := meter.Float64ObservableGauge("avatars_active_by_user",
		metric.WithDescription("Number of active avatars per user"))
	if err != nil {
		return nil, err
	}

	registration, err := meter.RegisterCallback(
		func(ctx context.Context, observer metric.Observer) error {
			active, idle := stats.PoolConnections()
			observer.ObserveFloat64(activeConnections, float64(active))
			observer.ObserveFloat64(idleConnections, float64(idle))

			queryCtx, cancel := context.WithTimeout(ctx, databaseStatsQueryTimeout)
			defer cancel()

			total, byUser, err := stats.CountAvatars(queryCtx)
			if err != nil {
				return nil
			}

			observer.ObserveFloat64(activeAvatarsTotal, float64(total))
			for userID, count := range byUser {
				observer.ObserveFloat64(
					activeAvatarsByUser,
					float64(count),
					metric.WithAttributes(attribute.String("user_id", userID)),
				)
			}
			return nil
		},
		activeConnections,
		idleConnections,
		activeAvatarsTotal,
		activeAvatarsByUser,
	)
	if err != nil {
		return nil, err
	}

	return func() { _ = registration.Unregister() }, nil
}
