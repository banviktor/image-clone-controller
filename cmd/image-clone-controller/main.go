package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/banviktor/image-clone-controller/pkg/controller"
	"github.com/banviktor/image-clone-controller/pkg/imagecloner"
	dockerconfig "github.com/docker/cli/cli/config"
	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/name"
	"os"
	controllerconfig "sigs.k8s.io/controller-runtime/pkg/client/config"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

func init() {
	log.SetLogger(zap.New())
}

func main() {
	mgr, err := manager.New(controllerconfig.GetConfigOrDie(), manager.Options{})
	if err != nil {
		log.Log.Error(err, "unable to set up controller manager")
		os.Exit(1)
	}

	targetRepository, err := getTargetRepositoryPrefix()
	if err != nil {
		log.Log.Error(err, "unable to determine target repository")
		os.Exit(1)
	}
	log.Log.Info(fmt.Sprintf("using target repository: %s", targetRepository))

	cloner, err := imagecloner.NewFlatCloner(targetRepository)
	if err != nil {
		log.Log.Error(err, "unable to initialize image cloner")
		os.Exit(1)
	}

	if err := controller.AttachController("deployment", mgr, &controller.DeploymentManager{}, cloner); err != nil {
		log.Log.Error(err, "unable to set up Deployment controller")
		os.Exit(1)
	}

	if err := controller.AttachController("daemonset", mgr, &controller.DaemonSetManager{}, cloner); err != nil {
		log.Log.Error(err, "unable to set up DaemonSet controller")
		os.Exit(1)
	}

	if err := mgr.Start(context.Background()); err != nil {
		log.Log.Error(err, "unable to start controller manager")
		os.Exit(1)
	}
}

func getTargetRepositoryPrefix() (string, error) {
	cf, err := dockerconfig.Load(os.Getenv("DOCKER_CONFIG"))
	if err != nil {
		return "", err
	}

	for server, authConfig := range cf.AuthConfigs {
		if server == authn.DefaultAuthKey {
			server = name.DefaultRegistry
		}
		return server + "/" + authConfig.Username, nil
	}
	return "", errors.New("unable to determine target repository")
}
