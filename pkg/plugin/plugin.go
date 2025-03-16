package plugin

import (
	"encoding/json"
	"rolloutplugin-controller/api/v1alpha1"

	"github.com/containerd/containerd/log"
)

type RolloutPlugin interface {
	SetWeight(v1alpha1.RolloutPlugin) (*v1alpha1.RolloutPluginStatus, error)
	SetMirrorRoute(v1alpha1.RolloutPlugin) (*v1alpha1.RolloutPluginStatus, error)
	Rollback(v1alpha1.RolloutPlugin) (*v1alpha1.RolloutPluginStatus, error)
	SetCanaryScale(v1alpha1.RolloutPlugin) (*v1alpha1.RolloutPluginStatus, error)
}

type rolloutPlugin struct {
	rpc    RolloutPlugin
	index  int32
	name   string
	config json.RawMessage
	log    *log.Entry
}

func (r *rolloutPlugin) SetWeight(rollout v1alpha1.RolloutPlugin) (*v1alpha1.RolloutPluginStatus, error) {
	return nil, nil
}

func (r *rolloutPlugin) SetMirrorRoute(rollout v1alpha1.RolloutPlugin) (*v1alpha1.RolloutPluginStatus, error) {
	return nil, nil
}

func (r *rolloutPlugin) Rollback(rollout v1alpha1.RolloutPlugin) (*v1alpha1.RolloutPluginStatus, error) {
	return nil, nil
}

func (r *rolloutPlugin) SetCanaryScale(rollout v1alpha1.RolloutPlugin) (*v1alpha1.RolloutPluginStatus, error) {
	return nil, nil
}
