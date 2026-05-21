// Copyright (c) 2026 Red Hat Inc.
//
// SPDX-License-Identifier: Apache-2.0
//

package checkpoint

import (
	"context"
)

// CheckpointStorage defines the interface for uploading and downloading
// VM checkpoint artifacts to/from remote storage backends.
type CheckpointStorage interface {
	// Upload copies a local file at localPath to the given remoteURI.
	Upload(ctx context.Context, localPath, remoteURI string) error

	// Download copies a remote object at remoteURI to the given localPath.
	Download(ctx context.Context, remoteURI, localPath string) error
}
