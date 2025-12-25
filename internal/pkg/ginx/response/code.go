package response

type RespCode int

var msgMap = map[RespCode]string{
	RespCodeSuccess:     "success",
	RespCodeInternalErr: "internal error",
	RespCodeParamErr:    "param error",
}

func (c RespCode) msg() string {
	if msg, ok := msgMap[c]; ok {
		return msg
	}
	return "unknown error"
}

const (
	RespCodeSuccess     RespCode = 0
	RespCodeInternalErr RespCode = 500
	RespCodeParamErr    RespCode = 10001
)
