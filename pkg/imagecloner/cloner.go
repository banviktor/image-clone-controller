package imagecloner

import (
	"context"
	"errors"
	"fmt"
	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"strings"
)

var (
	// ErrInputHasTargetPrefix The input image already has the target prefix.
	ErrInputHasTargetPrefix = errors.New("input image already has target prefix")
)

// Cloner handles cloning of indexes/images to a target repository prefix.
type Cloner struct {
	targetPrefix string
}

// NewCloner creates a Cloner with a target repository prefix that may contain implicit parts.
func NewCloner(targetPrefix string) (*Cloner, error) {
	targetPrefix, err := expandRepositoryPrefix(targetPrefix)
	if err != nil {
		return nil, err
	}

	return &Cloner{targetPrefix: targetPrefix}, nil
}

// Clone copies an index/image to the target and returns the cloned resource's full name.
func (c *Cloner) Clone(ctx context.Context, image string) (string, error) {
	sourceReference, err := name.ParseReference(image)
	if err != nil {
		return "", err
	}

	if strings.HasPrefix(sourceReference.Name(), c.targetPrefix) {
		return "", ErrInputHasTargetPrefix
	}

	backupReference := c.buildBackupReference(sourceReference)
	if err := clone(ctx, sourceReference, backupReference); err != nil {
		return "", err
	}

	return backupReference.Name(), nil
}

// CloneMulti copies a list of images to the target and returns a map of the new image names keyed by their original
// counterparts.
//
// When an already cloned image is encountered, it will not be included in the returned map.
func (c *Cloner) CloneMulti(ctx context.Context, images []string) (map[string]string, error) {
	ch := make(chan [2]string)
	errCh := make(chan error)
	defer close(ch)
	defer close(errCh)

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	for _, image := range images {
		go func(image string) {
			newImage, err := c.Clone(ctx, image)
			if err != nil {
				errCh <- err
				return
			}
			ch <- [2]string{image, newImage}
		}(image)
	}

	imageMap := make(map[string]string)
	for i := 0; i < len(images); i++ {
		select {
		case err := <-errCh:
			if err == ErrInputHasTargetPrefix {
				continue
			}
			return nil, err
		case pair := <-ch:
			imageMap[pair[0]] = pair[1]
		}
	}
	return imageMap, nil
}

func (c *Cloner) buildBackupReference(source name.Reference) name.Reference {
	backupName := c.targetPrefix
	backupName += "/" + strings.Replace(source.Context().RegistryStr(), ":", "_", 1)
	backupName += "_" + strings.ReplaceAll(source.Context().RepositoryStr(), "/", "_")
	if tag, ok := source.(name.Tag); ok {
		backupName += ":" + tag.TagStr()
	}
	if digest, ok := source.(name.Digest); ok {
		backupName += "@" + digest.DigestStr()
	}

	ref, err := name.ParseReference(backupName)
	if err != nil {
		panic(err)
	}
	return ref
}

func expandRepositoryPrefix(original string) (string, error) {
	newPrefix := strings.TrimRight(original, "/")
	fakeReference, err := name.ParseReference(newPrefix + "/foo")
	if err != nil {
		return "", err
	}
	newPrefix = strings.TrimSuffix(fakeReference.Name(), "/foo:latest")

	return newPrefix, nil
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
