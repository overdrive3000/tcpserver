package tcpserver

import (
	"bufio"
	"errors"
	"net"
	"regexp"
	"strings"
	"sync"
)

type HandlerFunc func(resp *Response, req *Request)

func (f HandlerFunc) ServeTCP(resp *Response, req *Request) {
	f(resp, req)
}

type Server struct {
	// Addr specify TCP address and port to liste on
	Addr   string
	Handle handleEntry

	mu         sync.Mutex
	doneChan   chan struct{}
	onShutdown []func(resp *Response, req *Request)
}

type handleEntry struct {
	entries map[string]HandlerFunc // Map of TCP entry endpoints and handler function
}

type Response struct {
	conn net.Conn
	req  *Request
}

func (r *Response) Write(b []byte) {
	r.conn.Write(b)
}

type Request struct {
	payload []byte
	command string // Command must be in the format of "STRING:"
	data    string
}

var DefaultServe = &defaultServe
var defaultServe Server

func (srv *Server) ListenAndServe() error {
	addr := srv.Addr
	if addr == "" {
		addr = ":8000"
	}
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	return srv.Serve(ln)
}

func (srv *Server) Serve(l net.Listener) error {
	for {
		c, err := l.Accept()
		if err != nil {
			// check if here but not yet implemented
			select {
			case <-srv.getDoneChan():
				return errors.New("Server Closed")
			default:
				{
				}
			}
			return err
		}
		go srv.serve(c)
	}
}

func (srv *Server) serve(c net.Conn) {
	defer c.Close()
	for {
		buf, err := bufio.NewReader(c).ReadBytes('\n')
		if err != nil {
			return
		}
		req, err := srv.parseRequest(buf)
		if err != nil {
			return
		}
		resp := &Response{
			conn: c,
			req:  req,
		}
		f, ok := srv.Handle.entries[req.command]
		if !ok {
			panic("handler not found")
		}
		f.ServeTCP(resp, req)
	}

}

func (srv *Server) parseRequest(b []byte) (*Request, error) {
	raw := strings.TrimRight(string(b), "\n")
	arr := strings.Split(raw, ": ")
	if len(arr) != 2 {
		return &Request{}, errors.New("invalid request format")
	}
	return &Request{
		payload: b,
		command: arr[0] + ":",
		data:    arr[1],
	}, nil
}

func (srv *Server) getDoneChan() <-chan struct{} {
	srv.mu.Lock()
	defer srv.mu.Unlock()
	return srv.getDoneChanLocked()
}

func (srv *Server) getDoneChanLocked() chan struct{} {
	if srv.doneChan == nil {
		srv.doneChan = make(chan struct{})
	}
	return srv.doneChan
}

func (srv *Server) HandleFunc(pattern string, handler func(resp *Response, req *Request)) {
	if handler == nil {
		panic("No function handler")
	}
	srv.handle(pattern, HandlerFunc(handler))
}

// register new handler function for given pattern
func (srv *Server) handle(pattern string, handler HandlerFunc) {
	if pattern == "" {
		panic("empty patern")
	}
	m, _ := regexp.MatchString("([A-Z]+)([0-9]+)?:", pattern)
	if !m {
		panic("invalid pattern format, it have to match ([A-Z]+:) pattern")
	}
	if handler == nil {
		panic("nil handler")
	}
	if _, exist := srv.Handle.entries[pattern]; exist {
		panic("multiple registration for: " + pattern)
	}

	if srv.Handle.entries == nil {
		srv.Handle.entries = make(map[string]HandlerFunc)
	}
	srv.Handle.entries[pattern] = handler
}

func NewServer() *Server { return new(Server) }

func HandleFunc(pattern string, handler func(resp *Response, req *Request)) {
	DefaultServe.HandleFunc(pattern, handler)
}

func ListenAndServe(addr string) error {
	server := DefaultServe
	(server).Addr = addr
	return server.ListenAndServe()
}
