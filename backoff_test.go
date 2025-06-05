package pgsql_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/acoshift/pgsql"
)

func TestExponentialBackoff(t *testing.T) {
	t.Parallel()

	config := pgsql.ExponentialBackoffConfig{
		BackoffConfig: pgsql.BackoffConfig{
			BaseDelay: 10 * time.Millisecond,
			MaxDelay:  1 * time.Second,
		},
		Multiplier: 2.0,
	}
	backoff := pgsql.NewExponentialBackoff(config)

	// Test exponential growth
	delays := []time.Duration{}
	for i := 0; i < 5; i++ {
		delay := backoff(i)
		delays = append(delays, delay)
	}

	// Verify exponential growth
	for i := 1; i < len(delays); i++ {
		if delays[i] < delays[i-1] {
			t.Errorf("Expected delay[%d] >= delay[%d], got %v < %v", i, i-1, delays[i], delays[i-1])
		}
	}
}

func TestExponentialBackoffWithFullJitter(t *testing.T) {
	t.Parallel()

	config := pgsql.ExponentialBackoffWithJitterConfig{
		ExponentialBackoffConfig: pgsql.ExponentialBackoffConfig{
			BackoffConfig: pgsql.BackoffConfig{
				BaseDelay: 100 * time.Millisecond,
				MaxDelay:  1 * time.Second,
			},
			Multiplier: 2.0,
		},
		JitterType: pgsql.FullJitter,
	}
	backoff := pgsql.NewExponentialBackoffWithJitter(config)

	// Test that jitter introduces randomness
	var delays []time.Duration
	for i := 0; i < 10; i++ {
		delay := backoff(3) // Use same attempt number
		delays = append(delays, delay)
	}

	// Check that not all delays are the same (indicating jitter is working)
	allSame := true
	for i := 1; i < len(delays); i++ {
		if delays[i] != delays[0] {
			allSame = false
			break
		}
	}
	if allSame {
		t.Error("Expected jitter to produce different delays, but all delays were the same")
	}
}

func TestExponentialBackoffWithEqualJitter(t *testing.T) {
	t.Parallel()

	config := pgsql.ExponentialBackoffWithJitterConfig{
		ExponentialBackoffConfig: pgsql.ExponentialBackoffConfig{
			BackoffConfig: pgsql.BackoffConfig{
				BaseDelay: 100 * time.Millisecond,
				MaxDelay:  1 * time.Second,
			},
			Multiplier: 2.0,
		},
		JitterType: pgsql.EqualJitter,
	}
	backoff := pgsql.NewExponentialBackoffWithJitter(config)

	delay := backoff(2)

	// With equal jitter, delay should be at least half of the calculated delay
	expectedMin := 200 * time.Millisecond // (100ms * 2^2) / 2 = 200ms
	if delay < expectedMin {
		t.Errorf("Expected delay >= %v with equal jitter, got %v", expectedMin, delay)
	}
}

func TestLinearBackoff(t *testing.T) {
	t.Parallel()

	config := pgsql.LinearBackoffConfig{
		BackoffConfig: pgsql.BackoffConfig{
			BaseDelay: 50 * time.Millisecond,
			MaxDelay:  500 * time.Millisecond,
		},
		Increment: 50 * time.Millisecond,
	}
	backoff := pgsql.NewLinearBackoff(config)

	// Test linear growth
	delays := []time.Duration{}
	for i := 0; i < 5; i++ {
		delay := backoff(i)
		delays = append(delays, delay)
	}

	// Verify linear growth
	for i := 1; i < len(delays); i++ {
		expectedIncrease := 50 * time.Millisecond
		actualIncrease := delays[i] - delays[i-1]

		if actualIncrease != expectedIncrease {
			t.Errorf("Expected linear increase of %v, got %v", expectedIncrease, actualIncrease)
		}
	}
}

func TestDefaultBackoffFunctions(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name    string
		backoff pgsql.BackoffDelayFunc
	}{
		{"DefaultExponentialBackoff", pgsql.DefaultExponentialBackoff()},
		{"DefaultExponentialBackoffWithFullJitter", pgsql.DefaultExponentialBackoffWithFullJitter()},
		{"DefaultExponentialBackoffWithEqualJitter", pgsql.DefaultExponentialBackoffWithEqualJitter()},
		{"DefaultLinearBackoff", pgsql.DefaultLinearBackoff()},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			delay := tc.backoff(0)

			// Should have some delay (base delay should be at least 50ms for all defaults)
			if delay < 50*time.Millisecond {
				t.Errorf("Expected some delay, got %v", delay)
			}
		})
	}
}

func TestMaxDelayIsRespected(t *testing.T) {
	t.Parallel()

	config := pgsql.ExponentialBackoffConfig{
		BackoffConfig: pgsql.BackoffConfig{
			BaseDelay: 100 * time.Millisecond,
			MaxDelay:  200 * time.Millisecond, // Very low max delay
		},
		Multiplier: 2.0,
	}
	backoff := pgsql.NewExponentialBackoff(config)

	// Test that max delay is respected even with high attempt numbers
	delay := backoff(10) // This would normally result in a very long delay

	maxExpected := 200 * time.Millisecond

	if delay > maxExpected {
		t.Errorf("Expected delay capped at %v, got %v", maxExpected, delay)
	}
}

// Example demonstrating usage with transaction retry
func ExampleNewExponentialBackoff() {
	// Create a custom exponential backoff
	backoff := pgsql.NewExponentialBackoff(pgsql.ExponentialBackoffConfig{
		BackoffConfig: pgsql.BackoffConfig{
			BaseDelay: 100 * time.Millisecond,
			MaxDelay:  5 * time.Second,
		},
		Multiplier: 2.0,
	})

	// Use with transaction options
	opts := &pgsql.TxOptions{
		MaxAttempts:      5,
		BackoffDelayFunc: backoff,
	}

	fmt.Printf("Transaction options configured with custom backoff (MaxAttempts: %d)\n", opts.MaxAttempts)
	// Output: Transaction options configured with custom backoff (MaxAttempts: 5)
}

func ExampleDefaultExponentialBackoffWithFullJitter() {
	// Use a pre-configured exponential backoff with full jitter
	opts := &pgsql.TxOptions{
		MaxAttempts:      3,
		BackoffDelayFunc: pgsql.DefaultExponentialBackoffWithFullJitter(),
	}

	fmt.Printf("Transaction options with full jitter backoff (MaxAttempts: %d)\n", opts.MaxAttempts)
	// Output: Transaction options with full jitter backoff (MaxAttempts: 3)
}
