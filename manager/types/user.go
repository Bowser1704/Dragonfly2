package types

type LoginRequest struct {
	Name     string `form:"name" binding:"required,min=3,max=10"`
	Password string `form:"password" binding:"required,min=12,max=20"`
}

type RegisterRequest struct {
	LoginRequest
	Email    string `form:"email" binding:"required,email"`
	Phone    string `form:"phone" binding:"omitempty"`
	Avatar   string `form:"avatar" binding:"omitempty"`
	Location string `form:"location" binding:"omitempty"`
	BIO      string `form:"bio" binding:"omitempty"`
}
