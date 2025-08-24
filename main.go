package main

import (
	"dreamproxy/config"
	"dreamproxy/dream"
	"dreamproxy/logger"
	"fmt"
	_ "net/http/pprof"
	"strconv"
)

const PORT string = "8080"
const ROOT_FS string = "staticfiles"
const LOG_FILE string = "/var/log/dreamserver/access.log"
const LOG_FORMAT string = "text"
const CONFIG_FILE string = "./Dreamfile"

var dreamconfig config.Config

func WriteLog(log logger.RequestLog) {
	if LOG_FORMAT == "text" {
		fmt.Println(log.ToText())

	}
}

func main() {
	ctxts := []dream.DreamContext{}
	dreamconfig = config.LoadDreamFile(CONFIG_FILE)

	for _, server_config := range dreamconfig.Servers {
		ctxt := dream.NewDreamContext(strconv.Itoa(server_config.Listen.Port), []config.Server{})
		ctxts = append(ctxts, ctxt)
	}

	// Server Loop
	for _, ctxt := range ctxts {
		ctxt.RunDreamContext()
	}
}
