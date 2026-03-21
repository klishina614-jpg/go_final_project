package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/anaklisina/go_final_project/pkg/db"
)

func getTaskHandler(w http.ResponseWriter, r *http.Request) {
	id := r.FormValue("id")
	if id == "" {
		sendError(w, "Не указан идентификатор")
		return
	}

	task, err := db.GetTask(id)
	if err != nil {
		sendError(w, "Задача не найдена")
		return
	}

	resp, err := json.Marshal(task)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.Write(resp)
}

func editTaskHandler(w http.ResponseWriter, r *http.Request) {
	var task db.Task
	var buf bytes.Buffer

	_, err := buf.ReadFrom(r.Body)
	if err != nil {
		sendError(w, "Ошибка чтения запроса")
		return
	}

	if err = json.Unmarshal(buf.Bytes(), &task); err != nil {
		sendError(w, "Ошибка десериализации JSON")
		return
	}

	if task.ID == "" {
		sendError(w, "Не указан идентификатор задачи")
		return
	}

	if task.Title == "" {
		sendError(w, "Не указан заголовок задачи")
		return
	}

	now := time.Now()
	today := now.Format(DateFormat)

	if task.Date == "" {
		task.Date = today
	}

	t, err := time.Parse(DateFormat, task.Date)
	if err != nil {
		sendError(w, "Неверный формат даты")
		return
	}

	var next string
	if task.Repeat != "" {
		next, err = NextDate(now, task.Date, task.Repeat)
		if err != nil {
			sendError(w, err.Error())
			return
		}
	}

	nowDate := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	if t.Before(nowDate) {
		if task.Repeat == "" {
			task.Date = today
		} else {
			task.Date = next
		}
	}

	err = db.UpdateTask(&task)
	if err != nil {
		sendError(w, err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	fmt.Fprint(w, "{}")
}
