// Copyright (c) 2026 Red Hat Inc.
//
// SPDX-License-Identifier: Apache-2.0
//

package checkpoint

import (
	"context"
	"fmt"
	"io"
	"net/url"
	"os"
	"strings"

	"cloud.google.com/go/storage"
)

// gcsStorage implements CheckpointStorage for Google Cloud Storage.
// URIs use the gs:// scheme (e.g. gs://bucket/path/to/checkpoint).
type gcsStorage struct {
	client *storage.Client
}

func newGCSStorage(ctx context.Context) (CheckpointStorage, error) {
	client, err := storage.NewClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("create GCS client: %w", err)
	}
	return &gcsStorage{client: client}, nil
}

func (g *gcsStorage) Upload(ctx context.Context, localPath, remoteURI string) error {
	bucket, object, err := parseGCSURI(remoteURI)
	if err != nil {
		return fmt.Errorf("gcs upload: %w", err)
	}

	f, err := os.Open(localPath)
	if err != nil {
		return fmt.Errorf("gcs upload: open local file %q: %w", localPath, err)
	}
	defer f.Close()

	w := g.client.Bucket(bucket).Object(object).NewWriter(ctx)
	if _, err := io.Copy(w, f); err != nil {
		w.Close()
		return fmt.Errorf("gcs upload: copy to gs://%s/%s: %w", bucket, object, err)
	}
	if err := w.Close(); err != nil {
		return fmt.Errorf("gcs upload: finalize gs://%s/%s: %w", bucket, object, err)
	}

	return nil
}

func (g *gcsStorage) Download(ctx context.Context, remoteURI, localPath string) error {
	bucket, object, err := parseGCSURI(remoteURI)
	if err != nil {
		return fmt.Errorf("gcs download: %w", err)
	}

	r, err := g.client.Bucket(bucket).Object(object).NewReader(ctx)
	if err != nil {
		return fmt.Errorf("gcs download: read gs://%s/%s: %w", bucket, object, err)
	}
	defer r.Close()

	f, err := os.Create(localPath)
	if err != nil {
		return fmt.Errorf("gcs download: create local file %q: %w", localPath, err)
	}
	defer f.Close()

	if _, err := io.Copy(f, r); err != nil {
		return fmt.Errorf("gcs download: copy from gs://%s/%s: %w", bucket, object, err)
	}

	return nil
}

func parseGCSURI(uri string) (bucket, object string, err error) {
	u, err := url.Parse(uri)
	if err != nil {
		return "", "", fmt.Errorf("invalid GCS URI %q: %w", uri, err)
	}
	if u.Scheme != "gs" {
		return "", "", fmt.Errorf("expected gs:// scheme, got %q", u.Scheme)
	}
	bucket = u.Host
	if bucket == "" {
		return "", "", fmt.Errorf("missing bucket in GCS URI %q", uri)
	}
	object = strings.TrimPrefix(u.Path, "/")
	if object == "" {
		return "", "", fmt.Errorf("missing object path in GCS URI %q", uri)
	}
	return bucket, object, nil
}
