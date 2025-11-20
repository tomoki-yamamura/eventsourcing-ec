package command

import "github.com/google/uuid"

type SubmitCartCommand struct {
	CartID uuid.UUID
}
