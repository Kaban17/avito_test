package entity

import "time"

type Team struct {
	Name      string
	Members   []User
	CreatedAt time.Time
}
