package controller

import (
	"net/http"

	"app/internal/kernel"
)

type UserController struct {
	Kernel *kernel.Kernel
}

func NewUserController(k *kernel.Kernel) *UserController {
	return &UserController{Kernel: k}
}

func (c *UserController) Index(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement
	c.Kernel.Logger.Println("UserController.Index called")
	w.Write([]byte("User index"))
}

func (c *UserController) Show(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement
	c.Kernel.Logger.Println("UserController.Show called")
	w.Write([]byte("User show"))
}

func (c *UserController) Create(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement
	w.Write([]byte("User create"))
}

func (c *UserController) Update(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement
	w.Write([]byte("User update"))
}

func (c *UserController) Delete(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement
	w.Write([]byte("User delete"))
}
