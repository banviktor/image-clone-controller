package controller

import (
	"context"
	"github.com/banviktor/image-clone-controller/pkg/imagecloner"
	"github.com/stretchr/testify/assert"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	fakeclient "sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"testing"
)

const targetRepositoryPrefix = "localhost:5000/icc"

func TestReconcileDaemonSets(t *testing.T) {
	cloner, err := imagecloner.NewFlatCloner(targetRepositoryPrefix)
	if err != nil {
		t.Fatalf("error while creating cloner: %v", err)
	}
	testClient := newTestClient()

	r := Reconciler{
		client: testClient,
		om:     &DaemonSetManager{},
		cloner: cloner,
	}

	tests := []struct {
		namespacedName types.NamespacedName
		expectError    bool
		expectedImages []string
	}{
		{
			namespacedName: types.NamespacedName{Name: "nginx", Namespace: "default"},
			expectError:    true,
		},
		{
			namespacedName: types.NamespacedName{Name: "ingress-nginx-controller", Namespace: "ingress-nginx"},
			expectedImages: []string{
				targetRepositoryPrefix + "/k8s.gcr.io_ingress-nginx_controller@sha256:c4390c53f348c3bd4e60a5dd6a11c35799ae78c49388090140b9d72ccede1755",
			},
		},
	}

	for _, test := range tests {
		_, err = r.Reconcile(context.Background(), reconcile.Request{
			NamespacedName: test.namespacedName,
		})
		if test.expectError {
			assert.Error(t, err, "reconcile should error")
			continue
		}
		assert.NoError(t, err, "reconcile should not error")

		o := &appsv1.DaemonSet{}
		err = testClient.Get(context.Background(), test.namespacedName, o)
		assert.NoError(t, err, "resource should still exist")
		for i, expectedImage := range test.expectedImages {
			assert.Equal(t, expectedImage, o.Spec.Template.Spec.Containers[i].Image)
		}
	}
}

func TestReconcileDeployments(t *testing.T) {
	cloner, err := imagecloner.NewFlatCloner(targetRepositoryPrefix)
	if err != nil {
		t.Fatalf("error while creating cloner: %v", err)
	}
	testClient := newTestClient()

	r := Reconciler{
		client: testClient,
		om:     &DeploymentManager{},
		cloner: cloner,
	}

	tests := []struct {
		namespacedName types.NamespacedName
		expectError    bool
		expectedImages []string
	}{
		{
			namespacedName: types.NamespacedName{Name: "typo", Namespace: "default"},
			expectError:    true,
		},
		{
			namespacedName: types.NamespacedName{Name: "nginx", Namespace: "default"},
			expectedImages: []string{
				targetRepositoryPrefix + "/index.docker.io_library_nginx:latest",
			},
		},
		{
			namespacedName: types.NamespacedName{Name: "multi", Namespace: "test"},
			expectedImages: []string{
				targetRepositoryPrefix + "/index.docker.io_library_alpine:3.13",
				targetRepositoryPrefix + "/quay.io_prometheus_node-exporter:v1.1.2",
			},
		},
		{
			namespacedName: types.NamespacedName{Name: "nginx", Namespace: "default"},
			expectedImages: []string{
				targetRepositoryPrefix + "/index.docker.io_library_nginx:latest",
			},
		},
	}

	for _, test := range tests {
		_, err = r.Reconcile(context.Background(), reconcile.Request{
			NamespacedName: test.namespacedName,
		})
		if test.expectError {
			assert.Error(t, err, "reconcile should error")
			continue
		}
		assert.NoError(t, err, "reconcile should not error")

		o := &appsv1.Deployment{}
		err = testClient.Get(context.Background(), test.namespacedName, o)
		assert.NoError(t, err, "resource should still exist")
		for i, expectedImage := range test.expectedImages {
			assert.Equal(t, expectedImage, o.Spec.Template.Spec.Containers[i].Image)
		}
	}
}

func newTestClient() client.Client {
	return fakeclient.NewClientBuilder().WithObjects(
		&appsv1.DaemonSet{
			ObjectMeta: metav1.ObjectMeta{Name: "ingress-nginx-controller", Namespace: "ingress-nginx"},
			Spec: appsv1.DaemonSetSpec{
				Template: corev1.PodTemplateSpec{
					Spec: corev1.PodSpec{
						Containers: []corev1.Container{
							{Name: "controller", Image: "k8s.gcr.io/ingress-nginx/controller:v0.45.0@sha256:c4390c53f348c3bd4e60a5dd6a11c35799ae78c49388090140b9d72ccede1755"},
						},
					},
				},
			},
		},
		&appsv1.Deployment{
			ObjectMeta: metav1.ObjectMeta{Name: "typo", Namespace: "default"},
			Spec: appsv1.DeploymentSpec{
				Template: corev1.PodTemplateSpec{
					Spec: corev1.PodSpec{
						Containers: []corev1.Container{
							{Name: "nginx", Image: "nnnnginx"},
						},
					},
				},
			},
		},
		&appsv1.Deployment{
			ObjectMeta: metav1.ObjectMeta{Name: "nginx", Namespace: "default"},
			Spec: appsv1.DeploymentSpec{
				Template: corev1.PodTemplateSpec{
					Spec: corev1.PodSpec{
						Containers: []corev1.Container{
							{Name: "nginx", Image: "nginx"},
						},
					},
				},
			},
		},
		&appsv1.Deployment{
			ObjectMeta: metav1.ObjectMeta{Name: "multi", Namespace: "test"},
			Spec: appsv1.DeploymentSpec{
				Template: corev1.PodTemplateSpec{
					Spec: corev1.PodSpec{
						Containers: []corev1.Container{
							{Name: "alpine", Image: "alpine:3.13"},
							{Name: "mysql", Image: "quay.io/prometheus/node-exporter:v1.1.2"},
						},
					},
				},
			},
		},
	).Build()
}
