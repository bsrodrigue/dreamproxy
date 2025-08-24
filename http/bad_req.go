package http

import (
	"dreamproxy/logger"
	"fmt"

	"github.com/google/uuid"
)

func NewFailedToParseRes(remoteAddr string, msg string) *HttpRes {
	res := CreateHttpRes()
	res.Status = StatusBadRequest

	log := logger.NewRequestLog(logger.DREAM_SERVER, logger.ERROR, logger.REQ_PARSE_ERROR, msg)
	log.Request.ClientIP = remoteAddr

	// Create a log handler
	fmt.Println(log.ToText())
	return res
}

func NewBadRequestRes(req HttpReq, remoteAddr string, err error) *HttpRes {
	res := CreateHttpRes()
	res.Status = StatusBadRequest

	log := logger.NewRequestLog(logger.DREAM_SERVER, logger.ERROR, logger.BAD_REQUEST, err.Error())
	log.Request.ID = uuid.New().String()
	log.Request.Method = req.Method
	log.Request.Path = req.Target
	log.Request.ClientIP = remoteAddr
	log.Response.StatusCode = int(res.Status)
	log.Response.BytesSent = 0

	// Create a log handler
	fmt.Println(log.ToText())
	return res
}
