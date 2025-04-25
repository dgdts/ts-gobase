package minio

import (
	"context"
	"fmt"
	"io"
	"net/url"
	"sync"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

var (
	instance *minioClient
	once     sync.Once
)

// Config MinIO configuration
type Config struct {
	Endpoint        string `yaml:"endpoint"`
	AccessKeyID     string `yaml:"access_key_id"`
	SecretAccessKey string `yaml:"secret_access_key"`
	UseSSL          bool   `yaml:"use_ssl"`
	BucketName      string `yaml:"bucket_name"`
}

type minioClient struct {
	client     *minio.Client
	bucketName string
}

// Init initializes the global MinIO client
func InitMinioClient(cfg *Config) error {
	var initErr error
	once.Do(func() {
		client, err := minio.New(cfg.Endpoint, &minio.Options{
			Creds:  credentials.NewStaticV4(cfg.AccessKeyID, cfg.SecretAccessKey, ""),
			Secure: cfg.UseSSL,
		})
		if err != nil {
			initErr = fmt.Errorf("failed to initialize MinIO client: %w", err)
			return
		}

		// Ensure bucket exists
		exists, err := client.BucketExists(context.Background(), cfg.BucketName)
		if err != nil {
			initErr = fmt.Errorf("failed to check bucket existence: %w", err)
			return
		}

		if !exists {
			err = client.MakeBucket(context.Background(), cfg.BucketName, minio.MakeBucketOptions{})
			if err != nil {
				initErr = fmt.Errorf("failed to create bucket: %w", err)
				return
			}
		}

		instance = &minioClient{
			client:     client,
			bucketName: cfg.BucketName,
		}
	})
	return initErr
}

// UploadFile uploads a file to MinIO and returns its URL
// expiry parameter specifies how long the URL will be valid
func UploadFile(objectName string, reader io.Reader, contentType string, expiry time.Duration) (string, error) {
	if instance == nil {
		return "", fmt.Errorf("MinIO client not initialized")
	}

	ctx := context.Background()

	// Upload the file
	_, err := instance.client.PutObject(ctx, instance.bucketName, objectName, reader, -1,
		minio.PutObjectOptions{ContentType: contentType})
	if err != nil {
		return "", fmt.Errorf("failed to upload file: %w", err)
	}

	// Get the URL after successful upload
	presignedURL, err := instance.client.PresignedGetObject(ctx, instance.bucketName, objectName, expiry, nil)
	if err != nil {
		return "", fmt.Errorf("file uploaded but failed to get URL: %w", err)
	}

	return presignedURL.String(), nil
}

// DownloadFile downloads a file from MinIO
func DownloadFile(objectName string) (io.Reader, error) {
	if instance == nil {
		return nil, fmt.Errorf("MinIO client not initialized")
	}
	ctx := context.Background()
	obj, err := instance.client.GetObject(ctx, instance.bucketName, objectName, minio.GetObjectOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to download file: %w", err)
	}
	return obj, nil
}

// DeleteFile deletes a file from MinIO
func DeleteFile(objectName string) error {
	if instance == nil {
		return fmt.Errorf("MinIO client not initialized")
	}
	ctx := context.Background()
	err := instance.client.RemoveObject(ctx, instance.bucketName, objectName, minio.RemoveObjectOptions{})
	if err != nil {
		return fmt.Errorf("failed to delete file: %w", err)
	}
	return nil
}

// GetFileURL gets a temporary URL for file access
func GetFileURL(objectName string, expiry time.Duration) (*url.URL, error) {
	if instance == nil {
		return nil, fmt.Errorf("MinIO client not initialized")
	}
	ctx := context.Background()
	presignedURL, err := instance.client.PresignedGetObject(ctx, instance.bucketName, objectName, expiry, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get file URL: %w", err)
	}
	return presignedURL, nil
}

// ListFiles lists all files with the specified prefix
func ListFiles(prefix string) ([]string, error) {
	if instance == nil {
		return nil, fmt.Errorf("MinIO client not initialized")
	}
	ctx := context.Background()
	var files []string

	objects := instance.client.ListObjects(ctx, instance.bucketName, minio.ListObjectsOptions{
		Prefix:    prefix,
		Recursive: true,
	})

	for object := range objects {
		if object.Err != nil {
			return nil, fmt.Errorf("failed to list files: %w", object.Err)
		}
		files = append(files, object.Key)
	}

	return files, nil
}

func GetUploadPresignedURL(objectName string, expiry time.Duration) (*url.URL, error) {
	if instance == nil {
		return nil, fmt.Errorf("MinIO client not initialized")
	}
	ctx := context.Background()
	presignedURL, err := instance.client.PresignedPutObject(ctx, instance.bucketName, objectName, expiry)
	if err != nil {
		return nil, fmt.Errorf("failed to get upload presigned URL: %w", err)
	}
	return presignedURL, nil
}
