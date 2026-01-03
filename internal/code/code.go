package code

type ErrorCode int

const (
	ErrSuccess                     ErrorCode = 0  // 成功
	ErrError                       ErrorCode = 1  // 错误
	ErrInvalidParam                ErrorCode = 2  // 无效参数
	ErrUnauthorized                ErrorCode = 3  // 未授权
	ErrForbidden                   ErrorCode = 4  // 禁止访问
	ErrNotFound                    ErrorCode = 5  // 未找到
	ErrInternalServerError         ErrorCode = 6  // 内部服务器错误
	ErrBadRequest                  ErrorCode = 7  // 坏请求
	ErrUnprocessableEntity         ErrorCode = 8  // 不可处理实体
	ErrTooManyRequests             ErrorCode = 9  // 请求太多
	ErrUsernameOrPasswordIncorrect ErrorCode = 10 // 用户名或密码错误
)

var errorCodeMap = map[ErrorCode]string{
	ErrSuccess:                     "success",
	ErrError:                       "error",
	ErrInvalidParam:                "invalid param",
	ErrUnauthorized:                "unauthorized",
	ErrForbidden:                   "forbidden",
	ErrNotFound:                    "not found",
	ErrInternalServerError:         "internal server error",
	ErrBadRequest:                  "bad request",
	ErrUnprocessableEntity:         "unprocessable entity",
	ErrTooManyRequests:             "too many requests",
	ErrUsernameOrPasswordIncorrect: "username or password incorrect",
}

func (e ErrorCode) errorMessage() string {
	if msg, ok := errorCodeMap[e]; ok {
		return msg
	}
	return "unknown error"
}

// CodeError defines the structured error for REST API
type CodeError struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
}

func (e *CodeError) Error() string {
	return e.Msg
}

// New returns a structured error
func New(errCode ErrorCode) error {
	return &CodeError{
		Code: int(errCode),
		Msg:  errCode.errorMessage(),
	}
}

// NewMsg returns a structured error with custom message
func NewMsg(errCode ErrorCode, msg string) error {
	return &CodeError{
		Code: int(errCode),
		Msg:  msg,
	}
}
