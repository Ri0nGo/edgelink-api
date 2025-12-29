package bizErr

type BizError struct {
	Msg string
}

func (e *BizError) Error() string {
	return e.Msg
}

func NewBizError(msg string) *BizError {
	return &BizError{Msg: msg}
}
