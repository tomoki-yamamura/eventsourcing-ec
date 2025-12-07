package dto

import (
	"time"
)

type TenantPolicyViewDTO struct {
	ID               string    `json:"id"`
	Title            string    `json:"title"`
	AbandonedMinutes int       `json:"abandoned_minutes"`
	QuietTimeFrom    time.Time `json:"quiet_time_from"`
	QuietTimeTo      time.Time `json:"quiet_time_to"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
	Version          int       `json:"version"`
}
