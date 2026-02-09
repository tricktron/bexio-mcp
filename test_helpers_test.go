package main

type fakeBexioCapturedRequest struct {
	Method        string
	Path          string
	Authorization string
	Body          bexioCreateTimesheetRequest
}

func intPtr(v int) *int {
	return &v
}
