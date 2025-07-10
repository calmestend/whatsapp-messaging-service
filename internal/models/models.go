package models

// Struct with parameters for basic message sending
type MessageRequest struct {
	To    string `json:"to" validate:"required,number"`
	From  string `json:"from" validate:"required,number"`
	Token string `json:"token" validate:"required"`
}

// Struct with parameters for template creation
type TemplateRequest struct {
	PhoneID string `json:"phone_id" validate:"required,number"`
	AppID   string `json:"app_ID" validate:"required,number"`
	WbaID   string `json:"wba_id" validate:"required,number"`
	Token   string `json:"token" validate:"required"`
}
