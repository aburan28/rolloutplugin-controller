package rpc

import (
	"encoding/gob"
	"fmt"
	"net/rpc"

	"github.com/aburan28/rolloutplugin-controller/api/v1alpha1"
	types "github.com/aburan28/rolloutplugin-controller/pkg/types"

	"github.com/hashicorp/go-plugin"
)

type RunArgs struct {
	RolloutPlugin *v1alpha1.RolloutPlugin
	Context       types.RpcRolloutContext
}

type TerminateArgs struct {
	RolloutPlugin *v1alpha1.RolloutPlugin
	Context       types.RpcRolloutContext
}

type AbortArgs struct {
	RolloutPlugin *v1alpha1.RolloutPlugin
	Context       types.RpcRolloutContext
}

type Response struct {
	Result types.RpcRolloutResult
	Error  types.RpcError
}

func init() {
	gob.RegisterName("rollout.RunArgs", new(RunArgs))
	gob.RegisterName("rollout.TerminateArgs", new(TerminateArgs))
	gob.RegisterName("rollout.AbortArgs", new(AbortArgs))
	gob.RegisterName("rollout.Response", new(Response))

}

type RolloutPlugin interface {
	InitPlugin() types.RpcError
	SetWeight(*v1alpha1.RolloutPlugin) types.RpcError
	SetMirrorRoute(*v1alpha1.RolloutPlugin) types.RpcError
	Rollback(*v1alpha1.RolloutPlugin) types.RpcError
	SetCanaryScale(*v1alpha1.RolloutPlugin) types.RpcError
	Run(*v1alpha1.RolloutPlugin, types.RpcRolloutContext) (types.RpcRolloutResult, types.RpcError)
	Terminate(*v1alpha1.RolloutPlugin, types.RpcRolloutContext) (types.RpcRolloutResult, types.RpcError)
	Abort(*v1alpha1.RolloutPlugin, types.RpcRolloutContext) (types.RpcRolloutResult, types.RpcError)
	Type() string
	Sync(rollout *v1alpha1.RolloutPlugin) types.RpcError
	SetHeaderRoute(*v1alpha1.RolloutPlugin) types.RpcError
}

// RolloutPluginRPC Here is an implementation that talks over RPC
type RolloutPluginRPC struct{ client *rpc.Client }

// InitPlugin is the client aka the controller side function that calls the server side rpc (plugin)
// this gets called once during startup of the plugin and can be used to set up informers, k8s clients, etc.
func (g *RolloutPluginRPC) InitPlugin() types.RpcError {
	var resp types.RpcError
	err := g.client.Call("Plugin.InitPlugin", new(any), &resp)
	if err != nil {
		return types.RpcError{ErrorString: fmt.Sprintf("InitPlugin rpc call error: %s", err)}
	}
	return resp
}

// Run executes the rollout
func (g *RolloutPluginRPC) Run(rollout *v1alpha1.RolloutPlugin, context types.RpcRolloutContext) (types.RpcRolloutResult, types.RpcError) {
	var resp Response
	var args any = RunArgs{
		RolloutPlugin: rollout,
		Context:       context,
	}
	err := g.client.Call("Plugin.Run", &args, &resp)
	if err != nil {
		return types.RpcRolloutResult{}, types.RpcError{ErrorString: fmt.Sprintf("Run rpc call error: %s", err)}
	}
	return resp.Result, resp.Error
}

// Terminate stops the execution of a running rollout and exits early
func (g *RolloutPluginRPC) Terminate(rollout *v1alpha1.RolloutPlugin, context types.RpcRolloutContext) (types.RpcRolloutResult, types.RpcError) {
	var resp Response
	var args any = TerminateArgs{
		RolloutPlugin: rollout,
		Context:       context,
	}
	err := g.client.Call("Plugin.Terminate", &args, &resp)
	if err != nil {
		return types.RpcRolloutResult{}, types.RpcError{ErrorString: fmt.Sprintf("Terminate rpc call error: %s", err)}
	}
	return resp.Result, resp.Error
}

// Abort reverts previous operation executed by the rollout if necessary
func (g *RolloutPluginRPC) Abort(rollout *v1alpha1.RolloutPlugin, context types.RpcRolloutContext) (types.RpcRolloutResult, types.RpcError) {
	var resp Response
	var args any = AbortArgs{
		RolloutPlugin: rollout,
		Context:       context,
	}
	err := g.client.Call("Plugin.Abort", &args, &resp)
	if err != nil {
		return types.RpcRolloutResult{}, types.RpcError{ErrorString: fmt.Sprintf("Abort rpc call error: %s", err)}
	}
	return resp.Result, resp.Error
}

// Type returns the type of the traffic routing reconciler
func (g *RolloutPluginRPC) Type() string {
	var resp string
	err := g.client.Call("Plugin.Type", new(any), &resp)
	if err != nil {
		return fmt.Sprintf("Type rpc call error: %s", err)
	}

	return resp
}

func (g *RolloutPluginRPC) Sync(rollout *v1alpha1.RolloutPlugin) types.RpcError {
	var resp string
	err := g.client.Call("Plugin.Sync", new(any), &resp)
	if err != nil {
		return types.RpcError{ErrorString: fmt.Sprintf("Sync rpc call error: %s", err)}
	}
	return types.RpcError{ErrorString: resp}
}

func (g *RolloutPluginRPC) SetWeight(rollout *v1alpha1.RolloutPlugin) types.RpcError {
	var resp types.RpcError
	err := g.client.Call("Plugin.SetWeight", rollout, &resp)
	if err != nil {
		return types.RpcError{ErrorString: fmt.Sprintf("SetWeight rpc call error: %s", err)}
	}
	return resp
}

func (g *RolloutPluginRPC) SetMirrorRoute(rollout *v1alpha1.RolloutPlugin) types.RpcError {
	var resp types.RpcError
	err := g.client.Call("Plugin.SetMirrorRoute", rollout, &resp)
	if err != nil {
		return types.RpcError{ErrorString: fmt.Sprintf("SetMirrorRoute rpc call error: %s", err)}
	}
	return resp
}
func (g *RolloutPluginRPC) Rollback(rollout *v1alpha1.RolloutPlugin) types.RpcError {
	var resp types.RpcError
	err := g.client.Call("Plugin.Rollback", rollout, &resp)
	if err != nil {
		return types.RpcError{ErrorString: fmt.Sprintf("Rollback rpc call error: %s", err)}
	}
	return resp
}

func (g *RolloutPluginRPC) SetCanaryScale(rollout *v1alpha1.RolloutPlugin) types.RpcError {
	var resp types.RpcError
	err := g.client.Call("Plugin.SetCanaryScale", rollout, &resp)
	if err != nil {
		return types.RpcError{ErrorString: fmt.Sprintf("SetCanaryScale rpc call error: %s", err)}
	}
	return resp
}

func (g *RolloutPluginRPC) SetHeaderRoute(rollout *v1alpha1.RolloutPlugin) types.RpcError {
	var resp types.RpcError
	err := g.client.Call("Plugin.SetHeaderRoute", rollout, &resp)
	if err != nil {
		return types.RpcError{ErrorString: fmt.Sprintf("SetHeaderRoute rpc call error: %s", err)}
	}
	return resp
}

// RolloutPluginServerRPC Here is the RPC server that MetricsPluginRPC talks to, conforming to
// the requirements of net/rpc
type RolloutPluginServerRPC struct {
	// This is the real implementation
	Impl RolloutPlugin
}

// InitPlugin this is the server aka the controller side function that receives calls from the client side rpc (controller)
// this gets called once during startup of the plugin and can be used to set up informers or k8s clients etc.
func (s *RolloutPluginServerRPC) InitPlugin(args any, resp *types.RpcError) error {
	*resp = s.Impl.InitPlugin()
	return nil
}

// Run executes the rollout
func (s *RolloutPluginServerRPC) Run(args any, resp *Response) error {
	runArgs, ok := args.(*RunArgs)
	if !ok {
		return fmt.Errorf("invalid args %s", args)
	}
	result, err := s.Impl.Run(runArgs.RolloutPlugin, runArgs.Context)
	*resp = Response{
		Result: result,
		Error:  err,
	}
	return nil
}

// Terminate stops the execution of a running rollout and exits early
func (s *RolloutPluginServerRPC) Terminate(args any, resp *Response) error {
	runArgs, ok := args.(*TerminateArgs)
	if !ok {
		return fmt.Errorf("invalid args %s", args)
	}
	result, err := s.Impl.Terminate(runArgs.RolloutPlugin, runArgs.Context)
	*resp = Response{
		Result: result,
		Error:  err,
	}
	return nil
}

// Abort reverts previous operation executed by the rollout if necessary
func (s *RolloutPluginServerRPC) Abort(args any, resp *Response) error {
	runArgs, ok := args.(*AbortArgs)
	if !ok {
		return fmt.Errorf("invalid args %s", args)
	}
	result, err := s.Impl.Abort(runArgs.RolloutPlugin, runArgs.Context)
	*resp = Response{
		Result: result,
		Error:  err,
	}
	return nil
}

func (s *RolloutPluginServerRPC) SetWeight(args *v1alpha1.RolloutPlugin, resp *types.RpcError) error {
	*resp = s.Impl.SetWeight(args)
	return nil
}

func (s *RolloutPluginServerRPC) SetMirrorRoute(args *v1alpha1.RolloutPlugin, resp *types.RpcError) error {
	*resp = s.Impl.SetMirrorRoute(args)
	return nil
}

func (s *RolloutPluginServerRPC) Rollback(args *v1alpha1.RolloutPlugin, resp *types.RpcError) error {
	*resp = s.Impl.Rollback(args)
	return nil
}

func (s *RolloutPluginServerRPC) SetCanaryScale(args *v1alpha1.RolloutPlugin, resp *types.RpcError) error {
	*resp = s.Impl.SetCanaryScale(args)
	return nil
}

// Type returns the type of the traffic routing reconciler
func (s *RolloutPluginServerRPC) Type(args any, resp *string) error {
	*resp = s.Impl.Type()
	return nil
}

func (s *RolloutPluginServerRPC) Sync(args any, resp *string) error {
	syncErr := s.Impl.Sync()
	*resp = syncErr.ErrorString
	return nil
}

func (s *RolloutPluginServerRPC) SetHeaderRoute(args *v1alpha1.RolloutPlugin, resp *types.RpcError) error {
	*resp = s.Impl.SetHeaderRoute(args)
	return nil
}

// RpcrolloutPlugin This is the implementation of plugin.Plugin so we can serve/consume
//
// This has two methods: Server must return an RPC server for this plugin
// type. We construct a RolloutPluginServerRPC for this.
//
// Client must return an implementation of our interface that communicates
// over an RPC client. We return RolloutPluginRPC for this.
//
// Ignore MuxBroker. That is used to create more multiplexed streams on our
// plugin connection and is a more advanced use case.
type RpcRolloutPlugin struct {
	// Impl Injection
	Impl RolloutPlugin
}

func (p *RpcRolloutPlugin) Server(*plugin.MuxBroker) (any, error) {
	return &RolloutPluginServerRPC{Impl: p.Impl}, nil
}

func (RpcRolloutPlugin) Client(b *plugin.MuxBroker, c *rpc.Client) (any, error) {
	return &RolloutPluginRPC{client: c}, nil
}
