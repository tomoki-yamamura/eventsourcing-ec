package input

import "time"

type CreateTenantCartAbandonedPolicyInput struct {
	TenantID         string    `json:"tenant_id"`
	Title            string    `json:"title"`
	AbandonedMinutes int       `json:"abandoned_minutes"`
	QuietTimeFrom    time.Time `json:"quiet_time_from"`
	QuietTimeTo      time.Time `json:"quiet_time_to"`
}