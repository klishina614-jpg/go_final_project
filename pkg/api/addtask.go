package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/anaklisina/go_final_project/pkg/db"
)

func addTaskHandler(w http.ResponseWriter, r *http.Request) {
	var task db.Task
	var buf bytes.Buffer

	_, err := buf.ReadFrom(r.Body)
	if err != nil {
		sendError(w, "Ошибка чтения запроса", http.StatusBadRequest)
		return
	}

	if err = json.Unmarshal(buf.Bytes(), &task); err != nil {
		log.Println("ошибка десериализации:", err)
		sendError(w, "Ошибка десериализации JSON", http.StatusBadRequest)
		return
	}

	if task.Title == "" {
		sendError(w, "Не указан заголовок задачи", http.StatusBadRequest)
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
		sendError(w, "Неверный формат даты", http.StatusBadRequest)
		return
	}

	// если задано правило повторения, проверяем его
	var next string
	if task.Repeat != "" {
		next, err = NextDate(now, task.Date, task.Repeat)
		if err != nil {
			sendError(w, err.Error(), http.StatusBadRequest)
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
		sendError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	resp, err := json.Marshal(map[string]interface{}{"id": fmt.Sprintf("%d", id)})
	if err != nil {
		sendError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.Write(resp)
}

// sendError формирует JSON-ответ с ошибкой и выставляет HTTP-код
func sendError(w http.ResponseWriter, msg string, status int) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(status)
	resp, err := json.Marshal(map[string]string{"error": msg})
	if err != nil {
		log.Println("ошибка сериализации JSON:", err)
		return
	}
	if _, err := w.Write(resp); err != nil {
		log.Println("ошибка записи ответа:", err)
	}
}
