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
	database *os.File
}

func NewFileTaskStore(database *os.File) *FileTaskStore {
	return &FileTaskStore{database: database}
}

func (f *FileTaskStore) GetTasks(user string) ([]models.Task, error) {
	taskLists, err := f.getTaskLists()

	if err != nil {
		return nil, fmt.Errorf("could not get tasks: %v", err)
	}

	taskList := taskLists.find(user)

	if taskList != nil {
		return taskList.Tasks, nil
	}

	return nil, nil
}

func (f *FileTaskStore) GetTask(id uint64) (*models.Task, error) {
	taskLists, err := f.getTaskLists()

	if err != nil {
		return nil, fmt.Errorf("could not get tasks: %v", err)
	}

	for _, taskList := range *taskLists {
		for _, task := range taskList.Tasks {
			if task.ID == id {
				return &task, nil
			}
		}
	}

	return nil, nil
}

func (f *FileTaskStore) AddTask(user string, description string) error {
	taskLists, err := f.getTaskLists()

	if err != nil {
		return fmt.Errorf("could not get tasks: %v", err)
	}

	userTaskList := taskLists.find(user)
	task := models.Task{ID: 0, Description: description}

	if userTaskList != nil {
		userTaskList.Tasks = append(userTaskList.Tasks, task)
	} else {
		*taskLists = append(*taskLists, taskList{user, []models.Task{task}})
	}

	return f.updateDatabase(taskLists)
}

func (f *FileTaskStore) getTaskLists() (*taskLists, error) {
	_, err := f.database.Seek(0, io.SeekStart)

	if err != nil {
		return nil, fmt.Errorf("could not seek to the start of the database file: %v", err)
	}

	taskLists, err := newTaskLists(f.database)

	return &taskLists, err
}

func (f *FileTaskStore) updateDatabase(taskLists *taskLists) error {
	err := f.database.Truncate(0) // prevents writes that are smaller than previous file from leading to invalid JSON

	if err != nil {
		return fmt.Errorf("could not truncate database file to prepare for write: %v", err)
	}

	_, err = f.database.Seek(0, io.SeekStart)

	if err != nil {
		return fmt.Errorf("could not seek database file: %v", err)
	}

	err = json.NewEncoder(f.database).Encode(taskLists)

	return err
}

// A list of tasks for a user.
type taskList struct {
	User  string
	Tasks []models.Task
}

type taskLists []taskList

// Parse a list of tasksList objects from `database`.
func newTaskLists(database *os.File) (taskLists, error) {
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
