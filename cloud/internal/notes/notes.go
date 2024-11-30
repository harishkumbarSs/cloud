package notes

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/google/uuid"
)

type Note struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	Title     string    `json:"title"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

var (
	notesLock sync.RWMutex
)

func getUserNotesPath(userID string) string {
	return filepath.Join("notes", userID)
}

func CreateNote(note Note) (*Note, error) {
	notesLock.Lock()
	defer notesLock.Unlock()

	// Create user-specific notes directory
	userDir := getUserNotesPath(note.UserID)
	if err := os.MkdirAll(userDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create user notes directory: %v", err)
	}

	// Generate unique note ID
	note.ID = uuid.New().String()
	note.CreatedAt = time.Now()
	note.UpdatedAt = time.Now()

	// Save note to file
	filePath := filepath.Join(userDir, note.ID+".json")
	noteJSON, err := json.Marshal(note)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal note: %v", err)
	}

	if err := os.WriteFile(filePath, noteJSON, 0644); err != nil {
		return nil, fmt.Errorf("failed to save note: %v", err)
	}

	log.Printf("Created note for user %s: %s", note.UserID, note.Title)
	return &note, nil
}

func ListUserNotes(userID string) ([]Note, error) {
	notesLock.RLock()
	defer notesLock.RUnlock()

	userDir := getUserNotesPath(userID)

	// Check if directory exists
	if _, err := os.Stat(userDir); os.IsNotExist(err) {
		return []Note{}, nil
	}

	files, err := os.ReadDir(userDir)
	if err != nil {
		return nil, fmt.Errorf("failed to list notes: %v", err)
	}

	var notes []Note
	for _, file := range files {
		if !file.IsDir() && filepath.Ext(file.Name()) == ".json" {
			filePath := filepath.Join(userDir, file.Name())
			noteData, err := os.ReadFile(filePath)
			if err != nil {
				log.Printf("Error reading note file: %v", err)
				continue
			}

			var note Note
			if err := json.Unmarshal(noteData, &note); err != nil {
				log.Printf("Error unmarshaling note: %v", err)
				continue
			}

			notes = append(notes, note)
		}
	}

	return notes, nil
}

func UpdateNote(noteID, userID string, updatedNote Note) (*Note, error) {
	notesLock.Lock()
	defer notesLock.Unlock()

	filePath := filepath.Join(getUserNotesPath(userID), noteID+".json")

	// Read existing note
	noteData, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("note not found: %v", err)
	}

	var existingNote Note
	if err := json.Unmarshal(noteData, &existingNote); err != nil {
		return nil, fmt.Errorf("failed to unmarshal note: %v", err)
	}

	// Update note fields
	existingNote.Title = updatedNote.Title
	existingNote.Content = updatedNote.Content
	existingNote.UpdatedAt = time.Now()

	// Save updated note
	updatedNoteJSON, err := json.Marshal(existingNote)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal updated note: %v", err)
	}

	if err := os.WriteFile(filePath, updatedNoteJSON, 0644); err != nil {
		return nil, fmt.Errorf("failed to save updated note: %v", err)
	}

	log.Printf("Updated note %s for user %s", noteID, userID)
	return &existingNote, nil
}

func DeleteNote(noteID, userID string) error {
	notesLock.Lock()
	defer notesLock.Unlock()

	filePath := filepath.Join(getUserNotesPath(userID), noteID+".json")

	if err := os.Remove(filePath); err != nil {
		return fmt.Errorf("failed to delete note: %v", err)
	}

	log.Printf("Deleted note %s for user %s", noteID, userID)
	return nil
}
