package models

type UploadTaskPayload struct {
	TaskID   string `json:"task_id"`
	TaskType string `json:"task_type"` // 任务类型
	Prompt   string `json:"prompt"`
	AuthorID string `json:"author_id"`
	Image    string `json:"image"` // Base64 编码的图像数据
}

type ProcessTaskPayload struct {
	TaskID   string `json:"task_id"`
	Prompt   string `json:"prompt"`
	ImageURL string `json:"image_url"`
}

type TaskResult struct {
	TaskID string
	Status string
	Error  error
}
