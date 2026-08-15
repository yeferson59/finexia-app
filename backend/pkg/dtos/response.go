package dtos

import "time"

type Response struct {
	Success   bool      `json:"success"`
	Message   string    `json:"message"`
	Details   string    `json:"details"`
	Action    string    `json:"action,omitempty"`
	Data      any       `json:"data,omitempty"`
	Timestamp time.Time `json:"timestamp"`
}
