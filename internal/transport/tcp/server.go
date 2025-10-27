package tcp

import (
	"bufio"
	"context"
	"errors"
	"log/slog"
	"net"
	"sync"
	"time"
)

const (
	MaxConnectionBuffSize = 17 << 20
	readChunkSize         = 64 << 10 // 64 kb
	readerBufSize         = 32 << 10
)

type StdServer struct {
	addr           string
	handler        *Handler
	ln             net.Listener
	wg             sync.WaitGroup
	mu             sync.Mutex
	stop           bool
	AcceptDeadline time.Duration
	ReadTimeout    time.Duration
	WriteTimeout   time.Duration
}

func NewServer(addr string, h *Handler) *StdServer {
	return &StdServer{addr: addr, handler: h}
}

var tmpPool = sync.Pool{
	New: func() any { b := make([]byte, readChunkSize); return &b },
}
var bufPool = sync.Pool{
	New: func() any { b := make([]byte, 0, readerBufSize); return &b },
}

func (s *StdServer) Start() error {
	ln, err := net.Listen("tcp", s.addr)
	if err != nil {
		return err
	}
	s.mu.Lock()
	s.ln = ln
	s.mu.Unlock()

	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		tl, _ := ln.(*net.TCPListener)
		for {
			if s.AcceptDeadline > 0 && tl != nil {
				_ = tl.SetDeadline(time.Now().Add(s.AcceptDeadline))
			}
			conn, err := ln.Accept()
			if err != nil {
				var ne net.Error
				if errors.As(err, &ne) && ne.Timeout() {
					continue
				}
				s.mu.Lock()
				stopped := s.stop
				s.mu.Unlock()
				if stopped {
					return
				}
				slog.Error("stdtcp: accept error", "error", err)
				return
			}
			s.wg.Add(1)
			go func(c net.Conn) {
				defer s.wg.Done()
				s.handleConn(c)
			}(conn)
		}
	}()

	slog.Info("stdtcp: listening", "addr", s.addr)
	return nil
}

func (s *StdServer) Stop(ctx context.Context) error {
	s.mu.Lock()
	if s.stop {
		s.mu.Unlock()
		return nil
	}
	s.stop = true
	ln := s.ln
	s.mu.Unlock()
	if ln != nil {
		_ = ln.Close()
	}
	done := make(chan struct{})
	go func() { s.wg.Wait(); close(done) }()
	select {
	case <-done:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (s *StdServer) handleConn(c net.Conn) {
	defer c.Close()

	if tc, ok := c.(*net.TCPConn); ok {
		_ = tc.SetKeepAlive(true)
		_ = tc.SetKeepAlivePeriod(2 * time.Minute)
	}

	br := bufio.NewReaderSize(c, readerBufSize)
	bufPtr := bufPool.Get().(*[]byte)
	tmpPtr := tmpPool.Get().(*[]byte)
	buf := *bufPtr
	tmp := *tmpPtr
	defer func() {
		*bufPtr = buf[:0]
		bufPool.Put(bufPtr)
		tmpPool.Put(tmpPtr)
	}()

	for {
		if s.ReadTimeout > 0 {
			_ = c.SetReadDeadline(time.Now().Add(s.ReadTimeout))
		}
		n, err := br.Read(tmp)
		if err != nil {
			var ne net.Error
			if errors.As(err, &ne) && ne.Timeout() {
				return
			}
			return
		}
		if len(buf)+n > MaxConnectionBuffSize {
			slog.Warn("stdtcp: buffer overflow", "remote", c.RemoteAddr().String())
			return
		}
		buf = append(buf, tmp[:n]...)

		for {
			if len(buf) < HeaderSize {
				break
			}
			frameLen := int(binaryBEUint32(buf[0:4]))
			if frameLen > HeaderSize+MaxPayloadSize {
				slog.Error("stdtcp: oversized frame", "remote", c.RemoteAddr().String())
				return
			}
			if len(buf) < frameLen {
				break
			}
			cmd := buf[4]
			reqID := binaryBEUint64(buf[5:13])
			payload := buf[HeaderSize:frameLen]

			f := &Frame{Length: uint32(frameLen), Command: cmd, RequestID: reqID, Payload: payload}

			resp := s.handler.HandleFrame(f)
			if s.WriteTimeout > 0 {
				_ = c.SetWriteDeadline(time.Now().Add(s.WriteTimeout))
			}
			if _, err := c.Write(resp.Encode()); err != nil {
				return
			}

			buf = buf[frameLen:]
			if len(buf) == 0 {
				break
			}
		}
	}
}

func binaryBEUint32(b []byte) uint32 {
	_ = b[3]
	return uint32(b[0])<<24 | uint32(b[1])<<16 | uint32(b[2])<<8 | uint32(b[3])
}
func binaryBEUint64(b []byte) uint64 {
	_ = b[7]
	return uint64(b[0])<<56 | uint64(b[1])<<48 | uint64(b[2])<<40 | uint64(b[3])<<32 |
		uint64(b[4])<<24 | uint64(b[5])<<16 | uint64(b[6])<<8 | uint64(b[7])
}
