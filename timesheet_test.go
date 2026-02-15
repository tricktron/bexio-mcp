package main

import (
	"encoding/json"
	"testing"

	"github.com/alecthomas/assert/v2"
)

func TestTimesheetResultTypesMarshalAsWrappedObjects(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name string
		in   any
		want string
	}{
		{
			name: "delete result marshals success field",
			in:   deleteTimesheetResult{Success: true},
			want: `{"success":true}`,
		},
		{
			name: "search result marshals results wrapper",
			in: searchTimesheetsResult{
				Results: []bexioTimesheet{{ID: 1}},
			},
			want: `{"results":[{"id":1,"user_id":0,"status_id":0,"allowable_bill":false,"client_service_id":0,"date":"","duration":"","running":false,"tracking":{"type":""}}]}`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got, err := json.Marshal(tc.in)
			assert.NoError(t, err)
			assert.Equal(t, tc.want, string(got))
		})
	}
}
