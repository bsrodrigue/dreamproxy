package main

import (
	"dreamproxy/config"
	"dreamproxy/dream"
	"strconv"
)

const LOG_FORMAT string = "text"
const CONFIG_FILE string = "./Dreamfile"

var dreamconfig config.Config

func main() {
	ctxts := []dream.DreamContext{}
	dreamconfig = config.LoadDreamFile(CONFIG_FILE)

	config_map := map[string][]config.Server{}

	// Map each server configuration to a unique port
	for _, server_config := range dreamconfig.Servers {
		port_str := strconv.Itoa(server_config.Listen.Port)
		if config_map[port_str] == nil {
			config_map[port_str] = make([]config.Server, 0, 10)
		}

		config_map[port_str] = append(config_map[port_str], server_config)
	}

	for port_str, server_configs := range config_map {
		ctxt := dream.NewDreamContext(port_str, server_configs)
		ctxts = append(ctxts, ctxt)
	}

	// Server Loop
	for _, ctxt := range ctxts {
		go ctxt.RunDreamContext()
	}

	for {
		// Keep main alive
	}
}
