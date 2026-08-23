package circuitbreaker_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"avatar/internal/circuitbreaker"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDoSuccessStaysClosed(t *testing.T) {
	breaker := circuitbreaker.New(circuitbreaker.Options{Name: "ok", FailureThreshold: 2})

	require.NoError(t, breaker.Do(func() error { return nil }))
	assert.Equal(t, circuitbreaker.StateClosed, breaker.State())
}

func TestDoOpensAfterFailureThreshold(t *testing.T) {
	breaker := circuitbreaker.New(circuitbreaker.Options{
		Name:             "fail",
		FailureThreshold: 2,
		Timeout:          time.Hour,
	})
	boom := errors.New("downstream down")

	require.ErrorIs(t, breaker.Do(func() error { return boom }), boom)
	assert.Equal(t, circuitbreaker.StateClosed, breaker.State())

	require.ErrorIs(t, breaker.Do(func() error { return boom }), boom)
	assert.Equal(t, circuitbreaker.StateOpen, breaker.State())

	err := breaker.Do(func() error {
		t.Fatal("must not call downstream while open")
		return nil
	})
	require.ErrorIs(t, err, circuitbreaker.ErrOpen)
	assert.Contains(t, err.Error(), "fail")
}

func TestDoIgnoresNonFailureErrors(t *testing.T) {
	ignored := errors.New("not found")
	breaker := circuitbreaker.New(circuitbreaker.Options{
		Name:             "ignore",
		FailureThreshold: 1,
		Timeout:          time.Hour,
		IsSuccessful: func(err error) bool {
			return err == nil || errors.Is(err, ignored)
		},
	})

	require.ErrorIs(t, breaker.Do(func() error { return ignored }), ignored)
	assert.Equal(t, circuitbreaker.StateClosed, breaker.State())
}

func TestDoIgnoresCallerCancellation(t *testing.T) {
	breaker := circuitbreaker.New(circuitbreaker.Options{
		Name:             "cancel",
		FailureThreshold: 1,
		Timeout:          time.Hour,
	})

	require.ErrorIs(t, breaker.Do(func() error { return context.Canceled }), context.Canceled)
	assert.Equal(t, circuitbreaker.StateClosed, breaker.State())

	require.ErrorIs(t, breaker.Do(func() error { return context.DeadlineExceeded }), context.DeadlineExceeded)
	assert.Equal(t, circuitbreaker.StateOpen, breaker.State())
}

type recorderStub struct {
	names  []string
	states []int64
}

func (stub *recorderStub) RecordCircuitBreakerState(_ context.Context, name string, state int64) {
	stub.names = append(stub.names, name)
	stub.states = append(stub.states, state)
}

func TestReportStateToEncodesStates(t *testing.T) {
	stub := &recorderStub{}
	report := circuitbreaker.ReportStateTo(stub)

	report("s3", circuitbreaker.StateClosed, circuitbreaker.StateOpen)
	report("s3", circuitbreaker.StateOpen, circuitbreaker.StateHalfOpen)
	report("s3", circuitbreaker.StateHalfOpen, circuitbreaker.StateClosed)

	assert.Equal(t, []string{"s3", "s3", "s3"}, stub.names)
	assert.Equal(t, []int64{2, 1, 0}, stub.states)
}

func TestDoHalfOpenRecovers(t *testing.T) {
	var transitions []circuitbreaker.State
	breaker := circuitbreaker.New(circuitbreaker.Options{
		Name:             "recover",
		FailureThreshold: 1,
		MaxRequests:      1,
		Timeout:          20 * time.Millisecond,
		OnStateChange: func(_ string, _, to circuitbreaker.State) {
			transitions = append(transitions, to)
		},
	})

	require.Error(t, breaker.Do(func() error { return errors.New("fail") }))
	assert.Equal(t, circuitbreaker.StateOpen, breaker.State())

	time.Sleep(30 * time.Millisecond)

	require.NoError(t, breaker.Do(func() error { return nil }))
	assert.Equal(t, circuitbreaker.StateClosed, breaker.State())
	assert.Equal(t, []circuitbreaker.State{
		circuitbreaker.StateOpen,
		circuitbreaker.StateHalfOpen,
		circuitbreaker.StateClosed,
	}, transitions)
}
