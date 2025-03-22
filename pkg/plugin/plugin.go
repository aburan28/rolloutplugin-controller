package plugin

import (
	"encoding/json"

	"github.com/aburan28/rolloutplugin-controller/api/v1alpha1"
	"github.com/aburan28/rolloutplugin-controller/pkg/types"

	log "github.com/sirupsen/logrus"
)

type RolloutPlugin interface {
	SetWeight(v1alpha1.RolloutPlugin) types.RpcError
	SetMirrorRoute(v1alpha1.RolloutPlugin) types.RpcError
	Rollback(v1alpha1.RolloutPlugin) types.RpcError
	SetCanaryScale(v1alpha1.RolloutPlugin) types.RpcError
	Run(v1alpha1.RolloutPlugin, types.RpcRolloutContext) (types.RpcRolloutResult, types.RpcError)
}

type rolloutPlugin struct {
	rpc    RolloutPlugin
	index  int32
	name   string
	config json.RawMessage
	log    *log.Entry
}

func (r *rolloutPlugin) SetWeight(rollout v1alpha1.RolloutPlugin) types.RpcError {
	return types.RpcError{}
}

func (r *rolloutPlugin) SetMirrorRoute(rollout v1alpha1.RolloutPlugin) types.RpcError {
	return types.RpcError{}
}

func (r *rolloutPlugin) Rollback(rollout v1alpha1.RolloutPlugin) types.RpcError {
	return types.RpcError{}
}

func (r *rolloutPlugin) SetCanaryScale(rollout v1alpha1.RolloutPlugin) types.RpcError {
	return types.RpcError{}
}
