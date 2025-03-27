package controller

import (
	"encoding/json"
	"fmt"
	"hash/fnv"

	"k8s.io/apimachinery/pkg/util/rand"

	"github.com/aburan28/rolloutplugin-controller/api/v1alpha1"
)

const ()

func ComputeStepHash(rollout v1alpha1.RolloutPlugin) string {
	if rollout.Spec.Strategy.BlueGreen != nil || rollout.Spec.Strategy.Canary == nil {
		return ""
	}
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
