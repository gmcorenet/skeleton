package repository

import (
	"app/internal/kernel"
)

type UserRepository struct {
	Kernel *kernel.Kernel
}

func NewUserRepository(k *kernel.Kernel) *UserRepository {
	return &UserRepository{Kernel: k}
}

func (r *UserRepository) FindAll() ([]*User, error) {
	// TODO: Implement
	return []*User{}, nil
}

func (r *UserRepository) FindByID(id string) (*User, error) {
	// TODO: Implement
	return nil, nil
}

func (r *UserRepository) Save(user *User) error {
	// TODO: Implement
	return nil
}

func (r *UserRepository) Delete(id string) error {
	// TODO: Implement
	return nil
}

type User struct {
	ID    string
	Name  string
	Email string
}
