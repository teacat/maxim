package maxim

import (
	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

// Client 呈現了一個 WebSocket 客戶端。
type Client struct {
	// Config 是客戶端設置。
	config *ClientConfig
	// conn 是底層的 WebSocket 連線。
	conn *websocket.Conn
	// isClosed 會表示此客戶端是否已經關閉連線了。
	isClosed bool
}

// ClientConfig 是客戶端設置。
type ClientConfig struct {
	// Address 是遠端伺服器位置（如：`ws://127.0.0.1:1234/echo`）。
	Address string
	// header 是 WebSocket 初次發送時順帶傳輸的 HTTP 標頭資訊。
	Header http.Header
	// WriteWait 是每次訊息寫入時的逾時時間。
	WriteWait time.Duration
}

// NewClient 會建立客戶端並連線到指定的 WebSocket 伺服端。
func NewClient(conf *ClientConfig) (*Client, *http.Response, error) {
	if conf.WriteWait == 0 {
		conf.WriteWait = time.Second * 30
	}
	conn, resp, err := websocket.DefaultDialer.Dial(conf.Address, conf.Header)
	if err != nil {
		return nil, resp, err
	}
	client := &Client{
		config: conf,
		conn:   conn,
	}
	go client.listener()
	return client, resp, nil
}

// listener 會持續監聽一些額外的訊息並自動回應。
func (c *Client) listener() {
	for {
		typ, msg, err := c.readAll()
		if err != nil {
			return
		}
		switch typ {
		case websocket.PingMessage:
			log.Printf("received PING!!! %v", msg)
			c.conn.WriteControl(websocket.PongMessage, msg, time.Now().Add(c.config.WriteWait))
		}
	}
}

// readAll 會阻塞程式直到有訊息為止，
// 這會接收到所有訊息像是 Ping-Pong 與 Close 或標準的文字甚至二進制訊息。
func (c *Client) readAll() (int, []byte, error) {
	if c.isClosed {
		return 0, []byte(``), ErrClientClosed
	}
	typ, msg, err := c.conn.ReadMessage()
	if err != nil {
		return typ, msg, err
	}
	log.Printf("%+v, %+v, %+v\n", typ, msg, err)
	return typ, msg, nil
}

// ReadMessage 會阻塞程式直到有訊息為止，接收到的訊息會 `string` 字串標準訊息。
// 任何系統訊息像是 Ping-Pong 與 Close 都不會出現在這裡。
func (c *Client) Read() (string, error) {
	if c.isClosed {
		return "", ErrClientClosed
	}
	for {
		typ, msg, err := c.readAll()
		if err != nil {
			return "", err
		}
		if typ != websocket.TextMessage {
			continue
		}
		return string(msg), nil
	}
}

// ReadBinary 會阻塞程式直到有訊息為止，接收到的訊息會是 `[]byte` 二進制標準訊息。
// 任何系統訊息像是 Ping-Pong 與 Close 都不會出現在這裡。
func (c *Client) ReadBinary() ([]byte, error) {
	if c.isClosed {
		return []byte(``), ErrClientClosed
	}
	for {
		typ, msg, err := c.readAll()
		if err != nil {
			return []byte(``), err
		}
		if typ != websocket.BinaryMessage {
			continue
		}
		return msg, nil
	}
}

// Close 會依照正常手續告訴伺服器關閉並結束客戶端連線。
func (c *Client) Close() error {
	if c.isClosed {
		return ErrClientClosed
	}
	c.isClosed = true
	return c.conn.WriteControl(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""), time.Now().Add(c.config.WriteWait))
}

// Write 能夠傳送文字訊息至伺服端。
func (c *Client) Write(msg string) error {
	if c.isClosed {
		return ErrClientClosed
	}
	return c.conn.WriteMessage(websocket.TextMessage, []byte(msg))
}

// WriteBinary 能夠傳送二進制訊息至伺服端。
func (c *Client) WriteBinary(msg []byte) error {
	if c.isClosed {
		return ErrClientClosed
	}
	return c.conn.WriteMessage(websocket.BinaryMessage, msg)
}

// IsClosed 會表示該連線是否已經關閉並結束了。
func (c *Client) IsClosed() bool {
	return c.isClosed
}
