package circuitbreaker

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/sony/gobreaker/v2"
)

const (
	defaultFailureThreshold uint32 = 5
	defaultMaxRequests      uint32 = 2
	defaultTimeout                 = 10 * time.Second
)

var ErrOpen = errors.New("circuit breaker is open")

type State string

const (
	StateClosed   State = "closed"
	StateHalfOpen State = "half-open"
	StateOpen     State = "open"
)

func (state State) code() int64 {
	switch state {
	case StateOpen:
		return 2
	case StateHalfOpen:
		return 1
	default:
		return 0
	}
}

type StateRecorder interface {
	RecordCircuitBreakerState(ctx context.Context, name string, state int64)
}

func ReportStateTo(recorder StateRecorder) func(name string, from, to State) {
	return func(name string, _, to State) {
		recorder.RecordCircuitBreakerState(context.Background(), name, to.code())
	}
}

type Options struct {
	Name             string
	FailureThreshold uint32
	MaxRequests      uint32
	Timeout          time.Duration
	IsSuccessful     func(error) bool
	OnStateChange    func(name string, from, to State)
}

type Breaker struct {
	name  string
	inner *gobreaker.CircuitBreaker[struct{}]
}

func New(options Options) *Breaker {
	if options.FailureThreshold == 0 {
		options.FailureThreshold = defaultFailureThreshold
	}
	if options.MaxRequests == 0 {
		options.MaxRequests = defaultMaxRequests
	}
	if options.Timeout <= 0 {
		options.Timeout = defaultTimeout
	}

	settings := gobreaker.Settings{
		Name:        options.Name,
		MaxRequests: options.MaxRequests,
		Timeout:     options.Timeout,
		ReadyToTrip: func(counts gobreaker.Counts) bool {
			return counts.ConsecutiveFailures >= options.FailureThreshold
		},
		IsSuccessful: func(err error) bool {
			return isSuccess(err, options.IsSuccessful)
		},
	}
	if options.OnStateChange != nil {
		settings.OnStateChange = func(name string, from, to gobreaker.State) {
			options.OnStateChange(name, toState(from), toState(to))
		}
	}

	return &Breaker{name: options.Name, inner: gobreaker.NewCircuitBreaker[struct{}](settings)}
}

func (breaker *Breaker) Do(fn func() error) error {
	_, err := breaker.inner.Execute(func() (struct{}, error) {
		return struct{}{}, fn()
	})
	if errors.Is(err, gobreaker.ErrOpenState) || errors.Is(err, gobreaker.ErrTooManyRequests) {
		return fmt.Errorf("%w: %s", ErrOpen, breaker.name)
	}
	return err
}

func (breaker *Breaker) State() State {
	return toState(breaker.inner.State())
}

func isSuccess(err error, custom func(error) bool) bool {
	if err == nil || errors.Is(err, context.Canceled) {
		return true
	}
	return custom != nil && custom(err)
}

func toState(state gobreaker.State) State {
	switch state {
	case gobreaker.StateOpen:
		return StateOpen
	case gobreaker.StateHalfOpen:
		return StateHalfOpen
	default:
		return StateClosed
	}
}
