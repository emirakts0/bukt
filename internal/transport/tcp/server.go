package tcp

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/panjf2000/gnet/v2"
)

const (
	MaxConnectionBufferSize = 17 * 1024 * 1024
	InitialBufferCapacity   = 4096
)

var bufferPool = sync.Pool{
	New: func() interface{} {
		buf := make([]byte, 0, InitialBufferCapacity)
		return &buf
	},
}

type Server struct {
	gnet.BuiltinEventEngine

	eng        gnet.Engine
	addr       string
	handler    *Handler
	isRunning  bool
	runningMu  sync.RWMutex
	workerPool chan struct{}
}

type connState struct {
	buffer []byte
}

func NewServer(addr string, handler *Handler) *Server {
	poolSize := 256

	return &Server{
		addr:       addr,
		handler:    handler,
		workerPool: make(chan struct{}, poolSize),
	}
}

func (s *Server) OnBoot(eng gnet.Engine) gnet.Action {
	s.eng = eng
	s.runningMu.Lock()
	s.isRunning = true
	s.runningMu.Unlock()
	slog.Info("TCP Server: Booted", "address", s.addr)
	return gnet.None
}

func (s *Server) OnOpen(c gnet.Conn) ([]byte, gnet.Action) {
	bufPtr := bufferPool.Get().(*[]byte)
	*bufPtr = (*bufPtr)[:0]

	c.SetContext(&connState{
		buffer: *bufPtr,
	})

	return nil, gnet.None
}

func (s *Server) OnClose(c gnet.Conn, err error) gnet.Action {
	if ctx := c.Context(); ctx != nil {
		if state, ok := ctx.(*connState); ok {
			bufferPool.Put(&state.buffer)
		}
	}

	return gnet.None
}

func (s *Server) OnTraffic(c gnet.Conn) gnet.Action {
	ctx := c.Context()
	if ctx == nil {
		slog.Error("TCP Server: Connection context not found")
		return gnet.Close
	}

	state, ok := ctx.(*connState)
	if !ok {
		slog.Error("TCP Server: Invalid connection context type")
		return gnet.Close
	}

	buf, err := c.Next(-1)
	if err != nil {
		slog.Error("TCP Server: Failed to read data", "error", err)
		return gnet.Close
	}

	if len(state.buffer)+len(buf) > MaxConnectionBufferSize {
		slog.Warn("TCP Server: Connection buffer overflow",
			"remote", c.RemoteAddr().String(),
			"current_size", len(state.buffer),
			"incoming_size", len(buf),
			"max_size", MaxConnectionBufferSize)
		return gnet.Close
	}

	state.buffer = append(state.buffer, buf...)

	for {
		if len(state.buffer) < HeaderSize {
			break
		}

		frame, err := DecodeFrame(state.buffer)
		if err != nil {
			if err == ErrIncompleteFrame {
				break
			}
			slog.Error("TCP Server: Invalid frame", "error", err, "remote", c.RemoteAddr().String())
			return gnet.Close
		}

		select {
		case s.workerPool <- struct{}{}:
			frameCopy := &Frame{
				Length:    frame.Length,
				Command:   frame.Command,
				RequestID: frame.RequestID,
				Payload:   make([]byte, len(frame.Payload)),
			}
			copy(frameCopy.Payload, frame.Payload)

			go func() {
				defer func() {
					<-s.workerPool
				}()

				responseFrame := s.handler.HandleFrame(frameCopy)

				responseData := responseFrame.Encode()
				if err := c.AsyncWrite(responseData, nil); err != nil {
					return
				}
			}()

		default:
			responseFrame := s.handler.HandleFrame(frame)
			responseData := responseFrame.Encode()
			if err := c.AsyncWrite(responseData, nil); err != nil {
				slog.Error("TCP Server: Failed to write response", "error", err)
				return gnet.Close
			}
		}

		state.buffer = state.buffer[frame.Length:]
	}

	return gnet.None
}

func (s *Server) OnShutdown(eng gnet.Engine) {
	slog.Info("TCP Server: Shutting down")
}

func (s *Server) OnTick() (delay time.Duration, action gnet.Action) {
	return 10 * time.Second, gnet.None
}

func (s *Server) Start() error {
	slog.Info("TCP Server: Starting", "address", s.addr)

	err := gnet.Run(
		s,
		s.addr,
		gnet.WithMulticore(true),
		gnet.WithReusePort(true),
		gnet.WithTCPKeepAlive(time.Minute),
		gnet.WithTCPNoDelay(gnet.TCPNoDelay),
		gnet.WithSocketRecvBuffer(256*1024),
		gnet.WithSocketSendBuffer(256*1024),
	)

	if err != nil {
		return fmt.Errorf("failed to start TCP server: %w", err)
	}

	return nil
}

func (s *Server) Stop(ctx context.Context) error {
	slog.Info("TCP Server: Stopping")

	s.runningMu.RLock()
	running := s.isRunning
	s.runningMu.RUnlock()

	if !running {
		slog.Debug("TCP Server: Already stopped or not started")
		return nil
	}

	return s.eng.Stop(ctx)
}
