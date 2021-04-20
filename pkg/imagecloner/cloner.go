package imagecloner

import (
	"context"
	"errors"
	"fmt"
	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/v1/remote"
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

func clone(ctx context.Context, source, destination name.Reference) error {
	options := []remote.Option{
		remote.WithContext(ctx),
		remote.WithAuthFromKeychain(authn.DefaultKeychain),
	}

	sourceDescriptor, err := remote.Get(source, options...)
	if err != nil {
		return fmt.Errorf("failed to fetch source resource: %w", err)
	}

	// Prefer cloning the whole index if available.
	index, _ := sourceDescriptor.ImageIndex()
	if index != nil {
		if err := remote.WriteIndex(destination, index, options...); err != nil {
			return fmt.Errorf("failed to clone index: %w", err)
		}
		return nil
	}

	// Fallback to cloning a singular image.
	image, err := sourceDescriptor.Image()
	if err != nil {
		return fmt.Errorf("unsupported resource: %w", err)
	}
	if err := remote.Write(destination, image, options...); err != nil {
		return fmt.Errorf("failed to clone image: %w", err)
	}
	return nil
}
