package maxim

import (
	"log"
	"net/http"
	"testing"
)

func TestPing(t *testing.T) {
	m := New(WithMessageHandler(xx))

	http.HandleFunc("/ws/", m.Handler)

	log.Fatal(http.ListenAndServe(":8080", nil))

}
