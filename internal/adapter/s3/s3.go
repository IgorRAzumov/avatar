package s3

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"

	"avatar/internal/adapter/objectkey"
	"avatar/internal/circuitbreaker"
	"avatar/internal/config"
	"avatar/internal/domain/model"
	"avatar/internal/observability"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

const breakerName = "s3"

type Storage struct {
	client  *minio.Client
	bucket  string
	kit     observability.Kit
	breaker *circuitbreaker.Breaker
}

func NewStorage(cfg config.S3Config, kit observability.Kit) (*Storage, error) {
	client, err := minio.New(cfg.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.AccessKey, cfg.SecretKey, ""),
		Secure: cfg.UseSSL,
		Region: cfg.Region,
	})
	if err != nil {
		return nil, fmt.Errorf("create s3 client: %w", err)
	}

	return &Storage{
		client: client,
		bucket: cfg.Bucket,
		kit:    kit,
		breaker: circuitbreaker.New(circuitbreaker.Options{
			Name: breakerName,
			IsSuccessful: func(err error) bool {
				return err == nil || errors.Is(err, model.ErrNotFound)
			},
			OnStateChange: circuitbreaker.ReportStateTo(kit.Metrics()),
		}),
	}, nil
}

func (storage *Storage) upload(ctx context.Context, key string, data []byte, contentType string) error {
	return storage.withBreaker(ctx, "s3.put_object", "put_object", func(ctx context.Context) error {
		opts := minio.PutObjectOptions{}
		if contentType != "" {
			opts.ContentType = contentType
		}

		_, err := storage.client.PutObject(
			ctx,
			storage.bucket,
			key,
			bytes.NewReader(data),
			int64(len(data)),
			opts,
		)
		if err != nil {
			return fmt.Errorf("upload object: %w", err)
		}
		return nil
	}, observability.String("s3.key", key), observability.Int("s3.size_bytes", len(data)))
}

func (storage *Storage) download(ctx context.Context, key string) ([]byte, error) {
	var result []byte
	err := storage.withBreaker(ctx, "s3.get_object", "get_object", func(ctx context.Context) error {
		obj, err := storage.client.GetObject(ctx, storage.bucket, key, minio.GetObjectOptions{})
		if err != nil {
			return fmt.Errorf("get object: %w", err)
		}
		defer func() { _ = obj.Close() }()

		if _, err := obj.Stat(); err != nil {
			if isNotFound(err) {
				return fmt.Errorf("%w", model.ErrNotFound)
			}
			return fmt.Errorf("stat object: %w", err)
		}

		data, err := io.ReadAll(obj)
		if err != nil {
			return fmt.Errorf("read object: %w", err)
		}
		result = data
		return nil
	}, observability.String("s3.key", key))
	return result, err
}

func (storage *Storage) withBreaker(
	ctx context.Context,
	spanName, operation string,
	fn func(context.Context) error,
	attrs ...observability.Attr,
) error {
	err := storage.breaker.Do(func() error {
		return storage.withSpan(ctx, spanName, operation, fn, attrs...)
	})
	if errors.Is(err, circuitbreaker.ErrOpen) {
		return fmt.Errorf("%w: %s", model.ErrUnavailable, breakerName)
	}
	return err
}

func (storage *Storage) withSpan(
	ctx context.Context,
	spanName, operation string,
	fn func(context.Context) error,
	attrs ...observability.Attr,
) error {
	return storage.kit.RunS3(ctx, spanName, operation, fn, attrs...)
}

func isNotFound(err error) bool {
	resp := minio.ToErrorResponse(err)
	return resp.Code == "NoSuchKey" || resp.Code == "NotFound"
}

func (storage *Storage) SaveOriginal(ctx context.Context, avatarID string, data []byte, contentType string) error {
	return storage.upload(ctx, objectkey.OriginalObjectKey(avatarID), data, contentType)
}

func (storage *Storage) SaveThumbnail(ctx context.Context, avatarID, size string, data []byte) error {
	return storage.upload(ctx, objectkey.ThumbnailObjectKey(avatarID, size), data, "image/jpeg")
}

func (storage *Storage) OpenOriginal(ctx context.Context, avatarID string) ([]byte, error) {
	return storage.download(ctx, objectkey.OriginalObjectKey(avatarID))
}

func (storage *Storage) OpenThumbnail(ctx context.Context, avatarID, size string) ([]byte, error) {
	return storage.download(ctx, objectkey.ThumbnailObjectKey(avatarID, size))
}

func (storage *Storage) DeleteAll(ctx context.Context, avatarID string) error {
	return storage.withBreaker(ctx, "s3.delete_all", "delete_all", func(ctx context.Context) error {
		prefix := objectkey.AvatarDir(avatarID) + "/"
		for obj := range storage.client.ListObjects(ctx, storage.bucket, minio.ListObjectsOptions{
			Prefix:    prefix,
			Recursive: true,
		}) {
			if obj.Err != nil {
				return fmt.Errorf("list objects: %w", obj.Err)
			}
			if err := storage.client.RemoveObject(ctx, storage.bucket, obj.Key, minio.RemoveObjectOptions{}); err != nil {
				return fmt.Errorf("remove object: %w", err)
			}
		}
		return nil
	}, observability.String("avatar_id", avatarID))
}

func (storage *Storage) Ping(ctx context.Context) error {
	return storage.withSpan(ctx, "s3.ping", "ping", func(ctx context.Context) error {
		exists, err := storage.client.BucketExists(ctx, storage.bucket)
		if err != nil {
			return fmt.Errorf("check bucket: %w", err)
		}
		if !exists {
			return fmt.Errorf("bucket %q not found", storage.bucket)
		}
		return nil
	})
}
