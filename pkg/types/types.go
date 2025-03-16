package types

import "encoding/gob"

func init() {
	gob.RegisterName("RpcError", new(RpcError))
}

type RpcError struct {
	ErrorString string
}

func (e RpcError) Error() string {
	return e.ErrorString
}

func (e RpcError) HasError() bool {
	return e.ErrorString != ""
}

type PluginType string

const (
	PluginTypeRolloutPlugin PluginType = "RolloutPlugin"
)

type PluginItem struct {
	// Name of the plugin to use in the Rollout custom resources
	Name string `json:"name" yaml:"name"`
	// Location of the plugin. Supports http(s):// urls and file:// prefix
	Location string `json:"location" yaml:"location"`
	// Sha256 is the checksum of the file specified at the provided Location
	Sha256 string `json:"sha256" yaml:"sha256"`
	// Type of the plugin
	Type PluginType
	// Disabled indicates if the plugin should be ignored when referenced in Rollout custom resources. Only valid for a plugin of type Step.
	Disabled bool `json:"disabled" yaml:"disabled"`
}
