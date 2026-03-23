package api

import (
	"fmt"
	"net/http"
	"time"

	"github.com/anaklisina/go_final_project/pkg/db"
)

func doneTaskHandler(w http.ResponseWriter, r *http.Request) {
	id := r.FormValue("id")
	if id == "" {
		sendError(w, "Не указан идентификатор", http.StatusBadRequest)
		return
	}

	task, err := db.GetTask(id)
	if err != nil {
		sendError(w, "Задача не найдена", http.StatusNotFound)
		return
	}

	if task.Repeat == "" {
		// одноразовая задача — удаляем
		err = db.DeleteTask(id)
		if err != nil {
			sendError(w, err.Error(), http.StatusInternalServerError)
			return
		}
	} else {
		// периодическая задача — вычисляем следующую дату
		now := time.Now()
		next, err := NextDate(now, task.Date, task.Repeat)
		if err != nil {
			sendError(w, err.Error(), http.StatusInternalServerError)
			return
		}
		err = db.UpdateDate(id, next)
		if err != nil {
			sendError(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	fmt.Fprint(w, "{}")
}

func deleteTaskHandler(w http.ResponseWriter, r *http.Request) {
	id := r.FormValue("id")
	if id == "" {
		sendError(w, "Не указан идентификатор", http.StatusBadRequest)
		return
	}

	err := db.DeleteTask(id)
	if err != nil {
		sendError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	fmt.Fprint(w, "{}")
}
