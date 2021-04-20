package imagecloner

import (
	"context"
	"github.com/google/go-containerregistry/pkg/name"
	"strings"
)

type flatCloner struct {
	targetPrefix string
}

// NewFlatCloner creates a Cloner that flattens input repository structure into a single level on
// the target side.
func NewFlatCloner(targetPrefix string) (*flatCloner, error) {
	targetPrefix, err := expandRepositoryPrefix(targetPrefix)
	if err != nil {
		return nil, err
	}

	return &flatCloner{targetPrefix: targetPrefix}, nil
}

// Clone implements Cloner.
func (c *flatCloner) Clone(ctx context.Context, image string) (string, error) {
	sourceReference, err := name.ParseReference(image)
	if err != nil {
		return "", err
	}

	if strings.HasPrefix(sourceReference.Name(), c.targetPrefix) {
		return "", ErrImageAlreadyCloned
	}

	backupReference := c.buildBackupReference(sourceReference)
	if err := clone(ctx, sourceReference, backupReference); err != nil {
		return "", err
	}

	return backupReference.Name(), nil
}

// CloneMulti implements Cloner.
func (c *flatCloner) CloneMulti(ctx context.Context, images []string) (map[string]string, error) {
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
			if err == ErrImageAlreadyCloned {
				continue
			}
			return nil, err
		case pair := <-ch:
			imageMap[pair[0]] = pair[1]
		}
	}
	return imageMap, nil
}

func (c *flatCloner) buildBackupReference(source name.Reference) name.Reference {
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
