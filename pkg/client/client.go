package client

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"sync"

	"github.com/aburan28/rolloutplugin-controller/pkg/plugin/rpc"
	types "github.com/aburan28/rolloutplugin-controller/pkg/types"
	"github.com/argoproj/argo-rollouts/utils/config"
	"github.com/argoproj/argo-rollouts/utils/defaults"

	goPlugin "github.com/hashicorp/go-plugin"
)

type RolloutPlugin struct {
	client map[string]*goPlugin.Client
	plugin map[string]rpc.RolloutPlugin
}

var pluginClients *RolloutPlugin
var once sync.Once
var mutex sync.Mutex

var handshakeConfig = goPlugin.HandshakeConfig{
	ProtocolVersion:  1,
	MagicCookieKey:   "ARGO_ROLLOUTS_RPC_PLUGIN",
	MagicCookieValue: "statefulset",
}

// pluginMap is the map of plugins we can dispense.
var pluginMap = map[string]goPlugin.Plugin{
	// "RpcRolloutPlugin": &rpc.RpcRolloutPlugin{},
	"RpcRolloutPlugin": &rpc.RpcRolloutPlugin{},
}

func GetPluginInfo(pluginName string, pluginType types.PluginType) (string, []string, error) {
	configMap, err := config.GetConfig()
	if err != nil {
		return "", nil, fmt.Errorf("failed to get config: %w", err)
	}

	plugin := configMap.GetPlugin(pluginName, pluginType)
	if plugin == nil {
		return "", nil, fmt.Errorf("plugin %s not configured in configmap", pluginName)
	}

	dir, filename, err := config.GetPluginDirectoryAndFilename(plugin.Name)
	if err != nil {
		return "", nil, err
	}
	absFilePath, err := filepath.Abs(filepath.Join(defaults.DefaultRolloutPluginFolder, dir, filename))
	if err != nil {
		return "", nil, fmt.Errorf("failed to get absolute path of plugin folder: %w", err)
	}
	return absFilePath, plugin.Args, nil

}

func (t *RolloutPlugin) startPlugin(pluginName string) (rpc.RolloutPlugin, error) {
	mutex.Lock()
	defer mutex.Unlock()

	if t.client[pluginName] == nil || t.client[pluginName].Exited() {

		// pluginPath, args, err := plugin.GetPluginInfo(pluginName, types.PluginTypeRollout)
		// if err != nil {
		// 	return nil, fmt.Errorf("unable to find plugin (%s): %w", pluginName, err)
		// }
		pluginPath := "rolloutplugin"
		t.client[pluginName] = goPlugin.NewClient(&goPlugin.ClientConfig{
			HandshakeConfig: handshakeConfig,
			Plugins:         pluginMap,
			Cmd:             exec.Command(pluginPath),
			Managed:         true,
		})

		rpcClient, err := t.client[pluginName].Client()
		if err != nil {
			return nil, fmt.Errorf("unable to get plugin client (%s): %w", pluginName, err)
		}

		// Request the plugin
		plugin, err := rpcClient.Dispense("RpcRolloutPlugin")
		if err != nil {
			return nil, fmt.Errorf("unable to dispense plugin (%s): %w", pluginName, err)
		}

		pluginType, ok := plugin.(rpc.RolloutPlugin)
		if !ok {
			return nil, fmt.Errorf("unexpected type from plugin")
		}
		t.plugin[pluginName] = pluginType

		resp := t.plugin[pluginName].InitPlugin()
		if resp.HasError() {
			return nil, fmt.Errorf("unable to initialize plugin via rpc (%s): %w", pluginName, resp)
		}
	}

	client, err := t.client[pluginName].Client()
	if err != nil {
		// If we are not able to create the client, something is utterly wrong
		// we should try to re-download the plugin and restart because the file
		// can be corrupted
		return nil, fmt.Errorf("unable to get plugin client (%s) for ping: %w", pluginName, err)
	}
	if err := client.Ping(); err != nil {
		t.client[pluginName].Kill()
		t.client[pluginName] = nil
		return nil, fmt.Errorf("could not ping plugin will cleanup process so we can restart it next reconcile (%w)", err)
	}

	return t.plugin[pluginName], nil
}
