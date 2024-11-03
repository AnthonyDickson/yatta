package main

// Handles the creation and retrieval of todos.
type TodoStore interface {
	GetTodos(user string) []string
	AddTodo(user string, task string)
}
