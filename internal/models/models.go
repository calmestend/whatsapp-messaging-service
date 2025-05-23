package models

type MessageRequest struct {
	To   string `json:"to" validate:"required,number"`
	From string `json:"from" validate:"required,number"`
}
