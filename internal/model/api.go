package model

import "github.com/google/uuid"

const (
	defaultPageSize = 10
	defaultPage     = 1
)

type SuccessResponse struct {
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
}

type ErrorResponse struct {
	Message string `json:"message"`
	Detail  any    `json:"detail,omitempty"`
}

type PaginationQuery struct {
	Limit  int `form:"limit,default=100" binding:"min=1,max=100"`
	Offset int `form:"offset,default=0" binding:"min=0"`
}

// params with only ID
type IDParam struct {
	ID string `uri:"id" binding:"required,uuid"`
}

// GetID returns a uuid representation of the ID param string
func (p *IDParam) GetID() uuid.UUID {
	id, _ := uuid.Parse(p.ID)
	return id
}
