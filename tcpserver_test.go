package tcpserver_test

import (
	"bufio"
	"bytes"
	"net"
	"testing"
	"time"

	"github.com/overdrive3000/tcpserver"
)

func TestHandleFunc(t *testing.T) {
	t.Parallel()
	addr := ":8000"
	f1 := func(resp *tcpserver.Response, req *tcpserver.Request) { resp.Write([]byte("f1\n")) }
	tcpserver.HandleFunc("F1:", f1)

	go func() {
		tcpserver.ListenAndServe(addr)
	}()
	time.Sleep(time.Second)
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		t.Error(err)
	}
	defer conn.Close()
	want := []byte("f1\n")
	if _, err := conn.Write([]byte("F1: TEST\n")); err != nil {
		t.Error("could no write to server: ", err)
	}
	if got, err := bufio.NewReader(conn).ReadBytes('\n'); err == nil {
		if !bytes.Equal(want, got) {
			t.Errorf("want %q, got %q", string(want), string(got))
		}
	} else {
		t.Error("could not read from connection")
	}
}
