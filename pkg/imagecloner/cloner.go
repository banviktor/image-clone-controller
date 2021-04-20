package imagecloner

import (
	"context"
	"errors"
)

var (
	// ErrImageAlreadyCloned The input image is already part of target.
	ErrImageAlreadyCloned = errors.New("input image already part of target")
)

// Cloner handles cloning of indexes/images to a target repository prefix.
type Cloner interface {
	// Clone copies an index/image to the target and returns the cloned resource's full name.
	Clone(context.Context, string) (string, error)

	// CloneMulti copies a list of images to the target and returns a map of the new image names keyed by their original
	// counterparts.
	CloneMulti(context.Context, []string) (map[string]string, error)
}
