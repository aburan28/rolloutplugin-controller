package hash

import (
	"testing"

	"github.com/aburan28/rolloutplugin-controller/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
)

func TestComputeStepHash(t *testing.T) {
	rollout := v1alpha1.RolloutPlugin{
		Spec: v1alpha1.RolloutPluginSpec{
			Strategy: v1alpha1.Strategy{
				Type: "Canary",
				Canary: v1alpha1.Canary{
					Steps: []v1alpha1.CanaryStep{
						{SetWeight: int32Ptr(20)},
						{SetWeight: int32Ptr(50)},
					},
				},
			},
		},
	}

	hash1, err := ComputeStepHash(rollout)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if hash1 == "" {
		t.Fatal("expected non-empty hash")
	}

	// Same input should produce same hash
	hash2, err := ComputeStepHash(rollout)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if hash1 != hash2 {
		t.Fatalf("expected same hash for same input, got %s and %s", hash1, hash2)
	}

	// Different input should produce different hash
	rollout2 := v1alpha1.RolloutPlugin{
		Spec: v1alpha1.RolloutPluginSpec{
			Strategy: v1alpha1.Strategy{
				Type: "Canary",
				Canary: v1alpha1.Canary{
					Steps: []v1alpha1.CanaryStep{
						{SetWeight: int32Ptr(80)},
					},
				},
			},
		},
	}
	hash3, err := ComputeStepHash(rollout2)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if hash1 == hash3 {
		t.Fatalf("expected different hash for different input, got %s", hash1)
	}
}

func TestComputeStepHash_EmptySteps(t *testing.T) {
	rollout := v1alpha1.RolloutPlugin{
		Spec: v1alpha1.RolloutPluginSpec{
			Strategy: v1alpha1.Strategy{
				Type: "Canary",
				Canary: v1alpha1.Canary{
					Steps: []v1alpha1.CanaryStep{},
				},
			},
		},
	}

	hash, err := ComputeStepHash(rollout)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if hash == "" {
		t.Fatal("expected non-empty hash for empty steps")
	}
}

func TestComputePodTemplateHash(t *testing.T) {
	template := &corev1.PodTemplateSpec{
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Name:  "test",
					Image: "test:v1",
				},
			},
		},
	}

	collisionCount := int32(0)
	hash1, err := ComputePodTemplateHash(template, &collisionCount)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if hash1 == "" {
		t.Fatal("expected non-empty hash")
	}

	// Same input should produce same hash
	hash2, err := ComputePodTemplateHash(template, &collisionCount)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if hash1 != hash2 {
		t.Fatalf("expected same hash for same input, got %s and %s", hash1, hash2)
	}

	// Different collision count should produce different hash
	collisionCount2 := int32(1)
	hash3, err := ComputePodTemplateHash(template, &collisionCount2)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if hash1 == hash3 {
		t.Fatalf("expected different hash for different collision count, got %s", hash1)
	}
}

func TestComputePodTemplateHash_NilCollisionCount(t *testing.T) {
	template := &corev1.PodTemplateSpec{
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Name:  "test",
					Image: "test:v1",
				},
			},
		},
	}

	hash, err := ComputePodTemplateHash(template, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if hash == "" {
		t.Fatal("expected non-empty hash")
	}
}

func int32Ptr(i int32) *int32 {
	return &i
}
