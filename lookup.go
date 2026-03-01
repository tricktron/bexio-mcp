package main

type bexioContact struct {
	ID    int    `json:"id"`
	Name1 string `json:"name_1"`
	Name2 string `json:"name_2"`
}

type bexioProject struct {
	ID        int    `json:"id"`
	Name      string `json:"name"`
	ContactID int    `json:"contact_id"`
}

type bexioService struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type bexioPackage struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type bexioUser struct {
	ID        int    `json:"id"`
	FirstName string `json:"firstname"`
	LastName  string `json:"lastname"`
	Email     string `json:"email"`
}

type listContactsResult struct {
	Contacts []bexioContact `json:"contacts"`
}

type listProjectsResult struct {
	Projects []bexioProject `json:"projects"`
}

type listServicesResult struct {
	Services []bexioService `json:"services"`
}

type listPackagesResult struct {
	Packages []bexioPackage `json:"packages"`
}

type listStatusesResult struct {
	Statuses []timesheetStatus `json:"statuses"`
}
