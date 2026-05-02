package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/EugeneDFox/go_final_project/pkg/db"
)

// checkDate validates and normalizes the date of a task.
//   - If task.Date is empty, it is set to the current date (today).
//   - If task.Date is provided, it must be in "20060102" (YYYYMMDD) format.
//   - If the task has a repeat rule (task.Repeat != ""), the next occurrence after
//     the current date is computed using NextDate.
//   - If the provided date is in the past (before now), the date is adjusted:
//   - For non‑repeating tasks: set to today.
//   - For repeating tasks: set to the next occurrence after now.
//
// Returns an error if the date format is invalid or the repeat rule is malformed.
func checkDate(task *db.Task) error {
	now := time.Now()
	if task.Date == "" {
		task.Date = now.Format("20060102")
		return nil
	}
	t, err := time.Parse("20060102", task.Date)
	if err != nil {
		return errors.New("некорректный формат даты, требуется ГГГГММДД")
	}
	var next string
	if task.Repeat != "" {
		next, err = NextDate(now, task.Date, task.Repeat)
		if err != nil {
			return errors.New("некорректное правило повторения")
		}
	}
	if afterNow(now, t) {
		if len(task.Repeat) == 0 {
			task.Date = now.Format("20060102")
		} else {
			task.Date = next
		}
	}
	return nil
}

// writeJson is a helper that writes a JSON response with the correct Content‑Type header.
// If encoding fails, the error is silently ignored because the headers have already been written.
// In production, you might want to log such errors for debugging.
func writeJson(w http.ResponseWriter, data any, statusCode int) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(statusCode)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		// If encoding fails, we cannot change the headers anymore
		// Log the error but cannot send a different response
	}
}

// addTaskHandler handles POST /api/task requests (creating a new task).
// Expects a JSON body with fields:
//   - title (required)
//   - date (optional, format "20060102", defaults to today)
//   - comment (optional)
//   - repeat (optional, repetition rule)
//
// Validations:
//   - Title must not be empty.
//   - Date must be valid (see checkDate).
//   - Repeat rule must be valid (if provided).
//
// On success, returns a JSON object with the auto‑generated task ID.
// On error, returns a JSON object with an "error" field describing the problem.
func addTaskHandler(w http.ResponseWriter, r *http.Request) {
	var task db.Task
	if err := json.NewDecoder(r.Body).Decode(&task); err != nil {
		writeJson(w, map[string]string{"error": "неверный формат JSON"}, http.StatusBadRequest)
		return
	}
	if task.Title == "" {
		writeJson(w, map[string]string{"error": "Не указан заголовок задачи"}, http.StatusBadRequest)
		return
	}

	if err := checkDate(&task); err != nil {
		writeJson(w, map[string]string{"error": err.Error()}, http.StatusBadRequest)
		return
	}
	id, err := db.AddTask(&task)
	if err != nil {
		writeJson(w, map[string]string{"error": "ошибка сохранения задачи в базу данных"}, http.StatusInternalServerError)
		return
	}
	writeJson(w, map[string]string{"id": fmt.Sprintf("%d", id)}, http.StatusOK)
}
