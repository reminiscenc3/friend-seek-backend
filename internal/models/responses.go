package models

type Response struct {
	Msg  string `json:"msg,omitempty"`
	Data any    `json:"data,omitempty"`
}

type HealthResponse struct {
	Status   string `json:"status"`
	Database string `json:"database"`
}
