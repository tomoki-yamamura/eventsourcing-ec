package command

import (
	"time"

	"github.com/google/uuid"
)

type CreateTenantCartAbandonedPolicyCommand struct {
	TenantID         uuid.UUID
	Title            string
	AbandonedMinutes int
	QuietTimeFrom    time.Time
	QuietTimeTo      time.Time
}
