package observability

import "context"

type Kit interface {
	Metrics() Recorder
	Run(ctx context.Context, name string, fn func(context.Context) error, attrs ...Attr) error
	RunDB(ctx context.Context, operation string, fn func(context.Context) error, attrs ...Attr) error
	RunS3(ctx context.Context, spanName, operation string, fn func(context.Context) error, attrs ...Attr) error
	RegisterDatabaseMetrics(stats DatabaseStats) (unregister func(), err error)
	SetAttributes(ctx context.Context, attrs ...Attr)
	AddEvent(ctx context.Context, name string, attrs ...Attr)
}

func RunResult[T any](
	kit Kit,
	ctx context.Context,
	name string,
	fn func(context.Context) (T, error),
	attrs ...Attr,
) (T, error) {
	var result T
	err := kit.Run(ctx, name, func(ctx context.Context) error {
		var runErr error
		result, runErr = fn(ctx)
		return runErr
	}, attrs...)
	return result, err
}
