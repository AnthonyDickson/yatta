package main

// Handles the creation and retrieval of todos.
type TodoStore interface {
	GetTodos(user string) ([]Task, error)
	AddTodo(user string, description string) error
}
