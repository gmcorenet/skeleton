package form

type CreateUserForm struct {
	Name  string `form:"name"`
	Email string `form:"email"`
}

type UpdateUserForm struct {
	Name  string `form:"name"`
	Email string `form:"email"`
}
