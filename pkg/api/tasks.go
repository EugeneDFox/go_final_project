package api

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/EugeneDFox/go_final_project/pkg/db"
)

// TasksResp is the JSON response structure for the /api/tasks endpoint.
type TasksResp struct {
	Tasks []*db.Task `json:"tasks"`
}

// tasksHandler handles GET /api/tasks requests.
// Query parameters:
//   - search (optional): if present, performs either a date search or a text search.
//     Date search: if the search string can be parsed as "02.01.2006" (DD.MM.YYYY),
//     returns tasks with that exact date.
//     Text search: otherwise, searches for tasks where title or comment contains
//     the search string (case‑insensitive, partial match).
//   - If no search parameter is provided, returns all tasks ordered by date (limit 50).
func tasksHandler(w http.ResponseWriter, r *http.Request) {
	search := r.URL.Query().Get("search")
	if search == "" {
		// No search parameter, return all tasks
		tasks, err := db.Tasks(50)
		if err != nil {
			writeJson(w, map[string]string{"error": err.Error()}, http.StatusInternalServerError)
			return
		}
		writeJson(w, TasksResp{
			Tasks: tasks,
		}, http.StatusOK)
		return
	}
	// Search parameter present
	parsedDate, dateParseErr := time.Parse("02.01.2006", search)
	if dateParseErr == nil {
		// Date search
		date := parsedDate.Format("20060102")
		tasks, err := db.SearchDataTasks(date, 50)
		if err != nil {
			writeJson(w, map[string]string{"error": err.Error()}, http.StatusInternalServerError)
			return
		}
		writeJson(w, TasksResp{
			Tasks: tasks,
		}, http.StatusOK)
	} else {
		// Text search
		param := "%" + search + "%"
		tasks, err := db.SearchParamTasks(param, 50)
		if err != nil {
			writeJson(w, map[string]string{"error": err.Error()}, http.StatusInternalServerError)
			return
		}
		writeJson(w, TasksResp{
			Tasks: tasks,
		}, http.StatusOK)
	}
}

// taskHandler handles GET, PUT, and DELETE requests to /api/task.
// The method is determined by the HTTP verb:
//   - GET: retrieves a single task by ID (query parameter "id").
//   - PUT: updates an existing task. Expects a JSON body with all task fields.
//     Validates that title is not empty and the date is valid (via checkDate).
//   - DELETE: deletes a task by ID (query parameter "id").
func taskHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		handleGetTask(w, r)
	case "PUT":
		handleUpdateTask(w, r)
	case "DELETE":
		handleDeleteTask(w, r)
	default:
		writeJson(w, map[string]string{"error": "Метод не поддерживается"}, http.StatusMethodNotAllowed)
	}

}

// handleGetTask handles GET /api/task requests to retrieve a single task by ID.
// Query parameter "id" is required. Returns the task as JSON or an error if not found.
func handleGetTask(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	if id == "" {
		writeJson(w, map[string]string{"error": "Не указан идентификатор"}, http.StatusBadRequest)
		return
	}
	task, err := db.GetTask(id)
	if err != nil {
		writeJson(w, map[string]string{"error": err.Error()}, http.StatusNotFound)
		return
	}
	writeJson(w, task, http.StatusOK)
}

// handleUpdateTask handles PUT /api/task requests to update an existing task.
// Expects a JSON body with all task fields. Validates title and date.
// Returns empty JSON on success or an error message on failure.
func handleUpdateTask(w http.ResponseWriter, r *http.Request) {
	var task db.Task
	if err := json.NewDecoder(r.Body).Decode(&task); err != nil {
		writeJson(w, map[string]string{"error": "Некорректный JSON"}, http.StatusBadRequest)
		return
	}
	if task.Title == "" {
		writeJson(w, map[string]string{"error": "Заголовок не может быть пустым"}, http.StatusBadRequest)
		return
	}
	if err := checkDate(&task); err != nil {
		writeJson(w, map[string]string{"error": err.Error()}, http.StatusBadRequest)
		return
	}
	if err := db.UpdateTask(&task); err != nil {
		writeJson(w, map[string]string{"error": err.Error()}, http.StatusInternalServerError)
		return
	}
	writeJson(w, struct{}{}, http.StatusOK)
}

// handleDeleteTask handles DELETE /api/task requests to remove a task by ID.
// Query parameter "id" is required. Returns empty JSON on success or an error if not found.
func handleDeleteTask(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	if id == "" {
		writeJson(w, map[string]string{"error": "Не указан идентификатор"}, http.StatusBadRequest)
		return
	}
	if err := db.DeleteTask(id); err != nil {
		writeJson(w, map[string]string{"error": err.Error()}, http.StatusInternalServerError)
		return
	}
	writeJson(w, struct{}{}, http.StatusOK)
}

// taskDoneHandler handles POST /api/task/done requests.
// Query parameter "id" identifies the task to mark as done.
//   - If the task has no repeat rule (task.Repeat == ""), the task is deleted.
//   - If the task repeats, its date is updated to the next occurrence using NextDate
//     (based on the current date as reference).
//
// Returns an empty JSON object on success, or an error message on failure.
func taskDoneHandler(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	if id == "" {
		writeJson(w, map[string]string{"error": "Не указан идентификатор"}, http.StatusBadRequest)
		return
	}
	task, err := db.GetTask(id)
	if err != nil {
		writeJson(w, map[string]string{"error": err.Error()}, http.StatusNotFound)
		return
	}
	if task.Repeat == "" {
		if err := db.DeleteTask(id); err != nil {
			writeJson(w, map[string]string{"error": err.Error()}, http.StatusInternalServerError)
			return
		}
	} else {
		now := time.Now()
		nextDate, err := NextDate(now, task.Date, task.Repeat)
		if err != nil {
			writeJson(w, map[string]string{"error": "некорректное правило повторения"}, http.StatusBadRequest)
			return
		}
		if err := db.UpdateDate(nextDate, id); err != nil {
			writeJson(w, map[string]string{"error": err.Error()}, http.StatusInternalServerError)
			return
		}
	}
	writeJson(w, struct{}{}, http.StatusOK)
}
