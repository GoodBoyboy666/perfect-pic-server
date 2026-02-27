package dto

type UpdateSettingRequest struct {
	Key   string `json:"key" binding:"required"`
	Value string `json:"value"`
}

type SendTestEmailRequest struct {
	ToEmail string `json:"to_email" binding:"required,email"`
}

type WebInfoResponse struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}
