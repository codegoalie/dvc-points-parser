package models

type Resort struct {
	ID           int    `json:"id"`
	Name         string `json:"name"`
	Code         string `json:"code"`
	Abbreviation string `json:"abbreviation"`
}

type Room struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	ViewType    string `json:"view_type"`
	Code        string `json:"code"`
	Description string `json:"description"`
}
