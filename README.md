# Maxim [![GoDoc](https://godoc.org/github.com/teacat/maxim?status.svg)](https://godoc.org/github.com/teacat/maxim) [![Coverage Status](https://coveralls.io/repos/github/teacat/maxim/badge.svg?branch=master)](https://coveralls.io/github/teacat/maxim?branch=master) [![Build Status](https://travis-ci.com/teacat/maxim.svg?branch=master)](https://travis-ci.com/teacat/maxim) [![Go Report Card](https://goreportcard.com/badge/github.com/teacat/maxim)](https://goreportcard.com/report/github.com/teacat/maxim)

Golang 的 WebSocket 伺服與客戶端。

# 索引

* [安裝方式](#安裝方式)
* [使用方式](#使用方式)
    * [伺服端](#伺服端)
        * [監聽事件與訊息](#監聽事件與訊息)
		* [廣播寫入訊息](#廣播寫入訊息)
        * [連線階段](#連線階段)
            * [寫入訊息](#寫入訊息)
            * [鍵值存儲庫](#鍵值存儲庫)
            * [觸發錯誤](#觸發錯誤)
            * [Ping/Pong](#Ping-Pong)
            * [關閉連線](#關閉連線)
        * [連線階段水桶](#連線階段水桶)
        * [關閉引擎](#關閉引擎)
    * [客戶端](#客戶端)
        * [接收訊息](#接收訊息)
        * [寫入訊息](#寫入訊息)
        * [關閉連線](#關閉連線)

# 安裝方式

打開終端機並且透過 `go get` 安裝此套件即可。

```bash
$ go get github.com/teacat/maxim
```

# 使用方式

## 伺服端

使用 `New` 初始化一個未設置的引擎，或是 `NewDefault` 來在初始化時直接套用 Maxim 預設的偏好設定。引擎設置好後將其以 `HandleRequest` 公開在 HTTP 伺服器的路由中就能讓客戶端連線。

```go
func main() {
	m := maxim.NewDefault()
	// 監聽任何來自客戶端的文字訊息。
	m.HandleMessage(func(s *maxim.Session, msg string) {
		log.Println("已接收：" + msg)
	})
	// 客戶端現在可以透過 `localhost:8080/ws` 連線到 Maxim 的 WebSocket 引擎。
	http.HandleFunc("/ws", m.HandleRequest)
	http.ListenAndServe(":8080", nil)
}
```

### 監聽事件與訊息

你可以掌握大部分的 WebSocket 事件，例如：連線、斷線與訊息接收…等。這些函式通常被歸類在 `Handle...`。

```go
func main() {
	m := maxim.NewDefault()
	// 文字訊息事件，收到客戶端的文字訊息就會呼叫此處理函式。
	m.HandleMessage(func(s *maxim.Session, msg string) {
		log.Println("收到文字訊息：" + msg)
	})
	// 監聽連線事件，任何客戶端一旦連線就會呼叫此處理函式。
	m.HandleConnect(func(s *maxim.Session) {
		log.Printf("已連線：%+v", s)
	})
	// 監聽連線關閉事件，只要連線關閉就會呼叫此處理函式。
	m.HandleClose(func(s *maxim.Session, status maxim.CloseStatus, msg string) error {
		log.Printf("連線關閉：%+v, %+v, %+v", s, status, msg)
		return nil
	})
	// 監聽標準離線事件，只有在連線正常表明離線才會呼叫此處理函式。
	m.HandleDisconnect(func(s *maxim.Session) {
		log.Printf("正常離線：%+v", s)
	})
	// 監聽錯誤事件，若連線發生錯誤就會呼叫此處理函式，異常的斷線也會。
	m.HandleError(func(s *maxim.Session, err error) {
		log.Printf("錯誤：%+v, %+v", s, err)
	})
	// ...
}
```

倘若有一個良好的結構，你也可以傳入一個 `Handler` 的實作介面來一次處理所有的事件。

```go
// MyWebSocketHandler 會實作 maxim.Handler 來處理所有 WebSocket 事件。
type MyWebSocketHandler struct {
}

func (h MyWebSocketHandler) HandleMessage(*Session, string)                  {}
func (h MyWebSocketHandler) HandleMessageBinary(*Session, []byte)            {}
func (h MyWebSocketHandler) HandleError(*Session, error)                     {}
func (h MyWebSocketHandler) HandleClose(*Session, CloseStatus, string) error {}
func (h MyWebSocketHandler) HandleDisconnect(*Session)                       {}
func (h MyWebSocketHandler) HandleConnect(*Session)                          {}

func main() {
	m := maxim.NewDefault()
	// 讓 MyWebSocketHandler 處理所有 WebSocket 事件
	h := MyWebSocketHandler{}
	m.Handle(h)

	// ...
}
```

### 廣播寫入訊息

你可以直接對引擎呼叫 `Write` 或 `WriteBinary` 來向所有客戶端寫入訊息。

```go
func main() {
	m := maxim.NewDefault()
	m.HandleConnect(func(_ *maxim.Session) {
		// 注意，這是呼叫 `m` 來向所有連線階段發送訊息。
		m.Write("有新的人加入啦！")
	})
	// ...
}
```

### 連線階段

每個連線到 Maxim 的 WebSocket 連線都會成為一個 `*maxim.Session` 連線階段。這讓你可以個別管理每個客戶端連線。

```go
func main() {
	m := maxim.NewDefault()
	// 監聽連線事件，任何客戶端一旦連線就會呼叫此處理函式。
	m.HandleConnect(func(s *maxim.Session) {
		// `s` 即是單個連線階段。
		log.Printf("已連線：%+v", s)
	})
	// ...
}
```

#### 寫入訊息

透過 `Write` 或 `WriteBinray` 來向指定客戶端發送訊息。

```go
func main() {
	m := maxim.NewDefault()
	m.HandleMessage(func(s *maxim.Session, msg string) {
		// 發送文字訊息給此連線階段。
		s.Write("Hello, world!")
	})
	// ...
}
```

### 鍵值存儲庫

每個連線階段都有自己的鍵值存儲庫，你可以在連線階段中保存資料，用以在不同訊息、請求交互傳遞資料。使用 `Set` 來儲存資料、`Get` 來取得；而 `Delete` 即為刪除某個指定的鍵值資料。

```go
func main() {
	m := maxim.NewDefault()
	m.HandleMessage(func(s *maxim.Session, msg string) {
		// 在此連線階段儲存一個名為 `account` 的字串資料。
		s.Set("account", msg)
		// 從此連線階段中拿取名為 `account` 的字串資料。
		// 若找不到此資料則會回傳一個空字串（零值）。
		log.Println(s.GetString("account"))
	})
	// ...
}
```

#### 觸發錯誤

若某個連線階段內發生錯誤，能夠以指定連線階段呼叫引擎的錯誤處理函式。使用 `Error` 來將錯誤傳遞給引擎的錯誤處理函式。

```go
func main() {
	m := maxim.NewDefault()
	m.HandleError(func(s *maxim.Session, err error) {
		log.Printf("發生錯誤：%+v, %+v", s, err)
	})
	m.HandleConnect(func(s *maxim.Session) {
		// 手動呼叫錯誤處理函式。
		s.Error(errors.New("這個連線階段發生錯誤啦！"))
	})
	// ...
}
```

#### Ping/Pong

透過 `Ping` 可以要求客戶端傳遞一個 `Pong` 訊息來確認客戶端仍有反應且在線上。也能透過 `Pong` 來主動告知客戶端伺服器仍有反應。

Maxim 引擎預設會自動間隔一段時間傳遞 `Ping` 至客戶端，這部份能夠在 `EngineConfig` 中更改。

```go
func main() {
	m := maxim.NewDefault()
	m.HandleConnect(func(s *maxim.Session) {
		// 詢問客戶端是否在線上。
		s.Ping()
	})
	// ...
}
```

#### 關閉連線

若要結束與客戶端的連線則可以使用 `Close` 來正常關閉。

```go
func main() {
	m := maxim.NewDefault()
	m.HandleMessage(func(s *maxim.Session, msg string) {
		s.Close()
	})
	// ...
}
```

由於鍵值存儲庫能夠儲存許多不同的資料型態內容，因此可以使用 `GetInt`、`GetStringMap` 等多樣的函式來在取得時就直接轉換資料型態而非單純的 `interface{}`。

### 連線階段水桶

`NewBucket` 可以初始化一個連線階段水桶，用來建立一個群組以放入多個連線階段共同管理、傳遞訊息。

```go
func main() {
	m := maxim.NewDefault()
	// 初始化一個新的連線階段水桶。
	b := maxim.NewBucket(&maxim.BucketConfig{})

	// 監聽任何新的客戶端連線，並在連線之後將客戶端放入同個連線階段水桶中。
	m.HandleConnect(func(s *maxim.Session) {
		b.Put(s)
		// 對水桶裡的所有連線階段客戶端發送群組訊息。
		b.Write("Hello, everyone!")
	})
	// ...
}
```

連線階段水桶著重在群發訊息的功能。你可以透過 `WriteFilter` 來篩選不希望發送的指定客戶端，或是以 `WriteOthers` 來發送給指定客戶端以外的所有連線。亦能透過 `Close` 批次關閉位於相同水桶的客戶端。

### 關閉引擎

使用 `Close` 來關閉引擎並結束 WebSocket 連線。

```go
func main() {
	m := maxim.NewDefault()
	m.Close()
}
```

## 客戶端

透過 `NewClient` 並傳入一個 `ClientConfig` 客戶端設置來初始化並直接連線到遠端 WebSocket 伺服器。

```go
func main() {
	c, _, _ := maxim.NewClient(&maxim.ClientConfig{
		Address: "ws://localhost:8080/ws",
	})
}
```

### 接收訊息

使用 `Read` 或 `ReadBinary` 會將執行緒堵塞至收到伺服端傳送的訊息為止。

```go
func main() {
	c, _, _ := maxim.NewClient(&maxim.ClientConfig{
		Address: "ws://localhost:8080/ws",
	})
	// `Read` 會接收文字訊息，若要接收二進制位元組訊息請用 `ReadBinary`
	msg, _ := c.Read()
	log.Println("received: " + msg)
}
```

### 寫入訊息

透過 `Write` 或 `WriteBinray` 來向伺服器發送訊息。

```go
func main() {
	c, _, _ := maxim.NewClient(&maxim.ClientConfig{
		Address: "ws://localhost:8080/ws",
	})
	c.Write("Hello, world!")
}
```

### 關閉連線

若要結束客戶端與伺服器的連線則可以使用 `Close` 來正常關閉。

```go
func main() {
	c, _, _ := maxim.NewClient(&maxim.ClientConfig{
		Address: "ws://localhost:8080/ws",
	})
	c.Close()
}
```
