package dto

import "time"

type ServiceOverviewPoint struct {
	Timestamp time.Time `json:"timestamp"`
	Events    uint64    `json:"events"`
	Issues    uint64    `json:"issues"`
	Users     uint64    `json:"users"`
}
