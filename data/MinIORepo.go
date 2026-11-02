package data

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"net/http"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type MinIORepo struct {
	client *minio.Client
	bucket string
}

func NewMinIORepo(endpoint, username, password, bucket string) (*MinIORepo, error) {
	client, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(username, password, ""),
		Secure: true,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("Ошибка при создании клиента MinIO: %v", err)
	}

	exists, err := client.BucketExists(context.Background(), bucket)
	if err != nil {
		return nil, fmt.Errorf("Ошибка проверки бакета: %v", err)
	}
	if !exists {
		if err := client.MakeBucket(context.Background(), bucket, minio.MakeBucketOptions{}); err != nil {
			return nil, fmt.Errorf("Ошибка создания бакета '%s': %v", bucket, err)
		}
		fmt.Printf("Бакет '%s' успешно создан\n", bucket)
	}

	return &MinIORepo{
		client: client,
		bucket: bucket,
	}, nil
}

func (r *MinIORepo) Put(ctx context.Context, fileName string, reader io.Reader, size int64) (minio.UploadInfo, error) {
	return r.client.PutObject(
		ctx,
		r.bucket,
		fileName,
		reader,
		size,
		minio.PutObjectOptions{
			ContentType: "application/octet-stream",
		},
	)
}

func (r *MinIORepo) Remove(ctx context.Context, fileName string) error {
	return r.client.RemoveObject(ctx, r.bucket, fileName, minio.RemoveObjectOptions{})
}

func (r *MinIORepo) GetObject(ctx context.Context, objectName string) (io.ReadCloser, error) {
	obj, err := r.client.GetObject(ctx, r.bucket, objectName, minio.GetObjectOptions{})
	if err != nil {
		return nil, err
	}
	return obj, nil
}
