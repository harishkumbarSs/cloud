package storage

import (
    "context"
    "fmt"
    "io"
    "log"
    "os"

    "github.com/minio/minio-go/v7"
    "github.com/minio/minio-go/v7/pkg/credentials"
)

var minioClient *minio.Client

func InitStorage() error {
    endpoint := os.Getenv("MINIO_ENDPOINT")
    accessKeyID := os.Getenv("MINIO_ACCESS_KEY")
    secretAccessKey := os.Getenv("MINIO_SECRET_KEY")
    useSSL := false

    var err error
    minioClient, err = minio.New(endpoint, &minio.Options{
        Creds:  credentials.NewStaticV4(accessKeyID, secretAccessKey, ""),
        Secure: useSSL,
    })
    if err != nil {
        return fmt.Errorf("failed to connect to MinIO: %v", err)
    }

    // Create bucket if it doesn't exist
    bucketName := "cloud-storage"
    exists, err := minioClient.BucketExists(context.Background(), bucketName)
    if err != nil {
        return fmt.Errorf("failed to check bucket: %v", err)
    }

    if !exists {
        err = minioClient.MakeBucket(context.Background(), bucketName, minio.MakeBucketOptions{})
        if err != nil {
            return fmt.Errorf("failed to create bucket: %v", err)
        }
        log.Printf("Created bucket: %s", bucketName)
    }

    return nil
}

func UploadFile(userEmail, fileName string, fileSize int64, reader io.Reader) (string, error) {
    bucketName := "cloud-storage"
    objectName := fmt.Sprintf("%s/%s", userEmail, fileName)

    _, err := minioClient.PutObject(context.Background(), bucketName, objectName, reader, fileSize, minio.PutObjectOptions{})
    if err != nil {
        return "", fmt.Errorf("failed to upload file: %v", err)
    }

    return objectName, nil
}

func DownloadFile(userEmail, fileName string) (*minio.Object, error) {
    bucketName := "cloud-storage"
    objectName := fmt.Sprintf("%s/%s", userEmail, fileName)

    object, err := minioClient.GetObject(context.Background(), bucketName, objectName, minio.GetObjectOptions{})
    if err != nil {
        return nil, fmt.Errorf("failed to download file: %v", err)
    }

    return object, nil
}

func DeleteFile(userEmail, fileName string) error {
    bucketName := "cloud-storage"
    objectName := fmt.Sprintf("%s/%s", userEmail, fileName)

    err := minioClient.RemoveObject(context.Background(), bucketName, objectName, minio.RemoveObjectOptions{})
    if err != nil {
        return fmt.Errorf("failed to delete file: %v", err)
    }

    return nil
}
