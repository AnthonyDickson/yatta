package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
)

// Persists todos to disk.
type FileTodoStore struct {
	database *os.File
}

func NewFileTodoStore(database *os.File) *FileTodoStore {
	return &FileTodoStore{database: database}
}

func (f *FileTodoStore) GetTodos(user string) ([]string, error) {
	todoLists, err := f.getTodoLists()

	if err != nil {
		return nil, fmt.Errorf("could not get todos: %v", err)
	}

	todoList := todoLists.find(user)

	if todoList != nil {
		return todoList.Tasks, nil
	}

	return nil, nil
}

func (f *FileTodoStore) AddTodo(user string, todo string) error {
	todoLists, err := f.getTodoLists()

	if err != nil {
		return fmt.Errorf("could not get todos: %v", err)
	}

	userTodoList := todoLists.find(user)

	if userTodoList != nil {
		userTodoList.Tasks = append(userTodoList.Tasks, todo)
	} else {
		*todoLists = append(*todoLists, todoList{user, []string{todo}})
	}

	return f.updateDatabase(todoLists)
}

func (f *FileTodoStore) getTodoLists() (*todoLists, error) {
	_, err := f.database.Seek(0, io.SeekStart)

	if err != nil {
		return nil, fmt.Errorf("could not seek to the start of the database file: %v", err)
	}

	todoLists, err := newTodoLists(f.database)

	return &todoLists, err
}

func (f *FileTodoStore) updateDatabase(todoLists *todoLists) error {
	err := f.database.Truncate(0) // prevents writes that are smaller than previous file from leading to invalid JSON

	if err != nil {
		return fmt.Errorf("could not truncate database file to prepare for write: %v", err)
	}

	_, err = f.database.Seek(0, io.SeekStart)

	if err != nil {
		return fmt.Errorf("could not seek database file: %v", err)
	}

	err = json.NewEncoder(f.database).Encode(todoLists)

	return err
}

// A list of todos for a user.
type todoList struct {
	User  string
	Tasks []string
}

type todoLists []todoList

// Parse a list of todosList objects from `database`.
func newTodoLists(database *os.File) (todoLists, error) {
	data, err := io.ReadAll(database)

	if err != nil {
		return nil, fmt.Errorf("could not read database: %v", err)
	}

	// returning an empty slice avoids errors when decoding a new, empty file.
	if len(data) == 0 {
		return nil, nil
	}

	var todos todoLists
	err = json.Unmarshal(data, &todos)

	if err != nil {
		return nil, fmt.Errorf("could not decode the todo store %s: %v", data, err)
	}

	return todos, nil
}

// Search a `todoLists` for the todos for `user`.
// Returns `nil` if not found.
func (t todoLists) find(user string) *todoList {
	for i, todoList := range t {
		if todoList.User == user {
			return &t[i]
		}
	}

	return nil
}
