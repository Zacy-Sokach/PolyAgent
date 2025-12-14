package tui

// ModelState 临时模型状态接口
type ModelState interface {
	GetMessages() []Message
	AddMessage(role, content string)
	GetAPIMessages() interface{} // 临时使用interface{}
	AddAPIMessage(msg interface{})
	IsThinking() bool
	SetThinking(thinking bool)
	GetCurrentResponse() string
	SetCurrentResponse(resp string)
	GetCurrentThinking() string
	SetCurrentThinking(think string)
	SaveHistory()
	GetViewport() interface{} // 临时使用interface{}
	SetViewport(vp interface{})
	GetTextarea() interface{} // 临时使用interface{}
	SetTextarea(ta interface{})
	IsReady() bool
	SetReady(ready bool)
}