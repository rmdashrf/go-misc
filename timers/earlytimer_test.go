package timers

import (
	"sync"
	"testing"
	"time"
)

func convertToMsOffsets(start time.Time, timestamps []time.Time) (offsets []int) {
	for _, t := range timestamps {
		offsets = append(offsets, int(t.Sub(start).Seconds()*1000.0))
	}
	return
}

const (
	// 5ms of testing slack
	SLACK = 5
)

func TestEarlyTimer(t *testing.T) {
	var (
		timestamps   []time.Time
		timestampsMu sync.Mutex
	)

	task := func() {
		timestampsMu.Lock()
		timestamps = append(timestamps, time.Now())
		timestampsMu.Unlock()
	}

	startTime := time.Now()
	earlyTimer := NewEarlyPeriodicTimer(100*time.Millisecond, task, RunOnStart())
	time.Sleep(150 * time.Millisecond)
	earlyTimer.RunNow()
	time.Sleep(200 * time.Millisecond)
	earlyTimer.Stop()

	// Make sure that there isn't another activation that is recorded within the next 200ms
	time.Sleep(200 * time.Millisecond)

	offsets := convertToMsOffsets(startTime, timestamps)

	expectedOffsets := []int{0, 100, 150, 250}
	if len(expectedOffsets) != len(offsets) {
		t.Fatal("Expected:", expectedOffsets, "\nGot", offsets)
	}

	for i, _ := range expectedOffsets {
		if abs(expectedOffsets[i]-offsets[i]) >= SLACK {
			t.Fatal("Expected:", expectedOffsets, "\nGot", offsets)
		}
	}
}

func abs(x int) int {
	if x < 0 {
		return -x
	} else {
		return x
	}
}
