package controller

import (
	appsv1 "k8s.io/api/apps/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// ObjectManager abstracts away the details of dealing with Kubernetes objects for Reconciler.
type ObjectManager interface {
	// Create returns a new concrete instance of the object.
	Create() client.Object

	// GetContainerImages returns the list of container images referenced by the object.
	GetContainerImages(client.Object) []string

	// ReplaceContainerImages returns a copy of the object with its images replaced according to the provided map.
	ReplaceContainerImages(client.Object, map[string]string) client.Object
}

// DaemonSetManager is an ObjectManager for v1.DaemonSet resources.
type DaemonSetManager struct {
}

// Create implements ObjectManager.
func (d DaemonSetManager) Create() client.Object {
	return &appsv1.DaemonSet{}
}

// GetContainerImages implements ObjectManager.
func (d DaemonSetManager) GetContainerImages(object client.Object) []string {
	o, ok := object.(*appsv1.DaemonSet)
	if !ok {
		panic("unexpected resource type")
	}
	return getUniqueImagesFromPodTemplate(o.Spec.Template)
}

// ReplaceContainerImages implements ObjectManager.
func (d DaemonSetManager) ReplaceContainerImages(object client.Object, images map[string]string) client.Object {
	o, ok := object.(*appsv1.DaemonSet)
	if !ok {
		panic("unexpected resource type")
	}

	o2 := *o
	o2.Spec.Template = replaceImagesInPodTemplate(o.Spec.Template, images)
	return &o2
}

// DeploymentManager is an ObjectManager for v1.Deployment resources.
type DeploymentManager struct {
}

// Create implements ObjectManager.
func (d DeploymentManager) Create() client.Object {
	return &appsv1.Deployment{}
}

// GetContainerImages implements ObjectManager.
func (d DeploymentManager) GetContainerImages(object client.Object) []string {
	o, ok := object.(*appsv1.Deployment)
	if !ok {
		panic("unexpected resource type")
	}
	return getUniqueImagesFromPodTemplate(o.Spec.Template)
}

// ReplaceContainerImages implements ObjectManager.
func (d DeploymentManager) ReplaceContainerImages(object client.Object, images map[string]string) client.Object {
	o, ok := object.(*appsv1.Deployment)
	if !ok {
		panic("unexpected resource type")
	}

	o2 := *o
	o2.Spec.Template = replaceImagesInPodTemplate(o.Spec.Template, images)
	return &o2
}
