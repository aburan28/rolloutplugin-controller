package client

import (
	"fmt"
	"os/exec"
	"rolloutplugin-controller/pkg/plugin"
	"sync"

	goPlugin "github.com/hashicorp/go-plugin"
)

type RolloutPlugin struct {
	client map[string]*goPlugin.Client
	plugin map[string]plugin.Plugin
}

var pluginClients *stepPlugin
var once sync.Once
var mutex sync.Mutex

var handshakeConfig = goPlugin.HandshakeConfig{
	ProtocolVersion:  1,
	MagicCookieKey:   "ARGO_ROLLOUTS_RPC_PLUGIN",
	MagicCookieValue: "step",
}

// pluginMap is the map of plugins we can dispense.
var pluginMap = map[string]goPlugin.Plugin{
	"RpcStepPlugin": &rpc.RpcStepPlugin{},
}

func (t *stepPlugin) startPlugin(pluginName string) (rpc.StepPlugin, error) {
	mutex.Lock()
	defer mutex.Unlock()

	if t.client[pluginName] == nil || t.client[pluginName].Exited() {

		pluginPath, args, err := plugin.GetPluginInfo(pluginName, types.PluginTypeStep)
		if err != nil {
			return nil, fmt.Errorf("unable to find plugin (%s): %w", pluginName, err)
		}

		t.client[pluginName] = goPlugin.NewClient(&goPlugin.ClientConfig{
			HandshakeConfig: handshakeConfig,
			Plugins:         pluginMap,
			Cmd:             exec.Command(pluginPath, args...),
			Managed:         true,
		})

		rpcClient, err := t.client[pluginName].Client()
		if err != nil {
			return nil, fmt.Errorf("unable to get plugin client (%s): %w", pluginName, err)
		}

		// Request the plugin
		plugin, err := rpcClient.Dispense("RpcStepPlugin")
		if err != nil {
			return nil, fmt.Errorf("unable to dispense plugin (%s): %w", pluginName, err)
		}

		pluginType, ok := plugin.(rpc.StepPlugin)
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
