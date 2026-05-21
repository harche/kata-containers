// Copyright (c) 2026 Red Hat Inc.
//
// SPDX-License-Identifier: Apache-2.0
//

package checkpoint

import (
	"context"
	"fmt"
	"net/url"
)

// NewStorage returns a CheckpointStorage backend appropriate for the given URI scheme.
// Supported schemes:
//   - gs://   -> Google Cloud Storage
//   - s3://   -> Amazon S3
//   - file:// -> Local filesystem copy
func NewStorage(ctx context.Context, uri string) (CheckpointStorage, error) {
	u, err := url.Parse(uri)
	if err != nil {
		return nil, fmt.Errorf("parse checkpoint URI %q: %w", uri, err)
	}

	switch u.Scheme {
	case "gs":
		return newGCSStorage(ctx)
	case "s3":
		return newS3Storage(ctx)
	case "file":
		return newLocalStorage(), nil
	default:
		return nil, fmt.Errorf("unsupported checkpoint storage scheme %q (from URI %q); supported: gs://, s3://, file://", u.Scheme, uri)
	}
}
