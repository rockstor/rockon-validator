package main

import (
	"encoding/json"
	"strings"
)

// A map with a single entry, the Rock-on name. eg: LSIO-Plex
type RockOn map[string]RockonDetails

func (r RockOn) ToJSON() (string, error) {
	var tmp strings.Builder
	enc := json.NewEncoder(&tmp)
	enc.SetEscapeHTML(false)
	enc.SetIndent("", "  ")

	err := enc.Encode(r)
	if err != nil {
		return "", err
	}

	out := strings.ReplaceAll(tmp.String(), "\\u0026", "&")
	out = strings.ReplaceAll(out, "\\u003c", "<")
	out = strings.ReplaceAll(out, "\\u003e", ">")
	return out, nil
}

type RockonDetails struct {
	Description      string                   `json:"description"`         // description of the Rock-on. Eg: Plex brought to you by Linuxserver.io
	Version          string                   `json:"version"`             // arbitrary version string
	Website          string                   `json:"website"`             // Underlying app website
	Icon             string                   `json:"icon,omitempty"`      // link to icon, if any
	MoreInfo         string                   `json:"more_info,omitempty"` // string or html with more information to display to the user in the Rockstor UI
	UI               *UISlug                  `json:"ui,omitempty"` // contains the slug, if applicable, that the main web ui will be accessible from
	VolumeAddSupport bool                     `json:"volume_add_support,omitempty"` // If the app allows arbitrary Shares to be mapped to the main container>,
	Containers       map[string]Container     `json:"containers"`                   // map of container names to Container objects
	CustomConfig     *map[string]CustomConfig `json:"custom_config,omitempty"`      // custom configuration object that a special install handler of this Rock-on expects
}

type UISlug struct {
	Slug string `json:"slug,omitempty"` // link to webui becomes ROCKSTOR_IP:PORT/gui with slug value gui
}

func (r RockonDetails) MarshalJSON() ([]byte, error) {
	type ro RockonDetails
	if r.UI != nil && r.UI.Slug == "" {
		r.UI = nil
	}
	return json.Marshal(ro(r))
}

type Container struct {
	Image        string                    `json:"image"`                   // docker image. eg: linuxserver/plex
	Tag          string                    `json:"tag,omitempty"`           // tag of the docker image, if any. latest is used by default.
	LaunchOrder  int                       `json:"launch_order"`            // 1 or above. If there are multiple containers and they must be started in order, specify here.
	Ports        map[string]Port           `json:"ports"`                   // Map of (container) port numbers to Port objects, mapping the container port to the host
	Volumes      map[string]Volume         `json:"volumes"`                 // Map of container mount points to Volume objects, representing Shares to be mounted in the container
	Opts         []Option                  `json:"opts,omitempty"`          // Array of Option objects that represent container options such as --net=host etc.
	CmdArguments []CmdArgument             `json:"cmd_arguments,omitempty"` // Array of CmdArgument objects that represent arguments to pass to the 'docker run' command.
	Environment  map[string]EnvironmentVar `json:"environment,omitempty"`   // Map of environment variable names to EnvironmentVar objects, representing the value
	Devices      map[string]Device         `json:"devices,omitempty"`       // Map of device paths to Device objects, to be passed through to the container
}

type Port struct {
	Description string   `json:"description"`        // A detailed description of this port mapping, why it's for etc..
	Label       string   `json:"label"`              // A short label for this mapping. eg: Web-UI port
	HostDefault int      `json:"host_default"`       // suggested port number on the host. eg: 8080
	Protocol    Protocol `json:"protocol,omitempty"` // tcp or udp, default is to map both tcp and udp simultaneously
	UI          bool     `json:"ui,omitempty"`       // Is port used for Web UI. Not needed if false
}

type Protocol string

const (
	TCP Protocol = "tcp"
	UDP Protocol = "udp"
)

type Volume struct {
	Description string `json:"description"`        // A detailed description. Eg: This is where all incoming syncthing data will be stored
	Label       string `json:"label"`              // A short label. eg: Data Storage
	MinSize     int    `json:"min_size,omitempty"` // suggested minimum size of the Share, in KB
}

// An options object is a list of exactly two elements.
//
// `--net=host` would be represented as: ["--net", "host"]
type Option [2]string

// A command arguments object is a list of exactly two elements detailing specific arguments to be passed onto the docker run command. As these arguments will simply be appended to the docker run command, they need to follow the same syntax and order. For instance,
//
// `docker run <...> image/name argument1 argument2="text2"` would be represented as:
//
// ["argument1", "argument2="text2"]
type CmdArgument [2]string

type EnvironmentVar struct {
	Description string `json:"description"`     // Detailed description. Eg: Login username for Syncthing UI
	Label       string `json:"label"`           // A short label. eg: Web-UI username
	Index       int    `json:"index,omitempty"` // 1 or above. Order of this environment variable, if relevant
}

type Device struct {
	Description string `json:"description"`     // Detailed description of the device and its intent or specificities. Eg: path to device (/dev/xxx)
	Label       string `json:"label"`           // A short label. eg: Hardware encoding device
	Index       int    `json:"index,omitempty"` // 1 or above. Order of this device, if relevant
}

type CustomConfig struct {
	Description string `json:"description"`
	Label       string `json:"label"`
}
