package e

var MsgFlag = map[uint]string{
	Success:       "ok",
	Error:         "failed",
	InvalidParams: "请求参数错误",
}

// GetMsg
func GetMsg(code uint) string {
	msg, ok := MsgFlag[code]
	if ok {
		return msg
	}
	return MsgFlag[Error]
}
