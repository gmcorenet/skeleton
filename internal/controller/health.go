package controller

import (
	"encoding/json"
	"net/http"
	"runtime"
	"time"
)

type HealthController struct{}

// #[Route(path="/health", methods={"GET"}, name="health", public=true)]
func (c *HealthController) Check(w http.ResponseWriter, r *http.Request, params map[string]string) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":    "ok",
		"timestamp": time.Now().UTC().Format(time.RFC3339),
		"go":        runtime.Version(),
	})
}
