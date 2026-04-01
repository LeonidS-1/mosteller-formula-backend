package minio

import (
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

func NewMinioClient(endpoint, accessKey, secretKey string, useSSL bool) (*minio.Client, error) {
	return minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure: useSSL,
	})
}

func UploadImage(ctx context.Context, client *minio.Client, bucket string, file *multipart.FileHeader, drugID uint) (string, error) {
	f, err := file.Open()
	if err != nil {
		return "", err
	}
	defer f.Close()

	contentType := file.Header.Get("Content-Type")
	ext := filepath.Ext(file.Filename)
	if ext == "" {
		switch contentType {
		case "image/jpeg":
			ext = ".jpg"
		case "image/png":
			ext = ".png"
		case "image/gif":
			ext = ".gif"
		default:
			ext = ".bin"
		}
	}
	objectName := fmt.Sprintf("drug_%d%s", drugID, ext)

	_, err = UploadFromReader(ctx, client, bucket, objectName, f, file.Size, contentType)
	if err != nil {
		return "", err
	}
	return objectName, nil
}

func UploadVideo(ctx context.Context, client *minio.Client, bucket string, file *multipart.FileHeader, drugID uint) (string, error) {
	f, err := file.Open()
	if err != nil {
		return "", err
	}
	defer f.Close()

	contentType := file.Header.Get("Content-Type")
	ext := filepath.Ext(file.Filename)
	if ext == "" {
		switch contentType {
		case "video/mp4":
			ext = ".mp4"
		case "video/webm":
			ext = ".webm"
		default:
			ext = ".mp4"
		}
	}
	objectName := fmt.Sprintf("drug_%d_video%s", drugID, ext)

	_, err = UploadFromReader(ctx, client, bucket, objectName, f, file.Size, contentType)
	if err != nil {
		return "", err
	}
	return objectName, nil
}

func UploadFromReader(ctx context.Context, client *minio.Client, bucket, objectName string, r io.Reader, size int64, contentType string) (minio.UploadInfo, error) {
	return client.PutObject(ctx, bucket, objectName, r, size, minio.PutObjectOptions{
		ContentType: contentType,
	})
}

func DeleteObject(ctx context.Context, client *minio.Client, bucket, objectName string) error {
	return client.RemoveObject(ctx, bucket, objectName, minio.RemoveObjectOptions{})
}

func InitMinio() (*minio.Client, error) {
	host := os.Getenv("MINIO_HOST")
	if host == "" {
		host = "localhost"
	}
	port := os.Getenv("MINIO_PORT")
	if port == "" {
		port = "9000"
	}
	user := os.Getenv("MINIO_USER")
	if user == "" {
		user = "minio"
	}
	pass := os.Getenv("MINIO_PASS")
	if pass == "" {
		pass = "minio124"
	}
	return NewMinioClient(host+":"+port, user, pass, false)
}

func GetMediaBucket() string {
	b := os.Getenv("MINIO_BUCKET")
	if b == "" {
		return "inv-media"
	}
	return b
}
