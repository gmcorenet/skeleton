package service

import (
	"app/internal/kernel"
)

type UserService struct {
	Kernel *kernel.Kernel
}

func NewUserService(k *kernel.Kernel) *UserService {
	return &UserService{Kernel: k}
}

func (s *UserService) FindAll() []string {
	// TODO: Implement
	return []string{}
}

func (s *UserService) FindByID(id string) *User {
	// TODO: Implement
	return nil
}

func (s *UserService) Create(data CreateUserData) (*User, error) {
	// TODO: Implement
	return nil, nil
}

func (s *UserService) Update(id string, data UpdateUserData) (*User, error) {
	// TODO: Implement
	return nil, nil
}

func (s *UserService) Delete(id string) error {
	// TODO: Implement
	return nil
}

type CreateUserData struct {
	Name  string
	Email string
}

type UpdateUserData struct {
	Name  string
	Email string
}

type User struct {
	ID    string
	Name  string
	Email string
}
