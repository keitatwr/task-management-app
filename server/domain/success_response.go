package domain

type SuccessResponse struct {
	Message string `json:"message,omitempty"`
	Tasks   []Task `json:"tasks,omitempty"`
}
