package main

import (
	"encoding/json"
	"testing"

	"github.com/alecthomas/assert/v2"
)

func TestTrackingRangeMarshalJSONDurationOmitsRangeFields(t *testing.T) {
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

func TestTrackingRangeMarshalJSONRangeOmitsDurationField(t *testing.T) {
	t.Parallel()

	tracking := trackingRange{
		Type:  "range",
		Date:  "2026-02-08",
		Start: "09:00",
		End:   "10:30",
	}

	raw, err := json.Marshal(tracking)
	assert.NoError(t, err)

	var payload map[string]any
	err = json.Unmarshal(raw, &payload)
	assert.NoError(t, err)

	start, hasStart := payload["start"]
	end, hasEnd := payload["end"]
	assert.True(t, hasStart, "range tracking payload should include start")
	assert.True(t, hasEnd, "range tracking payload should include end")
	assert.Equal(t, "09:00", start)
	assert.Equal(t, "10:30", end)

	_, hasDuration := payload["duration"]
	assert.False(t, hasDuration, "range tracking payload should not include duration")
}
