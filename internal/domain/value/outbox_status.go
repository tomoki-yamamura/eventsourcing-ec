package value

type OutboxStatus string

const (
	OutboxStatusPending    OutboxStatus = "PENDING"
	OutboxStatusProcessing OutboxStatus = "PROCESSING"
	OutboxStatusPublished  OutboxStatus = "PUBLISHED"
	OutboxStatusFailed     OutboxStatus = "FAILED"
)

func (s OutboxStatus) String() string {
	return string(s)
}

func (s OutboxStatus) IsPending() bool {
	return s == OutboxStatusPending
}

func (s OutboxStatus) IsPublished() bool {
	return s == OutboxStatusPublished
}

func (s OutboxStatus) IsFailed() bool {
	return s == OutboxStatusFailed
}

func (s OutboxStatus) IsProcessing() bool {
	return s == OutboxStatusProcessing
}
