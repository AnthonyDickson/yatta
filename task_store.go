package main

// Handles the creation and retrieval of tasks.
type TaskStore interface {
	// Get all tasks (possibly an empty slice) for `user`.
	//
	// Returns an empty slice and error if something prevented the tasks from being retrieved from the store.
	GetTasks(user string) ([]Task, error)

	// Get a single task by its `id`.
	//
	// Returns `nil` if a task with `id` was not found.
	//
	// Returns `nil` and an error if something prevented the tasks from being retrieved from the store.
	GetTask(id uint64) (*Task, error)

	// Create and add a new task for `user`.
	//
	// Returns an error if something prevented the task from being created or added to the store.
	AddTask(user string, description string) error
}
