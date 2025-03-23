package pluginclient

import (
	"context"
	"crypto/sha256"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"reflect"
	"sync"
	"time"

	"github.com/aburan28/rolloutplugin-controller/pkg/plugin/rpc"

	goPlugin "github.com/hashicorp/go-plugin"
)

type rolloutPlugin struct {
	Client map[string]*goPlugin.Client
	Plugin map[string]rpc.RolloutPlugin
}

var pluginClients *rolloutPlugin
var once sync.Once
var mutex sync.Mutex

var handshakeConfig = goPlugin.HandshakeConfig{
	ProtocolVersion:  1,
	MagicCookieKey:   "ARGO_ROLLOUTS_RPC_PLUGIN",
	MagicCookieValue: "rolloutplugin",
}

// pluginMap is the map of plugins we can dispense.
var pluginMap = map[string]goPlugin.Plugin{
	// "RpcRolloutPlugin": &rpc.RpcRolloutPlugin{},
	"RpcRolloutPlugin": &rpc.RpcRolloutPlugin{},
}

type rolloutPluginMap struct {
	pluginClient map[string]*goPlugin.Client
	plugin       map[string]rpc.RolloutPlugin
}

func NewRolloutPlugin() *rolloutPlugin {
	return &rolloutPlugin{
		Client: make(map[string]*goPlugin.Client),
		Plugin: make(map[string]rpc.RolloutPlugin),
	}
}

func GetRolloutPlugin(pluginName string) (rpc.RolloutPlugin, error) {
	once.Do(func() {
		pluginClients = NewRolloutPlugin()
	})

	return pluginClients.StartPlugin(pluginName)
}

func (t *rolloutPlugin) StartPlugin(pluginName string) (rpc.RolloutPlugin, error) {
	mutex.Lock()
	defer mutex.Unlock()
	if t.Client[pluginName] == nil || t.Client[pluginName].Exited() {

		fmt.Println("pluginPath: ", pluginName)

		t.Client[pluginName] = goPlugin.NewClient(&goPlugin.ClientConfig{
			HandshakeConfig: handshakeConfig,
			Plugins:         pluginMap,
			// AllowedProtocols: []goPlugin.Protocol{
			// 	goPlugin.ProtocolNetRPC,
			// 	goPlugin.ProtocolGRPC,
			// },
			StartTimeout: 90 * time.Second,
			Cmd:          exec.Command("./" + pluginName),
			Managed:      true,
		})

		rpcClient, err := t.Client[pluginName].Client()
		if err != nil {
			return nil, fmt.Errorf("unable to get plugin client (%s): %w", pluginName, err)
		}

		// Request the plugin
		plugin, err := rpcClient.Dispense("RpcRolloutPlugin")
		if err != nil {
			return nil, fmt.Errorf("unable to dispense plugin (%s): %w", pluginName, err)
		}
		fmt.Println(reflect.TypeOf(plugin))
		pluginType, ok := plugin.(rpc.RolloutPlugin)
		if !ok {
			return nil, fmt.Errorf("unexpected type from plugin")
		}
		t.Plugin[pluginName] = pluginType

		resp := t.Plugin[pluginName].InitPlugin()
		if resp.HasError() {
			return nil, fmt.Errorf("unable to initialize plugin via rpc (%s): %w", pluginName, resp)
		}
	}

	client, err := t.Client[pluginName].Client()
	if err != nil {
		// If we are not able to create the client, something is utterly wrong
		// we should try to re-download the plugin and restart because the file
		// can be corrupted
		return nil, fmt.Errorf("unable to get plugin client (%s) for ping: %w", pluginName, err)
	}
	if err := client.Ping(); err != nil {
		t.Client[pluginName].Kill()
		t.Client[pluginName] = nil
		return nil, fmt.Errorf("could not ping plugin will cleanup process so we can restart it next reconcile (%w)", err)
	}

	return t.Plugin[pluginName], nil
}

func DownloadPlugin(ctx context.Context, url string, pluginName string) error {

	pluginUrl := url + pluginName
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, pluginUrl, nil)
	if err != nil {
		return fmt.Errorf("unable to create plugin download request: %w", err)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("unable to download plugin: %w", err)
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("unable to read plugin response body: %w", err)
	}

	err = os.WriteFile(pluginName, body, 0777)
	if err != nil {
		return fmt.Errorf("unable to write plugin to file: %w", err)
	}
	return nil
}

func VerifyPlugin(pluginName string, checksum string) error {
	file, err := os.ReadFile(pluginName)
	if err != nil {
		return fmt.Errorf("unable to read plugin file: %w", err)
	}
	h := sha256.New()
	h.Write(file)
	bs := h.Sum(nil)
	hash := fmt.Sprintf("%x", bs)
	if fmt.Sprintf("%x", hash) != checksum {
		return fmt.Errorf("sha256 hash does not match")
	}
	return nil
}
