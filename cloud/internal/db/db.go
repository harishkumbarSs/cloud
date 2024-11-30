package db

import (
    "log"
    "time"

    "github.com/gocql/gocql"
)

var Session *gocql.Session

type User struct {
    Email     string    `json:"email"`
    Name      string    `json:"name"`
    CreatedAt time.Time `json:"created_at"`
    LastLogin time.Time `json:"last_login"`
}

type File struct {
    UserEmail    string    `json:"user_email"`
    FileID       string    `json:"file_id"`
    Filename     string    `json:"filename"`
    Size         int64     `json:"size"`
    ContentType  string    `json:"content_type"`
    StoragePath  string    `json:"storage_path"`
    UploadedAt   time.Time `json:"uploaded_at"`
}

type Note struct {
    UserEmail  string    `json:"user_email"`
    NoteID     string    `json:"note_id"`
    Title      string    `json:"title"`
    Content    string    `json:"content"`
    CreatedAt  time.Time `json:"created_at"`
    UpdatedAt  time.Time `json:"updated_at"`
}

func InitDB() error {
    cluster := gocql.NewCluster("localhost:9042")
    cluster.Keyspace = "cloud_storage"
    cluster.Consistency = gocql.Quorum
    cluster.ConnectTimeout = time.Second * 10

    var err error
    Session, err = cluster.CreateSession()
    if err != nil {
        return err
    }

    log.Println("Database connection established")
    return nil
}

// User operations
func CreateUser(user User) error {
    return Session.Query(`
        INSERT INTO users (email, name, created_at, last_login)
        VALUES (?, ?, ?, ?)`,
        user.Email, user.Name, user.CreatedAt, user.LastLogin,
    ).Exec()
}

func GetUser(email string) (User, error) {
    var user User
    err := Session.Query(`
        SELECT email, name, created_at, last_login
        FROM users WHERE email = ?`, email,
    ).Scan(&user.Email, &user.Name, &user.CreatedAt, &user.LastLogin)
    return user, err
}

// File operations
func SaveFileMetadata(file File) error {
    return Session.Query(`
        INSERT INTO files (user_email, file_id, filename, size, content_type, storage_path, uploaded_at)
        VALUES (?, ?, ?, ?, ?, ?, ?)`,
        file.UserEmail, file.FileID, file.Filename, file.Size, file.ContentType, file.StoragePath, file.UploadedAt,
    ).Exec()
}

func GetUserFiles(userEmail string) ([]File, error) {
    var files []File
    iter := Session.Query(`
        SELECT user_email, file_id, filename, size, content_type, storage_path, uploaded_at
        FROM files WHERE user_email = ?`, userEmail,
    ).Iter()

    var file File
    for iter.Scan(
        &file.UserEmail, &file.FileID, &file.Filename, &file.Size,
        &file.ContentType, &file.StoragePath, &file.UploadedAt,
    ) {
        files = append(files, file)
    }
    return files, iter.Close()
}

func DeleteFile(userEmail, filename string) error {
    return Session.Query(`
        DELETE FROM files
        WHERE user_email = ? AND filename = ?`,
        userEmail, filename,
    ).Exec()
}

// Note operations
func SaveNote(note Note) error {
    return Session.Query(`
        INSERT INTO notes (user_email, note_id, title, content, created_at, updated_at)
        VALUES (?, ?, ?, ?, ?, ?)`,
        note.UserEmail, note.NoteID, note.Title, note.Content, note.CreatedAt, note.UpdatedAt,
    ).Exec()
}

func GetUserNotes(userEmail string) ([]Note, error) {
    var notes []Note
    iter := Session.Query(`
        SELECT user_email, note_id, title, content, created_at, updated_at
        FROM notes WHERE user_email = ?`, userEmail,
    ).Iter()

    var note Note
    for iter.Scan(
        &note.UserEmail, &note.NoteID, &note.Title, &note.Content,
        &note.CreatedAt, &note.UpdatedAt,
    ) {
        notes = append(notes, note)
    }
    return notes, iter.Close()
}

func UpdateNote(note Note) error {
    return Session.Query(`
        UPDATE notes SET title = ?, content = ?, updated_at = ?
        WHERE user_email = ? AND note_id = ?`,
        note.Title, note.Content, note.UpdatedAt, note.UserEmail, note.NoteID,
    ).Exec()
}

func DeleteNote(userEmail, noteID string) error {
    return Session.Query(`
        DELETE FROM notes WHERE user_email = ? AND note_id = ?`,
        userEmail, noteID,
    ).Exec()
}
