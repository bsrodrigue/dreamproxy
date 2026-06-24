package http

func NewFailedToParseRes(remoteAddr string, msg string) *HTTPRes {
	res := CreateHTTPRes()
	res.Status = StatusBadRequest

	return res
}

func NewBadRequestRes(req HTTPReq, remoteAddr string, err error) *HTTPRes {
	res := CreateHTTPRes()
	res.Status = StatusBadRequest

	return res
}
