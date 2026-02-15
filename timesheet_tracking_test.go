package main

import (
	"encoding/json"
	"testing"

	"github.com/alecthomas/assert/v2"
)

func TestTrackingRangeMarshalDurationOmitsRangeFields(t *testing.T) {
	t.Parallel()

	tracking := trackingRange{
		Type:     "duration",
		Date:     "2026-02-08",
		Duration: "01:30",
	}

	raw, err := json.Marshal(tracking)
	assert.NoError(t, err)

	var payload map[string]any
	err = json.Unmarshal(raw, &payload)
	assert.NoError(t, err)

	duration, hasDuration := payload["duration"]
	assert.True(t, hasDuration, "duration tracking payload should include duration")
	assert.Equal(t, "01:30", duration)

	_, hasStart := payload["start"]
	_, hasEnd := payload["end"]
	assert.False(t, hasStart, "duration tracking payload should not include start")
	assert.False(t, hasEnd, "duration tracking payload should not include end")
}
