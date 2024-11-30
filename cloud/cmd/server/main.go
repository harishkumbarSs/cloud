package main

import (
    "log"
    "net/http"
    "os"
    "path/filepath"
    "encoding/json"
    "io/ioutil"
    "io"
    "fmt"
    "time"
    "sort"

    "github.com/gorilla/mux"
    "github.com/gorilla/sessions"
    "github.com/joho/godotenv"
    "github.com/google/uuid"

    "cloud/internal/auth"
    "cloud/internal/db"
    "cloud/internal/storage"
)

var (
    store *sessions.CookieStore
)

func init() {
    if err := godotenv.Load(); err != nil {
        log.Fatal("Error loading .env file")
    }

    // Initialize auth
    auth.Init()

    // Initialize session store
    store = sessions.NewCookieStore([]byte(os.Getenv("SESSION_SECRET")))

    // Initialize database connection
    if err := db.InitDB(); err != nil {
        log.Fatalf("Failed to initialize database: %v", err)
    }

    // Initialize MinIO storage
    if err := storage.InitStorage(); err != nil {
        log.Fatalf("Failed to initialize MinIO storage: %v", err)
    }
}

type Note struct {
    ID        string    `json:"id"`
    Title     string    `json:"title"`
    Content   string    `json:"content"`
    CreatedAt time.Time `json:"createdAt"`
    UpdatedAt time.Time `json:"updatedAt"`
}

func main() {
    r := mux.NewRouter()

    // Serve static files
    fs := http.FileServer(http.Dir("web/static"))
    r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", fs))

    // Auth routes
    r.HandleFunc("/", handleHome)
    r.HandleFunc("/login", auth.HandleGoogleLogin)
    r.HandleFunc("/auth/google/callback", auth.HandleGoogleCallback)
    r.HandleFunc("/logout", handleLogout)

    // Protected routes
    r.HandleFunc("/dashboard", requireAuth(handleDashboard))
    r.HandleFunc("/upload", requireAuth(handleFileUpload)).Methods("POST")
    r.HandleFunc("/files", requireAuth(handleListFiles)).Methods("GET")
    r.HandleFunc("/files/{filename}", requireAuth(handleDownloadFile)).Methods("GET")
    r.HandleFunc("/files/{filename}/delete", requireAuth(handleDeleteFile)).Methods("DELETE")

    // Note routes
    r.HandleFunc("/notes", requireAuth(handleListNotes)).Methods("GET")
    r.HandleFunc("/notes/create", requireAuth(handleCreateNote)).Methods("POST")
    r.HandleFunc("/notes/{id}", requireAuth(handleGetNote)).Methods("GET")
    r.HandleFunc("/notes/{id}/update", requireAuth(handleUpdateNote)).Methods("PUT")
    r.HandleFunc("/notes/{id}/delete", requireAuth(handleDeleteNote)).Methods("DELETE")

    port := os.Getenv("PORT")
    if port == "" {
        port = "8080"
    }

    log.Printf("Server starting on port %s...", port)
    if err := http.ListenAndServe(":"+port, r); err != nil {
        log.Fatal(err)
    }
}

func requireAuth(next http.HandlerFunc) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        session, _ := store.Get(r, "session")
        if email, ok := session.Values["email"].(string); !ok || email == "" {
            http.Redirect(w, r, "/", http.StatusSeeOther)
            return
        }
        next(w, r)
    }
}

func handleHome(w http.ResponseWriter, r *http.Request) {
    http.ServeFile(w, r, "web/templates/index.html")
}

func handleLogout(w http.ResponseWriter, r *http.Request) {
    session, _ := store.Get(r, "session")
    session.Values["email"] = ""
    session.Values["name"] = ""
    session.Save(r, w)
    http.Redirect(w, r, "/", http.StatusSeeOther)
}

func handleDashboard(w http.ResponseWriter, r *http.Request) {
    http.ServeFile(w, r, "web/templates/dashboard.html")
}

func handleFileUpload(w http.ResponseWriter, r *http.Request) {
    session, _ := store.Get(r, "session")
    email := session.Values["email"].(string)

    // Parse multipart form
    if err := r.ParseMultipartForm(32 << 20); err != nil {
        http.Error(w, "Error parsing form", http.StatusBadRequest)
        return
    }

    file, header, err := r.FormFile("file")
    if err != nil {
        http.Error(w, "Error getting file", http.StatusBadRequest)
        return
    }
    defer file.Close()

    // Create a new file record
    fileID := uuid.New().String()
    fileRecord := db.File{
        UserEmail:    email,
        FileID:       fileID,
        Filename:     header.Filename,
        Size:         header.Size,
        ContentType:  header.Header.Get("Content-Type"),
        StoragePath:  fmt.Sprintf("%s/%s", email, header.Filename),
        UploadedAt:   time.Now(),
    }

    // Upload file to MinIO
    storagePath, err := storage.UploadFile(email, header.Filename, header.Size, file)
    if err != nil {
        http.Error(w, "Error uploading file", http.StatusInternalServerError)
        return
    }
    fileRecord.StoragePath = storagePath

    // Save file metadata to database
    if err := db.SaveFileMetadata(fileRecord); err != nil {
        // Try to delete the uploaded file if metadata save fails
        storage.DeleteFile(email, header.Filename)
        http.Error(w, "Error saving file metadata", http.StatusInternalServerError)
        return
    }

    w.WriteHeader(http.StatusCreated)
    json.NewEncoder(w).Encode(fileRecord)
}

func handleDownloadFile(w http.ResponseWriter, r *http.Request) {
    session, _ := store.Get(r, "session")
    email := session.Values["email"].(string)
    vars := mux.Vars(r)
    filename := vars["filename"]

    // Get file metadata from database
    files, err := db.GetUserFiles(email)
    if err != nil {
        http.Error(w, "Error getting file metadata", http.StatusInternalServerError)
        return
    }

    var fileRecord db.File
    for _, f := range files {
        if f.Filename == filename {
            fileRecord = f
            break
        }
    }

    if fileRecord.FileID == "" {
        http.Error(w, "File not found", http.StatusNotFound)
        return
    }

    // Get file from MinIO
    object, err := storage.DownloadFile(email, filename)
    if err != nil {
        http.Error(w, "Error downloading file", http.StatusInternalServerError)
        return
    }
    defer object.Close()

    // Set response headers
    w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
    w.Header().Set("Content-Type", fileRecord.ContentType)
    w.Header().Set("Content-Length", fmt.Sprintf("%d", fileRecord.Size))

    // Stream file to response
    if _, err := io.Copy(w, object); err != nil {
        log.Printf("Error streaming file: %v", err)
    }
}

func handleDeleteFile(w http.ResponseWriter, r *http.Request) {
    session, _ := store.Get(r, "session")
    email := session.Values["email"].(string)
    vars := mux.Vars(r)
    filename := vars["filename"]

    // Delete file from MinIO
    if err := storage.DeleteFile(email, filename); err != nil {
        http.Error(w, "Error deleting file from storage", http.StatusInternalServerError)
        return
    }

    // Delete file metadata from database
    if err := db.DeleteFile(email, filename); err != nil {
        http.Error(w, "Error deleting file metadata", http.StatusInternalServerError)
        return
    }

    w.WriteHeader(http.StatusOK)
}

func handleListFiles(w http.ResponseWriter, r *http.Request) {
    session, _ := store.Get(r, "session")
    email := session.Values["email"].(string)

    files, err := db.GetUserFiles(email)
    if err != nil {
        http.Error(w, "Error getting files", http.StatusInternalServerError)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(files)
}

func handleCreateNote(w http.ResponseWriter, r *http.Request) {
    session, _ := store.Get(r, "session")
    email := session.Values["email"].(string)

    var note Note
    if err := json.NewDecoder(r.Body).Decode(&note); err != nil {
        http.Error(w, "Invalid request body", http.StatusBadRequest)
        return
    }

    note.ID = uuid.New().String()
    note.CreatedAt = time.Now()
    note.UpdatedAt = time.Now()

    userDir := filepath.Join("notes", email)
    if err := os.MkdirAll(userDir, 0755); err != nil {
        http.Error(w, "Error creating user directory", http.StatusInternalServerError)
        return
    }

    noteFile := filepath.Join(userDir, note.ID+".json")
    noteData, err := json.Marshal(note)
    if err != nil {
        http.Error(w, "Error encoding note", http.StatusInternalServerError)
        return
    }

    if err := ioutil.WriteFile(noteFile, noteData, 0644); err != nil {
        http.Error(w, "Error saving note", http.StatusInternalServerError)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(note)
}

func handleListNotes(w http.ResponseWriter, r *http.Request) {
    session, _ := store.Get(r, "session")
    email := session.Values["email"].(string)

    userDir := filepath.Join("notes", email)
    if err := os.MkdirAll(userDir, 0755); err != nil {
        http.Error(w, "Error accessing user directory", http.StatusInternalServerError)
        return
    }

    files, err := ioutil.ReadDir(userDir)
    if err != nil {
        http.Error(w, "Error reading notes", http.StatusInternalServerError)
        return
    }

    var notes []Note
    for _, file := range files {
        if filepath.Ext(file.Name()) == ".json" {
            noteData, err := ioutil.ReadFile(filepath.Join(userDir, file.Name()))
            if err != nil {
                continue
            }

            var note Note
            if err := json.Unmarshal(noteData, &note); err != nil {
                continue
            }
            notes = append(notes, note)
        }
    }

    // Sort notes by UpdatedAt (most recent first)
    sort.Slice(notes, func(i, j int) bool {
        return notes[i].UpdatedAt.After(notes[j].UpdatedAt)
    })

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(notes)
}

func handleGetNote(w http.ResponseWriter, r *http.Request) {
    session, _ := store.Get(r, "session")
    email := session.Values["email"].(string)
    vars := mux.Vars(r)
    noteID := vars["id"]

    noteFile := filepath.Join("notes", email, noteID+".json")
    if _, err := os.Stat(noteFile); os.IsNotExist(err) {
        http.Error(w, "Note not found", http.StatusNotFound)
        return
    }

    noteData, err := ioutil.ReadFile(noteFile)
    if err != nil {
        http.Error(w, "Error reading note", http.StatusInternalServerError)
        return
    }

    var note Note
    if err := json.Unmarshal(noteData, &note); err != nil {
        http.Error(w, "Error parsing note", http.StatusInternalServerError)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(note)
}

func handleUpdateNote(w http.ResponseWriter, r *http.Request) {
    session, _ := store.Get(r, "session")
    email := session.Values["email"].(string)
    vars := mux.Vars(r)
    noteID := vars["id"]

    var updatedNote Note
    if err := json.NewDecoder(r.Body).Decode(&updatedNote); err != nil {
        http.Error(w, "Invalid request body", http.StatusBadRequest)
        return
    }

    noteFile := filepath.Join("notes", email, noteID+".json")
    if _, err := os.Stat(noteFile); os.IsNotExist(err) {
        http.Error(w, "Note not found", http.StatusNotFound)
        return
    }

    // Read existing note
    noteData, err := ioutil.ReadFile(noteFile)
    if err != nil {
        http.Error(w, "Error reading note", http.StatusInternalServerError)
        return
    }

    var existingNote Note
    if err := json.Unmarshal(noteData, &existingNote); err != nil {
        http.Error(w, "Error parsing note", http.StatusInternalServerError)
        return
    }

    // Update note
    existingNote.Title = updatedNote.Title
    existingNote.Content = updatedNote.Content
    existingNote.UpdatedAt = time.Now()

    // Save updated note
    noteData, err = json.Marshal(existingNote)
    if err != nil {
        http.Error(w, "Error encoding note", http.StatusInternalServerError)
        return
    }

    if err := ioutil.WriteFile(noteFile, noteData, 0644); err != nil {
        http.Error(w, "Error saving note", http.StatusInternalServerError)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(existingNote)
}

func handleDeleteNote(w http.ResponseWriter, r *http.Request) {
    session, _ := store.Get(r, "session")
    email := session.Values["email"].(string)
    vars := mux.Vars(r)
    noteID := vars["id"]

    noteFile := filepath.Join("notes", email, noteID+".json")
    if err := os.Remove(noteFile); err != nil {
        if os.IsNotExist(err) {
            http.Error(w, "Note not found", http.StatusNotFound)
        } else {
            http.Error(w, "Error deleting note", http.StatusInternalServerError)
        }
        return
    }

    w.WriteHeader(http.StatusOK)
}
