package security

type User struct {
	ID       string
	Email    string
	Password string
	Roles    []string
}

func (u *User) GetID() string {
	return u.ID
}

func (u *User) GetEmail() string {
	return u.Email
}

func (u *User) GetPassword() string {
	return u.Password
}

func (u *User) GetRoles() []string {
	return u.Roles
}
