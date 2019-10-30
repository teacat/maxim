package maxim

import (
	"errors"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

var (
	ErrEngineClosed     = errors.New("junipero: upgrading connections when engine closed")
	ErrSessionTimedOut  = errors.New("junipero: interacting with a timed out session")
	ErrConnectionClosed = errors.New("junipero: interacting with a disconnected connection")
	ErrSessionClosed    = errors.New("junipero: interacting with a closed session")
	ErrKeyNotFound      = errors.New("junipero: accessing a undefined key from the session store")
	ErrWriteTimedOut    = errors.New("junipero: write timed out")
	ErrSessionExists    = errors.New("junipero: write timed out")
	ErrSessionNotFound  = errors.New("junipero: write timed out")
)

// CloseStatus 是連線被關閉時的狀態代號。
type CloseStatus int

const (
	CloseNormalClosure           CloseStatus = 1000
	CloseGoingAway               CloseStatus = 1001
	CloseProtocolError           CloseStatus = 1002
	CloseUnsupportedData         CloseStatus = 1003
	CloseNoStatusReceived        CloseStatus = 1005
	CloseAbnormalClosure         CloseStatus = 1006
	CloseInvalidFramePayloadData CloseStatus = 1007
	ClosePolicyViolation         CloseStatus = 1008
	CloseMessageTooBig           CloseStatus = 1009
	CloseMandatoryExtension      CloseStatus = 1010
	CloseInternalServerErr       CloseStatus = 1011
	CloseServiceRestart          CloseStatus = 1012
	CloseTryAgainLater           CloseStatus = 1013
	CloseTLSHandshake            CloseStatus = 1015
)

type EngineOption func(*Engine)

// Engine 是 WebSocket 引擎。
type Engine struct {
	// sessions 是此引擎的所有階段連線。
	sessions map[int]*Session
	// config 是引擎的設置。
	config *EngineConfig
	// isClosed 表示此引擎是否已經被中止。
	isClosed bool

	// closeHandler
	closeHandler func(*Session, CloseStatus, string) error
	// connectHandler
	connectHandler func(*Session)
	// errorHandler
	errorHandler func(*Session, error)
	// messageHandler
	messageHandler func(*Session, string)
	// messageBinaryHandler
	messageBinaryHandler func(*Session, []byte)
	// pingHandler
	pingHandler func(*Session)
	// pongHandler
	pongHandler func(*Session)
	// requestHandler
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

// WithCloseHandler
func WithCloseHandler(h func(*Session, CloseStatus, string) error) EngineOption {
	return func(e *Engine) {
		e.closeHandler = h
	}
}

// WithConnectHandler
func WithConnectHandler(h func(*Session)) EngineOption {
	return func(e *Engine) {
		e.connectHandler = h
	}
}

// WithErrorHandler
func WithErrorHandler(h func(*Session, error)) EngineOption {
	return func(e *Engine) {
		e.errorHandler = h
	}
}

// WithMessageHandler
func WithMessageHandler(h func(*Session, string)) EngineOption {
	return func(e *Engine) {
		e.messageHandler = h
	}
}

// WithMessageBinaryHandler
func WithMessageBinaryHandler(h func(*Session, []byte)) EngineOption {
	return func(e *Engine) {
		e.messageBinaryHandler = h
	}
}

// WithPingHandler
func WithPingHandler(h func(*Session)) EngineOption {
	return func(e *Engine) {
		e.pingHandler = h
	}
}

// WithPongHandler
func WithPongHandler(h func(*Session)) EngineOption {
	return func(e *Engine) {
		e.pongHandler = h
	}
}

// WithRequestHandler
func WithRequestHandler(h func(http.ResponseWriter, *http.Request, *Session)) EngineOption {
	return func(e *Engine) {
		e.requestHandler = h
	}
}

// WithConfig
func WithConfig(conf *EngineConfig) EngineOption {
	return func(e *Engine) {
		e.config = conf
	}
}

// NewServer 會建立一個新的 WebSocket 伺服器。
func NewServer(options ...EngineOption) *Engine {
	e := &Engine{
		config:   DefaultConfig(),
		sessions: make(map[int]*Session),
	}
	for _, v := range options {
		v(e)
	}
	return e
}

// DefaultConfig 會回傳一個新的預設引擎設置。
func DefaultConfig() *EngineConfig {
	return &EngineConfig{
		WriteWait:      30,
		PongWait:       10,
		PingPeriod:     20,
		MaxMessageSize: 10 * 1024 * 1024,
		Upgrader: &websocket.Upgrader{
			HandshakeTimeout: 30 * time.Second,
			ReadBufferSize:   1024,
			WriteBufferSize:  1024,
		},
	}
}

// Handler 是用以傳入 HTTP 伺服器協助升級與接收 WebSocket 相關資訊的最重要函式。
func (e *Engine) Handler() func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if e.isClosed {
			panic(ErrEngineClosed)
		}
		c, err := e.config.Upgrader.Upgrade(w, r, nil)
		s := e.NewSession(c)
		if err != nil {
			if e.errorHandler != nil {
				e.errorHandler(s, err)
			}
			return
		}
		if e.requestHandler != nil {
			e.requestHandler(w, r, s)
		}

		c.SetPingHandler(func(m string) error {
			e.handler.Ping(s)
			s.Pong()
			return nil
		})
		c.SetPongHandler(func(m string) error {
			e.handler.Pong(s)
			return nil
		})
		c.SetCloseHandler(func(code int, msg string) error {
			s.Close()
			e.handler.Close(s, CloseStatus(code), msg)
			if CloseStatus(code) == CloseNormalClosure {
				e.handler.Disconnect(s)
			}
			return nil
		})

		e.handler.Connect(s)

		defer func() {
			// close handler?
			s.Close()
		}()

		for {
			typ, msg, err := c.ReadMessage()
			if err != nil {
				if !s.isClosed {
					s.Close()
					e.handler.Error(s, err)
				}
				break
			}
			switch typ {
			case websocket.TextMessage:
				e.handler.Message(s, string(msg))
				break
			case websocket.BinaryMessage:
				e.handler.MessageBinary(s, msg)
				break
			}
		}
	}
}

// Close 會關閉整個引擎並中斷所有連線。
func (e *Engine) Close() {
	for _, v := range e.sessions {
		v.Close()
	}
	e.isClosed = true
}

// IsClosed 會表示該引擎是否已經關閉了。
func (e *Engine) IsClosed() bool {
	return e.isClosed
}

// Len 會取得正在連線的客戶端總數。
func (e *Engine) Len() int {
	return len(e.sessions)
}
