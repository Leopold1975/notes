package models

import "time"

type Note struct {
	ID          uint64        `json:"id"`
	Title       string        `json:"title"`
	Description string        `json:"description"`
	DateAdded   time.Time     `json:"dateAdded"`
	DateNotify  time.Time     `json:"dateNotify"`
	Delay       time.Duration `json:"delay"`
}
