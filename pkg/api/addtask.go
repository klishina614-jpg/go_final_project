package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/anaklisina/go_final_project/pkg/db"
)

func addTaskHandler(w http.ResponseWriter, r *http.Request) {
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

	if task.Title == "" {
		sendError(w, "Не указан заголовок задачи")
		return
	}

	now := time.Now()
	today := now.Format(DateFormat)

	if task.Date == "" {
		task.Date = today
	}

	// проверяем что дата корректная
	t, err := time.Parse(DateFormat, task.Date)
	if err != nil {
		sendError(w, "Неверный формат даты")
		return
	}

	// если задано правило повторения, проверяем его
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

	id, err := db.AddTask(&task)
	if err != nil {
		sendError(w, err.Error())
		return
	}

	resp, err := json.Marshal(map[string]interface{}{"id": fmt.Sprintf("%d", id)})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.Write(resp)
}

// sendError формирует JSON-ответ с ошибкой
func sendError(w http.ResponseWriter, msg string) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	resp, _ := json.Marshal(map[string]string{"error": msg})
	w.Write(resp)
}
