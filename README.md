
# Personal Cloud Storage

A secure and user-friendly cloud storage application built with Go and modern web technologies.

## Features

- User Authentication (Register/Login)
- File Upload/Download
- File Management (List/Delete)
- Secure File Storage
- Modern Web Interface
- SQLite Database for User and File Metadata
- JWT-based Authentication

## Prerequisites

- Go 1.21 or later
- SQLite3

## Installation

1. Clone the repository:
```bash
git clone https://github.com/yourusername/cloud.git
cd cloud
```

2. Install dependencies:
```bash
go mod download
```

3. Copy the environment file and configure it:
```bash
cp .env.example .env
```

Edit `.env` and set your preferred values for:
- `PORT`: Server port (default: 3000)
- `UPLOAD_DIR`: Directory for file storage (default: uploads)
- `JWT_SECRET`: Secret key for JWT tokens

## Running the Application

1. Start the server:
```bash
go run cmd/server/main.go
```

2. Open your browser and navigate to:
```
http://localhost:3000
```

## API Endpoints

### Authentication
- `POST /auth/register`: Register a new user
- `POST /auth/login`: Login and get JWT token

### File Management (Protected Routes)
- `POST /upload`: Upload a file
- `GET /download/{filename}`: Download a file
- `GET /files`: List all files
- `DELETE /delete/{filename}`: Delete a file

## Security Features

- Password hashing using bcrypt
- JWT-based authentication
- File path sanitization
- CORS protection
- User-specific file access

## Project Structure

```
.
├── cmd/
│   └── server/
│       └── main.go
├── internal/
│   ├── auth/
│   │   └── auth.go
│   ├── database/
│   │   └── database.go
│   └── storage/
│       └── storage.go
├── static/
│   └── index.html
├── uploads/
├── .env
├── .env.example
├── go.mod
└── README.md
```

## Contributing

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request


