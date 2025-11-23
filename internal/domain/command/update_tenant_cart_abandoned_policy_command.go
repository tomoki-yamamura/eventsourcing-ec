package command

import (
	"time"

	"github.com/google/uuid"
)

type UpdateTenantCartAbandonedPolicyCommand struct {
	TenantID         uuid.UUID
	Title            string
	AbandonedMinutes int
	QuietTimeFrom    time.Time
	QuietTimeTo      time.Time
}