package maxim

// Bucket 呈現了一個可以填裝連線階段的水桶。
type Bucket struct {
	// sessions 是位於此水桶內的所有階段客戶端連線。
	sessions []*Session
	// config 是水桶設置。
	config *BucketConfig
}

// BucketConfig 是水桶設置。
type BucketConfig struct {
}

// NewBucket 會建立一個新的階段水桶。
func NewBucket(conf *BucketConfig) *Bucket {
	return &Bucket{
		config: conf,
	}
}

// Put 能夠放入指定的客戶端連線。
func (b *Bucket) Put(s *Session) error {
	for _, v := range b.sessions {
		if v == s {
			return ErrDuplicatedSession
		}
	}
	b.sessions = append(b.sessions, s)
	return nil
}

// Delete 會從水桶中移除指定的客戶端連線。
func (b *Bucket) Delete(s *Session) error {
	for k, v := range b.sessions {
		if v == s {
			b.sessions = append(b.sessions[:k], b.sessions[k+1:]...)
			return nil
		}
	}
	return ErrSessionNotFound
}

// Write 能夠將文字訊息寫入到水桶中的所有客戶端。
func (b *Bucket) Write(msg string) {
	for _, v := range b.sessions {
		v.Write(msg)
	}
}

// WriteFilter 能夠將文字訊息寫入到水桶中被篩選的客戶端。
func (b *Bucket) WriteFilter(msg string, fn func(*Session) bool) {
	for _, v := range b.sessions {
		if fn(v) {
			v.Write(msg)
		}
	}
}

// WriteOthers 能夠將文字訊息寫入到水桶中指定以外的所有客戶端。
func (b *Bucket) WriteOthers(msg string, s *Session) {
	for _, v := range b.sessions {
		if v != s {
			v.Write(msg)
		}
	}
}

// WriteBinary 能夠將二進制訊息寫入到水桶中的所有客戶端。
func (b *Bucket) WriteBinary(msg []byte) {
	for _, v := range b.sessions {
		v.WriteBinary(msg)
	}
}

// WriteBinaryFilter 能夠將二進制訊息寫入到水桶中被篩選客戶端。
func (b *Bucket) WriteBinaryFilter(msg []byte, fn func(*Session) bool) {
	for _, v := range b.sessions {
		if fn(v) {
			v.WriteBinary(msg)
		}
	}
}

// WriteBinaryOthers 能夠將二進制訊息寫入到水桶中指定以外的所有客戶端。
func (b *Bucket) WriteBinaryOthers(msg []byte, s *Session) {
	for _, v := range b.sessions {
		if v != s {
			v.WriteBinary(msg)
		}
	}
}

// Contains 會表示指定的客戶端是否有在此水桶內。
func (b *Bucket) Contains(s *Session) bool {
	for _, v := range b.sessions {
		if v == s {
			return true
		}
	}
	return false
}

// Close 會關閉此水桶的所有客戶端連線。
func (b *Bucket) Close(c CloseStatus) {
	for _, v := range b.sessions {
		v.Close(CloseNormalClosure)
	}
}

// Len 會表示頻道的總訂閱客戶端數量。
func (b *Bucket) Len() int {
	return len(b.sessions)
}
