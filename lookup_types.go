package main

type bexioClientService struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type bexioContact struct {
	ID    int    `json:"id"`
	Name1 string `json:"name_1"`
	Name2 string `json:"name_2"`
}
