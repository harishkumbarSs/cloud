package storage

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
)

// FileStorage handles file operations
type FileStorage struct {
	uploadDir string
}

// NewFileStorage creates a new FileStorage instance
func NewFileStorage(uploadDir string) (*FileStorage, error) {
	// Create upload directory if it doesn't exist
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create upload directory: %v", err)
	}
	return &FileStorage{uploadDir: uploadDir}, nil
}

// SaveFile saves an uploaded file to disk
func (fs *FileStorage) SaveFile(filename string, content io.Reader) error {
	filepath := filepath.Join(fs.uploadDir, filename)
	
	// Create the file
	dst, err := os.Create(filepath)
	if err != nil {
		return fmt.Errorf("failed to create file: %v", err)
	}
	defer dst.Close()

	// Copy the content
	if _, err := io.Copy(dst, content); err != nil {
		return fmt.Errorf("failed to save file: %v", err)
	}

	return nil
}

// GetFile retrieves a file by name
func (fs *FileStorage) GetFile(filename string) (*os.File, error) {
	filepath := filepath.Join(fs.uploadDir, filename)
	file, err := os.Open(filepath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %v", err)
	}
	return file, nil
}

// ListFiles returns a list of all files in the upload directory
func (fs *FileStorage) ListFiles() ([]string, error) {
	var files []string
	
	entries, err := os.ReadDir(fs.uploadDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read directory: %v", err)
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			files = append(files, entry.Name())
		}
	}

	return files, nil
}

// DeleteFile deletes a file by name
func (fs *FileStorage) DeleteFile(filename string) error {
	filepath := filepath.Join(fs.uploadDir, filename)
	if err := os.Remove(filepath); err != nil {
		return fmt.Errorf("failed to delete file: %v", err)
	}
	return nil
}

type FileMetadata struct {
	UserID   string
	Filename string
	Size     int64
	Path     string
}

func getUserStoragePath(userID string) string {
	return filepath.Join("uploads", userID)
}

func SaveUserFile(userID, filename string, file io.Reader) (*FileMetadata, error) {
	// Create user-specific storage directory
	userDir := getUserStoragePath(userID)
	if err := os.MkdirAll(userDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create user directory: %v", err)
	}

	// Generate unique file path
	filePath := filepath.Join(userDir, filename)
	destFile, err := os.Create(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to create file: %v", err)
	}
	defer destFile.Close()

	// Copy file content
	size, err := io.Copy(destFile, file)
	if err != nil {
		return nil, fmt.Errorf("failed to save file: %v", err)
	}

	metadata := &FileMetadata{
		UserID:   userID,
		Filename: filename,
		Size:     size,
		Path:     filePath,
	}

	log.Printf("Saved file for user %s: %s (size: %d bytes)", userID, filename, size)
	return metadata, nil
}

func ListUserFiles(userID string) ([]FileMetadata, error) {
	userDir := getUserStoragePath(userID)
	
	// Check if directory exists
	if _, err := os.Stat(userDir); os.IsNotExist(err) {
		return []FileMetadata{}, nil
	}

	files, err := os.ReadDir(userDir)
	if err != nil {
		return nil, fmt.Errorf("failed to list files: %v", err)
	}

	var fileMetadata []FileMetadata
	for _, file := range files {
		if !file.IsDir() {
			filePath := filepath.Join(userDir, file.Name())
			fileInfo, err := file.Info()
			if err != nil {
				log.Printf("Error getting file info: %v", err)
				continue
			}

			fileMetadata = append(fileMetadata, FileMetadata{
				UserID:   userID,
				Filename: file.Name(),
				Size:     fileInfo.Size(),
				Path:     filePath,
			})
		}
	}

	return fileMetadata, nil
}

func GetUserFile(userID, filename string) (*os.File, error) {
	filePath := filepath.Join(getUserStoragePath(userID), filename)
	
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("file not found: %v", err)
	}

	return file, nil
}

func DeleteUserFile(userID, filename string) error {
	filePath := filepath.Join(getUserStoragePath(userID), filename)
	
	if err := os.Remove(filePath); err != nil {
		return fmt.Errorf("failed to delete file: %v", err)
	}

	log.Printf("Deleted file %s for user %s", filename, userID)
	return nil
}
