package domain

type TaskCreateRequest struct {
	Title       string   `json:"title" binding:"required"`
	Description string   `json:"description" binding:"required"`
	DueDate     DateOnly `json:"dueDate" binding:"required"`
}

type TaskUpdateRequest struct {
	ID          int      `uri:"taskID"`
	Title       string   `json:"title"`
	Description string   `json:"description"`
	DueDate     DateOnly `json:"dueDate"`
}

type TaskFetchRequest struct {
	ID int `uri:"taskID"`
}
