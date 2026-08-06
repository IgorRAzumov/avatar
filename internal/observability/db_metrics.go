package observability

import (
	"context"
	"time"

	"go.opentelemetry.io/otel/metric"
)

const databaseStatsQueryTimeout = 5 * time.Second

type DatabaseStats struct {
	PoolConnections func() (active, idle int64)
	CountAvatars    func(ctx context.Context) (total int64, err error)
}

func (metrics *Metrics) RegisterDatabaseMetrics(stats DatabaseStats) (unregister func(), err error) {
	if metrics == nil || metrics.meter == nil || stats.PoolConnections == nil || stats.CountAvatars == nil {
		return func() {}, nil
	}

	activeConnections, err := metrics.meter.Float64ObservableGauge("db_connections_active",
		metric.WithDescription("Number of active database connections"))
	if err != nil {
		return nil, err
	}

	idleConnections, err := metrics.meter.Float64ObservableGauge("db_connections_idle",
		metric.WithDescription("Number of idle database connections"))
	if err != nil {
		return nil, err
	}

	activeAvatarsTotal, err := metrics.meter.Float64ObservableGauge("avatars_active_total",
		metric.WithDescription("Number of active avatars"))
	if err != nil {
		return nil, err
	}

	registration, err := metrics.meter.RegisterCallback(
		func(ctx context.Context, observer metric.Observer) error {
			active, idle := stats.PoolConnections()
			observer.ObserveFloat64(activeConnections, float64(active))
			observer.ObserveFloat64(idleConnections, float64(idle))

			queryCtx, cancel := context.WithTimeout(ctx, databaseStatsQueryTimeout)
			defer cancel()

			total, err := stats.CountAvatars(queryCtx)
			if err != nil {
				return err
			}

			observer.ObserveFloat64(activeAvatarsTotal, float64(total))
			return nil
		},
		activeConnections,
		idleConnections,
		activeAvatarsTotal,
	)
	if err != nil {
		return nil, err
	}

	return func() { _ = registration.Unregister() }, nil
}
