package maxim

import (
	"time"

	"github.com/gorilla/websocket"
)

// Session 是單個客戶端階段。
type Session struct {
	// store 是階段存儲資料。
	store map[string]interface{}
	// isClosed 表示此階段是否已經關閉了。
	isClosed bool
	// conn 是該階段的 WebSocket 連線。
	conn *websocket.Conn
	// engine 是此階段所屬的引擎。
	engine *Engine
}

// newSession 會在引擎中建立一個新的客戶端階段。
func (e *Engine) newSession(conn *websocket.Conn) *Session {
	return &Session{
		store:  make(map[string]interface{}),
		conn:   conn,
		engine: e,
	}
}

// errorAndClose 會在呼叫錯誤函式後進行關閉行為。
func (s *Session) errorAndClose(err error, c CloseStatus) error {
	s.Error(err)
	return s.Close(c)
}

// Get 能夠從客戶端階段中取得暫存資料。
func (s *Session) Get(k string) (v interface{}, ok bool) {
	v, ok = s.store[k]
	return
}

// MustGet 能夠從客戶端階段中取得暫存資料，如果該資料不存在則呼叫 `panic`。
func (s *Session) MustGet(k string) interface{} {
	v, ok := s.Get(k)
	if !ok {
		panic(ErrKeyNotFound)
	}
	return v
}

// GetString 能夠從客戶端階段中取得 `string` 型態的暫存資料。
func (s *Session) GetString(k string) string {
	v, ok := s.Get(k)
	if !ok {
		return ""
	}
	return v.(string)
}

// GetBool 能夠從客戶端階段中取得 `bool` 型態的暫存資料。
func (s *Session) GetBool(k string) bool {
	v, ok := s.Get(k)
	if !ok {
		return false
	}
	return v.(bool)
}

// GetDuration 能夠從客戶端階段中取得 `time.Duration` 型態的暫存資料。
func (s *Session) GetDuration(k string) time.Duration {
	v, ok := s.Get(k)
	if !ok {
		return time.Duration(0)
	}
	return v.(time.Duration)
}

// GetFloat64 能夠從客戶端階段中取得 `float64` 型態的暫存資料。
func (s *Session) GetFloat64(k string) float64 {
	v, ok := s.Get(k)
	if !ok {
		return 0
	}
	return v.(float64)
}

// GetInt 能夠從客戶端階段中取得 `int` 型態的暫存資料。
func (s *Session) GetInt(k string) int {
	v, ok := s.Get(k)
	if !ok {
		return 0
	}
	return v.(int)
}

// GetInt64 能夠從客戶端階段中取得 `int64` 型態的暫存資料。
func (s *Session) GetInt64(k string) int64 {
	v, ok := s.Get(k)
	if !ok {
		return 0
	}
	return v.(int64)
}

// GetStringMap 能夠從客戶端階段中取得 `map[string]interface{}` 型態的暫存資料。
func (s *Session) GetStringMap(k string) map[string]interface{} {
	v, ok := s.Get(k)
	if !ok {
		return nil
	}
	return v.(map[string]interface{})
}

// GetStringMapString 能夠從客戶端階段中取得 `map[string]string` 型態的暫存資料。
func (s *Session) GetStringMapString(k string) map[string]string {
	v, ok := s.Get(k)
	if !ok {
		return nil
	}
	return v.(map[string]string)
}

// GetStringSlice 能夠從客戶端階段中取得 `[]string` 型態的暫存資料。
func (s *Session) GetStringSlice(k string) []string {
	v, ok := s.Get(k)
	if !ok {
		return nil
	}
	return v.([]string)
}

// GetTime 能夠從客戶端階段中取得 `time.Time` 型態的暫存資料。
func (s *Session) GetTime(k string) time.Time {
	v, ok := s.Get(k)
	if !ok {
		return time.Time{}
	}
	return v.(time.Time)
}

// Close 會良好地結束與此客戶端的連線。
func (s *Session) Close(c CloseStatus) error {
	if s.isClosed {
		return ErrSessionClosed
	}
	s.isClosed = true
	err := s.conn.WriteControl(websocket.CloseMessage, websocket.FormatCloseMessage(int(c), ""), time.Now().Add(s.engine.config.WriteWait))
	if err != nil {
		return err
	}
	if s.engine.closeHandler != nil {
		s.engine.closeHandler(s, c, "")
	}
	if CloseStatus(c) == CloseNormalClosure {
		if s.engine.disconnectHandler != nil {
			s.engine.disconnectHandler(s)
		}
	}
	return s.conn.Close()
}

// Error 會呼叫錯誤處理函式並傳入此客戶階段，這並不會中斷連線。
func (s *Session) Error(err error) {
	if v, ok := err.(*websocket.CloseError); ok && v.Code == websocket.CloseNormalClosure {
		return
	}
	if s.engine.errorHandler != nil {
		s.engine.errorHandler(s, err)
	}
}

// IsClosed 會表示此客戶端階段是否已經關閉連線了。
func (s *Session) IsClosed() bool {
	return s.isClosed
}

// Set 能夠將指定的資料存儲到此客戶端階段中作為暫存快取。
func (s *Session) Set(k string, v interface{}) {
	s.store[k] = v
}

// Delete 會將指定資料從暫存快取中移除。
func (s *Session) Delete(k string) error {
	_, ok := s.store[k]
	if !ok {
		return ErrKeyNotFound
	}
	delete(s.store, k)
	return nil
}

// Write 能透將文字訊息寫入到客戶端中。
func (s *Session) Write(msg string) error {
	s.conn.SetWriteDeadline(time.Now().Add(s.engine.config.WriteWait))
	err := s.conn.WriteMessage(websocket.TextMessage, []byte(msg))
	return err
}

// WriteBinary 能透將二進制訊息寫入到客戶端中。
func (s *Session) WriteBinary(msg []byte) error {
	s.conn.SetWriteDeadline(time.Now().Add(s.engine.config.WriteWait))
	err := s.conn.WriteMessage(websocket.BinaryMessage, msg)
	return err
}

// Pong 能夠自主地回應客戶端一個 Pong 訊息，表示伺服器仍然有回應。
func (s *Session) Pong() error {
	return s.conn.WriteControl(websocket.PongMessage, []byte(``), time.Now().Add(s.engine.config.WriteWait))
}

// Ping 能夠詢問此客戶端的連線反應狀況，
// 如果在指定時間內沒有接收到 Pong 回應則會關閉並結束此連線。
func (s *Session) Ping() error {
	return s.conn.WriteControl(websocket.PingMessage, []byte(``), time.Now().Add(s.engine.config.WriteWait))
}
