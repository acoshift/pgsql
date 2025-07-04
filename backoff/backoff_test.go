package backoff_test

import (
	"testing"
	"time"

	"github.com/acoshift/pgsql/backoff"
)

func TestExponential(t *testing.T) {
	t.Parallel()

	config := backoff.ExponentialConfig{
		Config: backoff.Config{
			BaseDelay: 10 * time.Millisecond,
			MaxDelay:  1 * time.Second,
		},
		Multiplier: 2.0,
	}
	backoff := backoff.NewExponential(config)

	// Test exponential growth
	delays := []time.Duration{}
	for i := 0; i < 10; i++ {
		delay := backoff(i)
		delays = append(delays, delay)
	}

	// Verify exponential growth
	for i := 1; i < len(delays); i++ {
		if delays[i] < delays[i-1] {
			t.Errorf("Expected delay[%d] >= delay[%d], got %v < %v", i, i-1, delays[i], delays[i-1])
		}
	}

	// Verify max delay
	for i := 0; i < 10; i++ {
		delay := backoff(i)
		if delay > config.MaxDelay {
			t.Errorf("Expected delay[%d] <= MaxDelay (%v), got %v", i, config.MaxDelay, delay)
		}
	}
}

func TestExponentialWithFullJitter(t *testing.T) {
	t.Parallel()

	config := backoff.ExponentialConfig{
		Config: backoff.Config{
			BaseDelay: 100 * time.Millisecond,
			MaxDelay:  1 * time.Second,
		},
		Multiplier: 2.0,
		JitterType: backoff.FullJitter,
	}
	backoff := backoff.NewExponential(config)

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

	// Verify max delay
	for i := 0; i < 15; i++ {
		delay := backoff(i)
		if delay > config.MaxDelay {
			t.Errorf("Expected delay[%d] <= MaxDelay (%v), got %v", i, config.MaxDelay, delay)
		}
	}
}

func TestExponentialWithEqualJitter(t *testing.T) {
	t.Parallel()

	config := backoff.ExponentialConfig{
		Config: backoff.Config{
			BaseDelay: 100 * time.Millisecond,
			MaxDelay:  1 * time.Second,
		},
		Multiplier: 2.0,
		JitterType: backoff.EqualJitter,
	}
	backoff := backoff.NewExponential(config)

	delay := backoff(2)

	// With equal jitter, delay should be at least half of the calculated delay
	expectedMin := 200 * time.Millisecond // (100ms * 2^2) / 2 = 200ms
	if delay < expectedMin {
		t.Errorf("Expected delay >= %v with equal jitter, got %v", expectedMin, delay)
	}

	// Verify max delay
	for i := 0; i < 15; i++ {
		delay := backoff(i)
		if delay > config.MaxDelay {
			t.Errorf("Expected delay[%d] <= MaxDelay (%v), got %v", i, config.MaxDelay, delay)
		}
	}
}

func TestLinearBackoff(t *testing.T) {
	t.Parallel()

	config := backoff.LinearConfig{
		Config: backoff.Config{
			BaseDelay: 100 * time.Millisecond,
			MaxDelay:  1 * time.Second,
		},
		Increment: 100 * time.Millisecond,
	}
	backoff := backoff.NewLinear(config)

	// Test linear growth
	delays := []time.Duration{}
	for i := 0; i < 5; i++ {
		delay := backoff(i)
		delays = append(delays, delay)
	}

	// Verify linear growth
	for i := 1; i < len(delays); i++ {
		expectedIncrease := 100 * time.Millisecond
		actualIncrease := delays[i] - delays[i-1]

		if actualIncrease != expectedIncrease {
			t.Errorf("Expected linear increase of %v, got %v", expectedIncrease, actualIncrease)
		}
	}

	// Verify max delay
	for i := 0; i < 15; i++ {
		delay := backoff(i)
		if delay > config.MaxDelay {
			t.Errorf("Expected delay[%d] <= MaxDelay (%v), got %v", i, config.MaxDelay, delay)
		}
	}
}
