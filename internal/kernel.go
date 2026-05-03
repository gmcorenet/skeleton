package kernel

import (
	"log"
	"os"
)

type Kernel struct {
	// TODO: Add framework components
	Logger *log.Logger
}

func New() *Kernel {
	return &Kernel{
		Logger: log.New(os.Stdout, "[app] ", log.LstdFlags),
	}
}
