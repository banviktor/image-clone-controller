package controller

import (
	"context"
	"fmt"
	"github.com/banviktor/image-clone-controller/pkg/imagecloner"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

// Reconciler is a reconcile.Reconciler that backs up images/indexes referenced by Kubernetes resources.
type Reconciler struct {
	client client.Client
	om     ObjectManager
	cloner imagecloner.Cloner
}

// AttachController creates a controller and attaches it to the provided manager.Manager.
func AttachController(name string, mgr manager.Manager, om ObjectManager, targetPrefix string) error {
	cloner, err := imagecloner.NewFlatCloner(targetPrefix)
	if err != nil {
		return err
	}

	c, err := controller.New(name, mgr, controller.Options{
		Reconciler: &Reconciler{
			client: mgr.GetClient(),
			om:     om,
			cloner: cloner,
		},
	})
	if err != nil {
		return err
	}

	err = c.Watch(&source.Kind{Type: om.Create()}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	return nil
}

// Reconcile implements reconcile.Reconciler.
func (r *Reconciler) Reconcile(ctx context.Context, req reconcile.Request) (reconcile.Result, error) {
	log := log.FromContext(ctx)

	// Ignore system and own resources.
	if req.Namespace == "kube-system" || req.Namespace == "image-clone-controller" {
		return reconcile.Result{}, nil
	}

	// Fetch resource.
	object := r.om.Create()
	err := r.client.Get(ctx, req.NamespacedName, object)
	if err != nil {
		log.Error(err, "failed to get resource")
		return reconcile.Result{}, err
	}

	// Clone images.
	imageOverrides, err := r.cloner.CloneMulti(ctx, r.om.GetContainerImages(object))
	if err != nil {
		log.Error(err, "failed to clone images")
		return reconcile.Result{}, err
	}
	if len(imageOverrides) == 0 {
		return reconcile.Result{}, nil
	}
	log.Info(fmt.Sprintf("cloned %d new image(s)", len(imageOverrides)))

	// Patch resource.
	patch := client.StrategicMergeFrom(object)
	newObject := r.om.ReplaceContainerImages(object, imageOverrides)
	err = r.client.Patch(ctx, newObject, patch)
	if err != nil {
		log.Error(err, "patch failed")
		return reconcile.Result{}, err
	}

	// Wrap up.
	log.Info("reconciliation done")
	return reconcile.Result{}, nil
}
