package hash

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"hash/fnv"

	"github.com/aburan28/rolloutplugin-controller/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/rand"
)

func ComputeStepHash(rollout v1alpha1.RolloutPlugin) (string, error) {
	rolloutStepHasher := fnv.New32a()
	stepsBytes, err := json.Marshal(rollout.Spec.Strategy.Canary.Steps)
	if err != nil {
		return "", fmt.Errorf("failed to marshal canary steps: %w", err)
	}
	_, err = rolloutStepHasher.Write(stepsBytes)
	if err != nil {
		return "", fmt.Errorf("failed to write step hash: %w", err)
	}
	return rand.SafeEncodeString(fmt.Sprint(rolloutStepHasher.Sum32())), nil
}

// ComputePodTemplateHash returns a hash value calculated from pod template.
// The hash will be safe encoded to avoid bad words.
func ComputePodTemplateHash(template *corev1.PodTemplateSpec, collisionCount *int32) (string, error) {
	podTemplateSpecHasher := fnv.New32a()
	stepsBytes, err := json.Marshal(template)
	if err != nil {
		return "", fmt.Errorf("failed to marshal pod template: %w", err)
	}
	_, err = podTemplateSpecHasher.Write(stepsBytes)
	if err != nil {
		return "", fmt.Errorf("failed to write pod template hash: %w", err)
	}
	if collisionCount != nil {
		collisionCountBytes := make([]byte, 8)
		binary.LittleEndian.PutUint32(collisionCountBytes, uint32(*collisionCount))
		_, err = podTemplateSpecHasher.Write(collisionCountBytes)
		if err != nil {
			return "", fmt.Errorf("failed to write collision count hash: %w", err)
		}
	}
	return rand.SafeEncodeString(fmt.Sprint(podTemplateSpecHasher.Sum32())), nil
}
