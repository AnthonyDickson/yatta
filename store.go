package main

// Handles the creation and retrieval of todos.
type TodoStore interface {
	GetTodos(user string) ([]string, error)
	AddTodo(user string, task string) error
}
