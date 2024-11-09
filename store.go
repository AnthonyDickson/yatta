package main

// Handles the creation and retrieval of tasks.
type TaskStore interface {
	GetTasks(user string) ([]Task, error)
	AddTask(user string, description string) error
}
