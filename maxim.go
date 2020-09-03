package maxim

import (
	"errors"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

var (
	// ErrEngineClosed 表示引擎已經關閉了，但卻要繼續升級新的連線。
	ErrEngineClosed = errors.New("maxim: 引擎已經關閉而導致無法升級連線")
	// ErrClientClosed 表示客戶端已經與遠端引擎結束連線，但卻仍要繼續執行操作。
	ErrClientClosed = errors.New("maxim: 客戶端已經關閉連線但卻繼續操作")
	// ErrSessionClosed 表示正在跟已經結束連線的階段進行互動。
	ErrSessionClosed = errors.New("maxim: 連線階段已經關閉連線但卻繼續操作")
	// ErrKeyNotFound 表示無法在連線階段的存儲空間中找到指定的鍵值資料。
	ErrKeyNotFound = errors.New("maxim: 無法在連線階段中找到指定鍵值資料")
	// ErrDuplicatedSession 表示水桶裡已經有相同的階段了。
	ErrDuplicatedSession = errors.New("maxim: 欲在指定水桶中放入重複的連線階段")
	// ErrSessionNotFound 表示刪除一個水桶裡不存在的連線階段。
	ErrSessionNotFound = errors.New("maxim: 找不到指定的連線階段")
)

// CloseStatus 是連線被關閉時的狀態代號。
type CloseStatus int

const (
	// CloseNormalClosure 表示正常關閉。
	CloseNormalClosure CloseStatus = 1000
	// CloseGoingAway 表示因為瀏覽器結束或是故障而離去。
	CloseGoingAway CloseStatus = 1001
	// CloseProtocolError 表示協定錯誤而無法連線。
	CloseProtocolError CloseStatus = 1002
	// CloseUnsupportedData 表示接收到無法處理的資料型態（如：文字處理函式卻接收到二進制資料）而錯誤。
	CloseUnsupportedData CloseStatus = 1003
	// CloseNoStatusReceived 表示沒有接收到狀態代碼。
	CloseNoStatusReceived CloseStatus = 1005
	// CloseAbnormalClosure 表示發生意外錯誤。
	CloseAbnormalClosure CloseStatus = 1006
	// CloseInvalidFramePayloadData 表示錯誤的資料幀。
	CloseInvalidFramePayloadData CloseStatus = 1007
	// ClosePolicyViolation 表示接收到違反原則的訊息而中止。
	ClosePolicyViolation CloseStatus = 1008
	// CloseMessageTooBig 表示接收的資料因過於龐大而拒絕。
	CloseMessageTooBig CloseStatus = 1009
	// CloseMandatoryExtension 表示伺服器必須交涉擴充功能，可能是缺少部份功能而結束。
	CloseMandatoryExtension CloseStatus = 1010
	// CloseInternalServerErr 表示伺服器內部發生錯誤。
	CloseInternalServerErr CloseStatus = 1011
	// CloseServiceRestart 表示服務為了重啟而中斷。
	CloseServiceRestart CloseStatus = 1012
	// CloseTryAgainLater 表示目前此連線不可能，稍後嘗試也許可行。
	CloseTryAgainLater CloseStatus = 1013
	// CloseTLSHandshake 表示這個連線是為了 TLS 握手協定才產生的。
	CloseTLSHandshake CloseStatus = 1015
)

// Handler 是一個引擎的處理界面。
type Handler interface {
	// HandleMessage 會將傳入的函式作為收到字串訊息時的處理函式。
	HandleMessage(*Session, string)
	// HandleMessageBinary 會將傳入的函式作為收到二進制訊息時的處理函式。
	HandleMessageBinary(*Session, []byte)
	// HandleError 會將傳入的函式作為發生錯誤時的處理函式。
	HandleError(*Session, error)
	// HandleClose 會將傳入的函式作為連線關閉時的處理函式，無論連線是怎麼關閉都會呼叫此函式。
	HandleClose(*Session, CloseStatus, string) error
	// HandleDisconnect 會將傳入的函式作為正常連線關閉時的處理函式。
	HandleDisconnect(*Session)
	// HandleConnect 會將傳入的函式作為連線建立時的處理函式。
	HandleConnect(*Session)
}

// Engine 是 WebSocket 引擎。
type Engine struct {
	// sessions 是此引擎的所有階段連線。
	sessions *Bucket
	// config 是引擎的設置。
	config *EngineConfig
	// isClosed 表示此引擎是否已經被中止。
	isClosed bool

	// closeHandler 是連線關閉時的處理函式，無論連線是怎麼關閉都會呼叫此函式。
	closeHandler func(*Session, CloseStatus, string) error
	// connectHandler 是連線建立時的處理函式。
	connectHandler func(*Session)
	// disconnectHandler 是正常連線關閉時的處理函式。
	disconnectHandler func(*Session)
	// errorHandler 是發生錯誤時的處理函式。
	errorHandler func(*Session, error)
	// messageHandler 是收到字串訊息時的處理函式。
	messageHandler func(*Session, string)
	// messageBinaryHandler 是收到二進制訊息時的處理函式。
	messageBinaryHandler func(*Session, []byte)
	// pongHandler 是收到 `PONG` 通知訊息的處理函式。
	pongHandler func(*Session)
	// requestHandler 是每個升級請求的監聽函式，這沒辦法改變程式流程。
	requestHandler func(http.ResponseWriter, *http.Request, *Session)
}

// EngineConfig 是引擎選項設置。
type EngineConfig struct {
	// WriteWait 是到逾時之前的等待時間。
	WriteWait time.Duration
	// PongWait 是等待 Pong 回應的時間，在指定時間內客戶端如果沒有任何響應，該 WebSocket 連線則會被迫中止。
	// 設置為 `0` 來停用無響應自動斷線的功能。
	PongWait time.Duration
	// PingPeriod 是 Ping 的週期時間。
	PingPeriod time.Duration
	// MaxMessageSize 是最大可接收的訊息位元組大小，
	// 溢出此大小的訊息會被拋棄。
	MaxMessageSize int64
	// Upgrader 是 WebSocket 升級的相關設置。
	Upgrader *websocket.Upgrader
}

// New 會建立一個新的 WebSocket 伺服器。
func New(conf *EngineConfig) *Engine {
	return &Engine{
		config:   conf,
		sessions: NewBucket(&BucketConfig{}),
	}
}

// NewDefault 會初始化一個帶有預設設置的引擎。
func NewDefault() *Engine {
	return New(DefaultConfig())
}

// DefaultConfig 會回傳一個新的預設引擎設置。
func DefaultConfig() *EngineConfig {
	return &EngineConfig{
		WriteWait:      time.Second * 10,
		PongWait:       time.Second * 60,
		PingPeriod:     time.Second * 54,
		MaxMessageSize: 4 * 1024 * 1024,
		Upgrader: &websocket.Upgrader{
			HandshakeTimeout: 30 * time.Second,
			ReadBufferSize:   1024,
			WriteBufferSize:  1024,
		},
	}
}

// Handle 能夠接收一個處理界面，用來處理所有動作。這會覆蓋先前指定的 `HandleMessage`…等所指定的處理函式。
func (e *Engine) Handle(h Handler) {
	e.HandleMessage(h.HandleMessage)
	e.HandleMessageBinary(h.HandleMessageBinary)
	e.HandleError(h.HandleError)
	e.HandleClose(h.HandleClose)
	e.HandleDisconnect(h.HandleDisconnect)
	e.HandleConnect(h.HandleConnect)
}

// HandleMessage 會將傳入的函式作為收到字串訊息時的處理函式。
func (e *Engine) HandleMessage(h func(*Session, string)) {
	e.messageHandler = h
}

// HandleMessageBinary 會將傳入的函式作為收到二進制訊息時的處理函式。
func (e *Engine) HandleMessageBinary(h func(*Session, []byte)) {
	e.messageBinaryHandler = h
}

// HandleError 會將傳入的函式作為發生錯誤時的處理函式。
func (e *Engine) HandleError(h func(*Session, error)) {
	e.errorHandler = h
}

// HandleClose 會將傳入的函式作為連線關閉時的處理函式，無論連線是怎麼關閉都會呼叫此函式。
func (e *Engine) HandleClose(h func(*Session, CloseStatus, string) error) {
	e.closeHandler = h
}

// HandleDisconnect 會將傳入的函式作為正常連線關閉時的處理函式。
func (e *Engine) HandleDisconnect(h func(*Session)) {
	e.disconnectHandler = h
}

// HandleConnect 會將傳入的函式作為連線建立時的處理函式。
func (e *Engine) HandleConnect(h func(*Session)) {
	e.connectHandler = h
}

// HandleRequest 是用以傳入 HTTP 伺服器協助升級與接收 WebSocket 相關資訊的最重要函式。
func (e *Engine) HandleRequest(w http.ResponseWriter, r *http.Request) {
	if e.isClosed {
		panic(ErrEngineClosed)
	}
	c, err := e.config.Upgrader.Upgrade(w, r, nil)
	c.SetReadLimit(e.config.MaxMessageSize)
	c.SetReadDeadline(time.Now().Add(e.config.PongWait))
	s := e.newSession(c)
	if err != nil {
		s.Error(err)
		return
	}
	err = e.sessions.Put(s)
	if err != nil {
		s.Error(err)
		return
	}
	if e.requestHandler != nil {
		e.requestHandler(w, r, s)
	}
	c.SetCloseHandler(func(code int, msg string) error {
		return s.Close(CloseStatus(code))
	})
	c.SetPongHandler(func(msg string) error {
		c.SetReadDeadline(time.Now().Add(e.config.PongWait))
		return nil
	})

	if e.connectHandler != nil {
		e.connectHandler(s)
	}

	ticker := time.NewTicker(e.config.PingPeriod)

	defer func() {
		ticker.Stop()
		s.Close(CloseNormalClosure)
	}()

	go func() {
		for {
			<-ticker.C
			if e.isClosed || s.isClosed {
				break
			}
			err := s.Ping()
			if err != nil {
				s.errorAndClose(err, CloseAbnormalClosure)
			}
		}
	}()

	for {
		if e.isClosed || s.isClosed {
			break
		}
		typ, msg, err := c.ReadMessage()
		if err != nil {
			s.errorAndClose(err, CloseAbnormalClosure)
			break
		}
		switch typ {
		case websocket.TextMessage:
			if e.messageHandler != nil {
				e.messageHandler(s, string(msg))
			}
		case websocket.BinaryMessage:
			if e.messageBinaryHandler != nil {
				e.messageBinaryHandler(s, msg)
			}
		}
	}
}

// Write 能夠將文字訊息寫入到所有客戶端。
func (e *Engine) Write(msg string) error {
	return e.sessions.Write(msg)
}

// WriteFilter 能夠將文字訊息寫入到被篩選的客戶端。
func (e *Engine) WriteFilter(msg string, fn func(*Session) bool) error {
	return e.sessions.WriteFilter(msg, fn)
}

// WriteOthers 能夠將文字訊息寫入到指定以外的所有客戶端。
func (e *Engine) WriteOthers(msg string, s *Session) error {
	return e.sessions.WriteOthers(msg, s)
}

// WriteBinary 能夠將二進制訊息寫入到所有客戶端。
func (e *Engine) WriteBinary(msg []byte) error {
	return e.sessions.WriteBinary(msg)
}

// WriteBinaryFilter 能夠將二進制訊息寫入到被篩選客戶端。
func (e *Engine) WriteBinaryFilter(msg []byte, fn func(*Session) bool) error {
	return e.sessions.WriteBinaryFilter(msg, fn)
}

// WriteBinaryOthers 能夠將二進制訊息寫入到指定以外的所有客戶端。
func (e *Engine) WriteBinaryOthers(msg []byte, s *Session) error {
	return e.sessions.WriteBinaryOthers(msg, s)
}

// Close 會關閉整個引擎並中斷所有連線。
func (e *Engine) Close() {
	e.isClosed = true
	e.sessions.Close(CloseNormalClosure)
}

// IsClosed 會表示該引擎是否已經關閉了。
func (e *Engine) IsClosed() bool {
	return e.isClosed
}

// Len 會取得正在連線的客戶端總數。
func (e *Engine) Len() int {
	return e.sessions.Len()
}
