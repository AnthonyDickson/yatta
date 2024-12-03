package stores

import (
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/AnthonyDickson/yatta/models"
)

type FileUserStore struct {
	database *json.Encoder
	users    userList
}

func NewFileUserStore(database *os.File) (*FileUserStore, error) {
	users, err := loadUserStore(database)

	if err != nil {
		return nil, err
	}

	store := &FileUserStore{
		database: json.NewEncoder(&tape{database}),
		users:    users,
	}

	return store, nil
}

func loadUserStore(database *os.File) (userList, error) {
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

	var users userList
	err = json.Unmarshal(data, &users)

	if err != nil {
		return nil, fmt.Errorf("could not decode JSON database: %v", err)
	}

	return users, nil
}

func (f *FileUserStore) AddUser(email, password string) error {
	nextID := f.users.nextID()
	f.users = append(f.users, models.User{ID: nextID, Email: email, Password: password})

	return f.database.Encode(f.users)
}

func (f *FileUserStore) GetUser(id uint64) (*models.User, error) {
	return f.users.find(id), nil
}

type userList []models.User

func (u userList) find(id uint64) *models.User {
	for _, user := range u {
		if user.ID == id {
			return &user
		}
	}

	return nil
}

func (u userList) nextID() uint64 {
	var maxID uint64 = 0

	for _, user := range u {
		maxID = max(maxID, user.ID)
	}

	return maxID + 1
}
