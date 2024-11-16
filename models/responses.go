package models

type Status struct {
	Status bool `json:"status"`
}

type Error struct {
	Detail string `json:"detail"`
}
