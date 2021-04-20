package controller

import (
	corev1 "k8s.io/api/core/v1"
)

func getUniqueImagesFromPodTemplate(s corev1.PodTemplateSpec) []string {
	var images []string
	for _, c := range s.Spec.Containers {
		if inStrings(images, c.Image) {
			continue
		}
		images = append(images, c.Image)
	}
	return images
}

func inStrings(haystack []string, needle string) bool {
	for _, v := range haystack {
		if v == needle {
			return true
		}
	}
	return false
}

func replaceImagesInPodTemplate(template corev1.PodTemplateSpec, images map[string]string) corev1.PodTemplateSpec {
	// Create a "deep enough" copy of the pod template spec to avoid overwriting the original slice of Containers.
	newTemplate := template
	newTemplate.Spec.Containers = make([]corev1.Container, len(template.Spec.Containers))
	copy(newTemplate.Spec.Containers, template.Spec.Containers)

	for i, c := range newTemplate.Spec.Containers {
		if newImage, ok := images[c.Image]; ok {
			newTemplate.Spec.Containers[i].Image = newImage
		}
	}
	return newTemplate
}
