package database

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

type DB struct {
	sync.RWMutex
	path string
	data Data
}

type Data struct {
	Users []User           `json:"users"`
	Files []FileMetadata   `json:"files"`
}

type User struct {
	ID        int64     `json:"id"`
	Email     string    `json:"email"`
	Password  string    `json:"-"`
	CreatedAt time.Time `json:"created_at"`
}

type FileMetadata struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name"`
	Size      int64     `json:"size"`
	UserID    int64     `json:"user_id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func NewDB(path string) (*DB, error) {
	db := &DB{
		path: path,
		data: Data{
			Users: make([]User, 0),
			Files: make([]FileMetadata, 0),
		},
	}

	if err := db.load(); err != nil && !os.IsNotExist(err) {
		return nil, err
	}

	return db, nil
}

func (db *DB) load() error {
	db.Lock()
	defer db.Unlock()

	data, err := os.ReadFile(db.path)
	if err != nil {
		return err
	}

	return json.Unmarshal(data, &db.data)
}

func (db *DB) save() error {
	db.Lock()
	defer db.Unlock()

	dir := filepath.Dir(db.path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(db.data, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(db.path, data, 0644)
}

func (db *DB) CreateUser(email, password string) (*User, error) {
	db.Lock()
	defer db.Unlock()

	// Check if user exists
	for _, u := range db.data.Users {
		if u.Email == email {
			return nil, fmt.Errorf("user already exists")
		}
	}

	user := User{
		ID:        time.Now().UnixNano(),
		Email:     email,
		Password:  password,
		CreatedAt: time.Now(),
	}

	db.data.Users = append(db.data.Users, user)
	if err := db.save(); err != nil {
		return nil, err
	}

	return &user, nil
}

func (db *DB) GetUser(id int64) (*User, error) {
	db.RLock()
	defer db.RUnlock()

	for _, u := range db.data.Users {
		if u.ID == id {
			return &u, nil
		}
	}

	return nil, fmt.Errorf("user not found")
}

func (db *DB) GetUserByEmail(email string) (*User, error) {
	db.RLock()
	defer db.RUnlock()

	for _, u := range db.data.Users {
		if u.Email == email {
			return &u, nil
		}
	}

	return nil, fmt.Errorf("user not found")
}

func (db *DB) CreateFileMetadata(name string, size int64, userID int64) (*FileMetadata, error) {
	db.Lock()
	defer db.Unlock()

	file := FileMetadata{
		ID:        time.Now().UnixNano(),
		Name:      name,
		Size:      size,
		UserID:    userID,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	db.data.Files = append(db.data.Files, file)
	if err := db.save(); err != nil {
		return nil, err
	}

	return &file, nil
}

func (db *DB) GetFileMetadata(id int64) (*FileMetadata, error) {
	db.RLock()
	defer db.RUnlock()

	for _, f := range db.data.Files {
		if f.ID == id {
			return &f, nil
		}
	}

	return nil, fmt.Errorf("file not found")
}

func (db *DB) GetUserFiles(userID int64) ([]FileMetadata, error) {
	db.RLock()
	defer db.RUnlock()

	var files []FileMetadata
	for _, f := range db.data.Files {
		if f.UserID == userID {
			files = append(files, f)
		}
	}

	return files, nil
}

func (db *DB) DeleteFileMetadata(id int64) error {
	db.Lock()
	defer db.Unlock()

	for i, f := range db.data.Files {
		if f.ID == id {
			// Remove the file from the slice
			db.data.Files = append(db.data.Files[:i], db.data.Files[i+1:]...)
			return db.save()
		}
	}

	return fmt.Errorf("file not found")
}

func (db *DB) Close() error {
	return db.save()
}
