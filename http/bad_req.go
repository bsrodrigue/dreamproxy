package http

func NewFailedToParseRes(remoteAddr string, msg string) *HttpRes {
	res := CreateHttpRes()
	res.Status = StatusBadRequest

	return res
}

func NewBadRequestRes(req HttpReq, remoteAddr string, err error) *HttpRes {
	res := CreateHttpRes()
	res.Status = StatusBadRequest

	return res
}
