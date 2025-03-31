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

func ComputeStepHash(rollout v1alpha1.RolloutPlugin) string {
	// if rollout.Spec.Strategy.Canary == nil {
	// 	return ""
	// }
	rolloutStepHasher := fnv.New32a()
	stepsBytes, err := json.Marshal(rollout.Spec.Strategy.Canary.Steps)
	if err != nil {
		panic(err)
	}
	_, err = rolloutStepHasher.Write(stepsBytes)
	if err != nil {
		panic(err)
	}
	return rand.SafeEncodeString(fmt.Sprint(rolloutStepHasher.Sum32()))
}

// ComputePodTemplateHash returns a hash value calculated from pod template.
// The hash will be safe encoded to avoid bad words.
func ComputePodTemplateHash(template *corev1.PodTemplateSpec, collisionCount *int32) string {
	podTemplateSpecHasher := fnv.New32a()
	stepsBytes, err := json.Marshal(template)
	if err != nil {
		panic(err)
	}
	_, err = podTemplateSpecHasher.Write(stepsBytes)
	if err != nil {

		panic(err)
	}
	if collisionCount != nil {
		collisionCountBytes := make([]byte, 8)
		binary.LittleEndian.PutUint32(collisionCountBytes, uint32(*collisionCount))
		_, err = podTemplateSpecHasher.Write(collisionCountBytes)
		if err != nil {
			panic(err)
		}
	}
	return rand.SafeEncodeString(fmt.Sprint(podTemplateSpecHasher.Sum32()))
}
