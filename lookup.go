package main

type bexioListProjectsRequest struct {
	ContactID *int `json:"contact_id"`
}

type bexioListProjectPackagesRequest struct {
	ProjectID int `json:"project_id"`
}
