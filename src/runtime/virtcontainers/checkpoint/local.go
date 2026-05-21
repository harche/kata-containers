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
	"path/filepath"
)

// localStorage implements CheckpointStorage for local file copies.
// URIs use the file:// scheme (e.g. file:///tmp/checkpoints/my-vm).
type localStorage struct{}

func newLocalStorage() CheckpointStorage {
	return &localStorage{}
}

func (l *localStorage) Upload(ctx context.Context, localPath, remoteURI string) error {
	dstPath, err := parseFileURI(remoteURI)
	if err != nil {
		return fmt.Errorf("local upload: %w", err)
	}

	if err := os.MkdirAll(filepath.Dir(dstPath), 0700); err != nil {
		return fmt.Errorf("local upload: create directory: %w", err)
	}

	return copyFile(localPath, dstPath)
}

func (l *localStorage) Download(ctx context.Context, remoteURI, localPath string) error {
	srcPath, err := parseFileURI(remoteURI)
	if err != nil {
		return fmt.Errorf("local download: %w", err)
	}

	if err := os.MkdirAll(filepath.Dir(localPath), 0700); err != nil {
		return fmt.Errorf("local download: create directory: %w", err)
	}

	return copyFile(srcPath, localPath)
}

func parseFileURI(uri string) (string, error) {
	u, err := url.Parse(uri)
	if err != nil {
		return "", fmt.Errorf("invalid file URI %q: %w", uri, err)
	}
	if u.Scheme != "file" {
		return "", fmt.Errorf("expected file:// scheme, got %q", u.Scheme)
	}
	return u.Path, nil
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("open source %q: %w", src, err)
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return fmt.Errorf("create destination %q: %w", dst, err)
	}
	defer func() {
		if cerr := out.Close(); err == nil {
			err = cerr
		}
	}()

	if _, err = io.Copy(out, in); err != nil {
		return fmt.Errorf("copy %q -> %q: %w", src, dst, err)
	}

	return nil
}
