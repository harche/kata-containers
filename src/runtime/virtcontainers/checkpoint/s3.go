// Copyright (c) 2026 Red Hat Inc.
//
// SPDX-License-Identifier: Apache-2.0
//

package checkpoint

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// s3Storage implements CheckpointStorage for Amazon S3.
// URIs use the s3:// scheme (e.g. s3://bucket/path/to/checkpoint).
type s3Storage struct {
	client *s3.Client
}

func newS3Storage(ctx context.Context) (CheckpointStorage, error) {
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, fmt.Errorf("load AWS config: %w", err)
	}
	client := s3.NewFromConfig(cfg)
	return &s3Storage{client: client}, nil
}

func (s *s3Storage) Upload(ctx context.Context, localPath, remoteURI string) error {
	bucket, key, err := parseS3URI(remoteURI)
	if err != nil {
		return fmt.Errorf("s3 upload: %w", err)
	}

	f, err := os.Open(localPath)
	if err != nil {
		return fmt.Errorf("s3 upload: open local file %q: %w", localPath, err)
	}
	defer f.Close()

	uploader := manager.NewUploader(s.client)
	_, err = uploader.Upload(ctx, &s3.PutObjectInput{
		Bucket: &bucket,
		Key:    &key,
		Body:   f,
	})
	if err != nil {
		return fmt.Errorf("s3 upload: put s3://%s/%s: %w", bucket, key, err)
	}

	return nil
}

func (s *s3Storage) Download(ctx context.Context, remoteURI, localPath string) error {
	bucket, key, err := parseS3URI(remoteURI)
	if err != nil {
		return fmt.Errorf("s3 download: %w", err)
	}

	f, err := os.Create(localPath)
	if err != nil {
		return fmt.Errorf("s3 download: create local file %q: %w", localPath, err)
	}
	defer f.Close()

	downloader := manager.NewDownloader(s.client)
	_, err = downloader.Download(ctx, f, &s3.GetObjectInput{
		Bucket: &bucket,
		Key:    &key,
	})
	if err != nil {
		return fmt.Errorf("s3 download: get s3://%s/%s: %w", bucket, key, err)
	}

	return nil
}

func parseS3URI(uri string) (bucket, key string, err error) {
	u, err := url.Parse(uri)
	if err != nil {
		return "", "", fmt.Errorf("invalid S3 URI %q: %w", uri, err)
	}
	if u.Scheme != "s3" {
		return "", "", fmt.Errorf("expected s3:// scheme, got %q", u.Scheme)
	}
	bucket = u.Host
	if bucket == "" {
		return "", "", fmt.Errorf("missing bucket in S3 URI %q", uri)
	}
	key = strings.TrimPrefix(u.Path, "/")
	if key == "" {
		return "", "", fmt.Errorf("missing key path in S3 URI %q", uri)
	}
	return bucket, key, nil
}
