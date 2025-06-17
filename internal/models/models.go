package models

// Struct with parameters for basic message sending
type MessageRequest struct {
	To   string `json:"to" validate:"required,number"`
	From string `json:"from" validate:"required,number"`
}
