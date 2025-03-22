package plugin

import (
	"encoding/json"
	"rolloutplugin-controller/api/v1alpha1"
	"rolloutplugin-controller/pkg/types"

	log "github.com/sirupsen/logrus"
)

type RolloutPlugin interface {
	SetWeight(v1alpha1.RolloutPlugin) types.RpcError
	SetMirrorRoute(v1alpha1.RolloutPlugin) types.RpcError
	Rollback(v1alpha1.RolloutPlugin) types.RpcError
	SetCanaryScale(v1alpha1.RolloutPlugin) types.RpcError
}

type rolloutPlugin struct {
	rpc    RolloutPlugin
	index  int32
	name   string
	config json.RawMessage
	log    *log.Entry
}

func (r *rolloutPlugin) SetWeight(rollout v1alpha1.RolloutPlugin) types.RpcError {
	return nil
}

func (r *rolloutPlugin) SetMirrorRoute(rollout v1alpha1.RolloutPlugin) types.RpcError {
	return nil
}

func (r *rolloutPlugin) Rollback(rollout v1alpha1.RolloutPlugin) types.RpcError {
	return nil
}

func (r *rolloutPlugin) SetCanaryScale(rollout v1alpha1.RolloutPlugin) types.RpcError {
	return nil
}
