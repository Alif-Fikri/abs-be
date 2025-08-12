package requests

type CreateTodoRequest struct {
	Tanggal   string `json:"tanggal" binding:"required"`
	Deskripsi string `json:"deskripsi" binding:"required"`
}

type UpdateTodoStatusRequest struct {
	IsDone bool `json:"is_done"`
}
