// Package s3 wraps an S3-compatible object store (MinIO locally, real S3 in prod)
// used as the async archive for raw OTLP payloads.
package s3

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"

	"error-logging/pkg/config"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type Client struct {
	mc     *minio.Client
	bucket string
}

// NewClient builds the object-store client. The MinIO SDK connects lazily, so
// construction only fails on bad configuration, not on the store being down —
// connectivity surfaces on the first Put/Get.
func NewClient(cfg config.S3Config) (*Client, error) {
	mc, err := minio.New(cfg.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.AccessKey, cfg.SecretKey, ""),
		Secure: cfg.UseSSL,
		Region: cfg.Region,
	})
	if err != nil {
		return nil, fmt.Errorf("init object store client: %w", err)
	}

	log.Println("S3/MinIO client initialized successfully")
	return &Client{mc: mc, bucket: cfg.Bucket}, nil
}

// EnsureBucket creates the configured bucket if it does not already exist.
func (c *Client) EnsureBucket(ctx context.Context) error {
	exists, err := c.mc.BucketExists(ctx, c.bucket)
	if err != nil {
		return fmt.Errorf("check bucket %q: %w", c.bucket, err)
	}
	if exists {
		return nil
	}
	if err := c.mc.MakeBucket(ctx, c.bucket, minio.MakeBucketOptions{}); err != nil {
		return fmt.Errorf("create bucket %q: %w", c.bucket, err)
	}
	return nil
}

// Put stores an object under key.
func (c *Client) Put(ctx context.Context, key string, data []byte, contentType string) error {
	_, err := c.mc.PutObject(ctx, c.bucket, key, bytes.NewReader(data), int64(len(data)),
		minio.PutObjectOptions{ContentType: contentType})
	if err != nil {
		return fmt.Errorf("put object %q: %w", key, err)
	}
	return nil
}

// Get retrieves an object by key.
func (c *Client) Get(ctx context.Context, key string) ([]byte, error) {
	obj, err := c.mc.GetObject(ctx, c.bucket, key, minio.GetObjectOptions{})
	if err != nil {
		return nil, fmt.Errorf("get object %q: %w", key, err)
	}
	defer obj.Close()

	data, err := io.ReadAll(obj)
	if err != nil {
		return nil, fmt.Errorf("read object %q: %w", key, err)
	}
	return data, nil
}

// Close is a no-op (the MinIO SDK has no explicit close) but satisfies io.Closer
// so the store can be managed alongside the other clients on shutdown.
func (c *Client) Close() error { return nil }
