package e

var MsgFlag = map[uint]string{
	SUCCESS:       "ok",
	ERROR:         "fail",
	InvalidParams: "请求参数错误",
}

// GetMsg
func GetMsg(code uint) string {
	msg, ok := MsgFlag[code]
	if ok {
		return msg
	}
	return MsgFlag[ERROR]
}
