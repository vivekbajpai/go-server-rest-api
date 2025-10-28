package main

import (
	"bytes"
	"context"
	"fmt"
	"mime"
	"net/url"
	"path/filepath"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

// UploadToIBMCOS uploads the provided data bytes to IBM Cloud Object Storage (S3-compatible).
// It uses the minio v7 client and will create the bucket if it doesn't exist.
func UploadToIBMCOS(ctx context.Context, endpoint, accessKey, secretKey, bucketName, objectName string, data []byte, useSSL bool) error {
	objectName = filepath.Base(objectName)

	// normalize endpoint: minio.New expects host without scheme
	cleaned := endpoint
	if u, err := url.Parse(endpoint); err == nil && u.Scheme != "" {
		cleaned = u.Host
	}

	client, err := minio.New(cleaned, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure: useSSL,
	})
	if err != nil {
		return fmt.Errorf("create minio client: %w", err)
	}

	// ensure bucket exists
	exists, err := client.BucketExists(ctx, bucketName)
	if err != nil {
		return fmt.Errorf("check bucket exists: %w", err)
	}
	if !exists {
		if err := client.MakeBucket(ctx, bucketName, minio.MakeBucketOptions{}); err != nil {
			return fmt.Errorf("create bucket: %w", err)
		}
	}

	// determine content-type
	contentType := "application/octet-stream"
	if ext := filepath.Ext(objectName); ext != "" {
		if ct := mime.TypeByExtension(ext); ct != "" {
			contentType = ct
		}
	}

	r := bytes.NewReader(data)
	_, err = client.PutObject(ctx, bucketName, objectName, r, int64(len(data)), minio.PutObjectOptions{ContentType: contentType})
	if err != nil {
		return fmt.Errorf("put object: %w", err)
	}
	return nil
}
