package requests

type AdminRequest struct {
	Nama     string `json:"nama" binding:"required"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}
