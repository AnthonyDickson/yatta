package stores

import (
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/AnthonyDickson/yatta/models"
)

// Persists tasks to disk.
type FileTaskStore struct {
	database  *json.Encoder
	taskLists taskLists
}

func NewFileTaskStore(database *os.File) (*FileTaskStore, error) {
	taskLists, err := newTaskLists(database)

	if err != nil {
		return nil, fmt.Errorf("could not parse task lists: %v", err)
	}

	store := &FileTaskStore{
		database:  json.NewEncoder(&tape{database}),
		taskLists: taskLists,
	}

	return store, nil
}

func (f *FileTaskStore) GetTasks(user string) ([]models.Task, error) {
	taskList := f.taskLists.find(user)

	if taskList != nil {
		return taskList.Tasks, nil
	}

	return nil, nil
}

func (f *FileTaskStore) GetTask(id uint64) (*models.Task, error) {
	for _, taskList := range f.taskLists {
		for _, task := range taskList.Tasks {
			if task.ID == id {
				return &task, nil
			}
		}
	}

	return nil, nil
}

func (f *FileTaskStore) AddTask(user string, description string) error {
	userTaskList := f.taskLists.find(user)
	id := f.taskLists.nextID()
	task := models.Task{ID: id, Description: description}

	if userTaskList != nil {
		userTaskList.Tasks = append(userTaskList.Tasks, task)
	} else {
		f.taskLists = append(f.taskLists, taskList{user, []models.Task{task}})
	}

	return f.database.Encode(f.taskLists)
}

// A list of tasks for a user.
type taskList struct {
	User  string
	Tasks []models.Task
}

type taskLists []taskList

// Parse a list of tasksList objects from `database`.
func newTaskLists(database *os.File) (taskLists, error) {
	_, err := database.Seek(0, io.SeekStart)

	if err != nil {
		return nil, fmt.Errorf("could not seek database file: %v", err)
	}

	data, err := io.ReadAll(database)

	if err != nil {
		return nil, fmt.Errorf("could not read database: %v", err)
	}

	// returning an empty slice avoids errors when decoding a new, empty file.
	if len(data) == 0 {
		return nil, nil
	}

	var tasks taskLists
	err = json.Unmarshal(data, &tasks)

	if err != nil {
		return nil, fmt.Errorf("could not decode the task store %s: %v", data, err)
	}

	return tasks, nil
}

// Search a `taskLists` for the tasks for `user`.
// Returns `nil` if not found.
func (t taskLists) find(user string) *taskList {
	for i, taskList := range t {
		if taskList.User == user {
			return &t[i]
		}
	}

	return nil
}

// Use this function when setting the ID of a new task to ensure that the ID is auto-incremented and unique.
func (t taskLists) nextID() (id uint64) {
	id = 0

	for _, taskList := range t {
		for _, task := range taskList.Tasks {
			id = max(id, task.ID)
		}
	}

	return id + 1
}
