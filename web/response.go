package web

type ResponseMessage struct {
	Code    int         `json:"code"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
	Debug   interface{} `json:"debug,omitempty"`
}

func NewResp(code int, message string, payload ...interface{}) *ResponseMessage {
	if len(payload) > 0 {
		return &ResponseMessage{code, message, payload[0], nil}
	} else {
		return &ResponseMessage{code, message, nil, nil}
	}
}

func NewDebugResp(code int, message string, debug interface{}, payload ...interface{}) *ResponseMessage {
	if len(payload) > 0 {
		return &ResponseMessage{code, message, payload[0], debug}
	} else {
		return &ResponseMessage{code, message, nil, debug}
	}
}
