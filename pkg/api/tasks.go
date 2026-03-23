package api

import (
	"encoding/json"
	"net/http"

	"github.com/anaklisina/go_final_project/pkg/db"
)

func tasksHandler(w http.ResponseWriter, r *http.Request) {
	search := r.FormValue("search")

	tasks, err := db.Tasks(search, 50)
	if err != nil {
		sendError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	resp, err := json.Marshal(map[string]interface{}{"tasks": tasks})
	if err != nil {
		sendError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.Write(resp)
}
