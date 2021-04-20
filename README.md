# image-clone-controller

**Not ready for production!**

A Kubernetes controller for automatically backing up public images used by
Deployments and DaemonSets to a dedicated backup registry repository.

## Quick start

Create namespace:

    kubectl create namespace image-clone-controller

Create Secret for storing target registry credentials (use Docker Hub):

    kubectl -n image-clone-controller create secret docker-registry target-registry \
            --docker-server=DOCKER_REGISTRY_SERVER \
            --docker-username=DOCKER_USER \
            --docker-password=DOCKER_PASSWORD

Build the controller and load into KinD:

    docker build -t banviktor/image-clone-controller:main .
    kind load docker-image banviktor/image-clone-controller:main

Deploy the application:

    kubectl -n image-clone-controller apply -k ./deployment/overlays/development

## Run tests

There's room for improvement when it comes to testing, however I've included a 
fairly simple integration test. It's better than nothing.

    docker run --rm -d --name registry -p 5000:5000 registry:2
    go test ./...

## Design considerations

### Separate controllers for the Deployments and DaemonSets

It is [recommended by the controller-runtime project](https://github.com/kubernetes-sigs/controller-runtime/blob/master/FAQ.md#q-how-do-i-know-which-type-of-object-a-controller-references)
that each controller should only reconcile one object type. Deployments and 
DaemonSets are very similar in terms of where the container images are in their
specification, which makes the controllers very similar. To avoid code
duplication I introduced the ObjectManager interface, which abstracts away the
operations the controller needs to make on the concrete objects:

  - Create new concrete object
  - Get container images
  - Replace container images

This approach allows easy expandability to allow handling more kinds of 
resources.

### Package for handling image cloning

The image cloning logic is fairly self-contained, and it made sense to extract 
it into its own package: `imagecloner`.

The logic prefers cloning whole indexes (if available) over singular images to
reduce architectural dependencies.

## Improvement ideas

### Queuing long-running cloning processes

Currently, the reconciliation callback blocks until all the images are cloned,
and patches changes after the cloning is done. This issue is somewhat mitigated
because of the following:

  - a strategic merge patch is used to update resources
  - cloning to a well-known registry usually ends up being a task of copying 
    only metadata, thus it's fairly fast

However, this could be further improved using a multi-step process:

  1. Queue images to be cloned and annotate object with "pending" state.
  2. Requeue "pending" resources until cloning is complete.
  3. If the resource's images have not changed, and the cloning is complete, do 
     the patch.

### Respect `imagePullPolicy`

The way the controller works the copied images become static. Upstream changes
to `:latest` or commonly used minor version tags (`:0.1`) will not be cloned to
the backup registry and used by applications.

Upstream image sources could be stored in annotations of the resource
originally referencing them, and a periodic job could ensure cloning of new 
images.

### Support for private target repositories

When using Docker Hub as a target registry every created repository 
automatically becomes public, so there's no need for dealing with 
`imagePullSecrets`. However, quay.io creates private repositories, so it cannot
be used with this solution as is.

Since Secrets are namespace-scoped, the solution would require making sure the
backup registry credentials are cloned to each namespace where they will be
referenced as `imagePullSecrets`, or attached to the default ServiceAccount.
