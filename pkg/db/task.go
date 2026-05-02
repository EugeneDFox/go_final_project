package db

import (
	"database/sql"
	"fmt"
)

// Task represents a single task in the scheduler.
// It maps to a row in the "scheduler" table.
type Task struct {
	ID      string `json:"id"`      // Unique identifier (auto-incremented)
	Date    string `json:"date"`    // Date in "20060102" format (YYYYMMDD)
	Title   string `json:"title"`   // Title of the task (required)
	Comment string `json:"comment"` // Optional comment/description
	Repeat  string `json:"repeat"`  // Repeat pattern (e.g., "y", "d 3", "m 1,15")
}

// AddTask inserts a new task into the database.
// It returns the auto-generated ID of the inserted task, or an error.
func AddTask(task *Task) (int64, error) {
	var id int64
	query := `INSERT INTO scheduler (date, title, comment, repeat) VALUES (?, ?, ?, ?)`
	res, err := db.Exec(query, task.Date, task.Title, task.Comment, task.Repeat)
	if err == nil {
		id, err = res.LastInsertId()
	}
	return id, err
}

// Tasks retrieves a list of tasks from the database, ordered by date ascending.
// The limit parameter controls the maximum number of tasks returned.
// Returns a slice of Task pointers, or an error if the query fails.
func Tasks(limit int) ([]*Task, error) {
	query := `SELECT id, date, title, comment, repeat FROM scheduler ORDER BY date ASC LIMIT ?`
	rows, err := db.Query(query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	tasks := make([]*Task, 0)
	for rows.Next() {
		task := &Task{}
		err := rows.Scan(&task.ID, &task.Date, &task.Title, &task.Comment, &task.Repeat)
		if err != nil {
			return nil, err
		}
		tasks = append(tasks, task)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return tasks, nil
}

// SearchParamTasks searches for tasks where the title or comment contains the searchParam
// (case‑insensitive). Results are ordered by date and limited to 'limit' rows.
func SearchParamTasks(searchParam string, limit int) ([]*Task, error) {
	query := `SELECT * FROM scheduler WHERE title LIKE ? COLLATE NOCASE OR comment LIKE ? COLLATE NOCASE ORDER BY date LIMIT ?`
	rows, err := db.Query(query, searchParam, searchParam, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	tasks := make([]*Task, 0)
	for rows.Next() {
		task := &Task{}
		err := rows.Scan(&task.ID, &task.Date, &task.Title, &task.Comment, &task.Repeat)
		if err != nil {
			return nil, err
		}
		tasks = append(tasks, task)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return tasks, nil
}

// SearchDataTasks retrieves tasks that have a specific date (exact match).
// The date parameter must be in "20060102" format.
// Returns up to 'limit' tasks, or an error.
func SearchDataTasks(data string, limit int) ([]*Task, error) {
	query := `SELECT * FROM scheduler WHERE date = ? LIMIT ?`
	rows, err := db.Query(query, data, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	tasks := make([]*Task, 0)
	for rows.Next() {
		task := &Task{}
		err := rows.Scan(&task.ID, &task.Date, &task.Title, &task.Comment, &task.Repeat)
		if err != nil {
			return nil, err
		}
		tasks = append(tasks, task)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return tasks, nil
}

// GetTask fetches a single task by its ID.
// If no task with the given ID exists, it returns an error with message "Задача не найдена".
func GetTask(id string) (*Task, error) {
	query := `SELECT id, date, title, comment, repeat FROM scheduler WHERE id = ?`
	task := &Task{}
	err := db.QueryRow(query, id).Scan(&task.ID, &task.Date, &task.Title, &task.Comment, &task.Repeat)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("Задача не найдена")
		}
		return nil, err
	}
	return task, nil
}

// UpdateTask updates an existing task identified by task.ID.
// All fields (date, title, comment, repeat) are updated.
// Returns an error if the task does not exist or the update fails.
func UpdateTask(task *Task) error {
	query := `UPDATE scheduler SET date=?, title=?, comment=?, repeat=? WHERE id=?`
	res, err := db.Exec(query, task.Date, task.Title, task.Comment, task.Repeat, task.ID)
	if err != nil {
		return err
	}
	count, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if count == 0 {
		return fmt.Errorf(`incorrect id for updating task`)
	}
	return nil
}

// DeleteTask removes a task from the database by its ID.
// Returns an error if the task does not exist or the deletion fails.
func DeleteTask(id string) error {
	query := `DELETE FROM scheduler WHERE id = ?`
	res, err := db.Exec(query, id)
	if err != nil {
		return err
	}
	count, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if count == 0 {
		return fmt.Errorf("Задача не найдена")
	}
	return nil
}

// UpdateDate updates only the date field of a task identified by id.
// This is used when a repeating task moves to its next occurrence.
// Returns an error if the task does not exist or the update fails.
func UpdateDate(next string, id string) error {
	query := `UPDATE scheduler SET date = ? WHERE id = ?`
	res, err := db.Exec(query, next, id)
	if err != nil {
		return err
	}
	count, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if count == 0 {
		return fmt.Errorf("Задача не найдена")
	}
	return nil
}
