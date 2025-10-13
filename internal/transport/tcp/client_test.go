// Note: This client implemented for testing purposes only.

package tcp

import (
	"bufio"
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"net"
	"sync"
	"sync/atomic"
	"time"

	"key-value-store/internal/util"
)

var (
	ErrTimeout        = errors.New("request timeout")
	ErrClosed         = errors.New("client closed")
	ErrResponseStatus = errors.New("non-ok response status")
)

type Options struct {
	ReadBufferSize  int
	WriteBufferSize int
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
}

type Client struct {
	conn       net.Conn
	opts       Options
	requestID  atomic.Uint64
	writer     *bufio.Writer
	writeCh    chan []byte
	pending    sync.Map
	closeOnce  sync.Once
	closeCh    chan struct{}
	closed     atomic.Bool
	ctx        context.Context
	cancelFunc context.CancelFunc
}

func New(ctx context.Context, addr string, opts Options) (*Client, error) {
	if opts.ReadBufferSize <= 0 {
		opts.ReadBufferSize = 256 * 1024
	}
	if opts.WriteBufferSize <= 0 {
		opts.WriteBufferSize = 256 * 1024
	}
	if opts.ReadTimeout <= 0 {
		opts.ReadTimeout = 10 * time.Second
	}
	if opts.WriteTimeout <= 0 {
		opts.WriteTimeout = 10 * time.Second
	}

	conn, err := net.DialTimeout("tcp", addr, 5*time.Second)
	if err != nil {
		return nil, fmt.Errorf("failed to connect: %w", err)
	}

	if tcpConn, ok := conn.(*net.TCPConn); ok {
		tcpConn.SetNoDelay(true)
		tcpConn.SetReadBuffer(opts.ReadBufferSize)
		tcpConn.SetWriteBuffer(opts.WriteBufferSize)
		tcpConn.SetKeepAlive(true)
		tcpConn.SetKeepAlivePeriod(30 * time.Second)
	}

	clientCtx, cancel := context.WithCancel(ctx)

	c := &Client{
		conn:       conn,
		opts:       opts,
		writer:     bufio.NewWriterSize(conn, opts.WriteBufferSize),
		writeCh:    make(chan []byte, 1024), // Buffered channel for lock-free writes
		closeCh:    make(chan struct{}),
		ctx:        clientCtx,
		cancelFunc: cancel,
	}

	go c.readLoop()

	go c.writeLoop()

	return c, nil
}

func (c *Client) nextRequestID() uint64 {
	return c.requestID.Add(1)
}

func (c *Client) sendFrame(frame *Frame) (<-chan *Frame, error) {
	if c.closed.Load() {
		return nil, ErrClosed
	}

	respCh := make(chan *Frame, 1)
	c.pending.Store(frame.RequestID, respCh)

	data := frame.Encode()

	select {
	case c.writeCh <- data:
		return respCh, nil
	case <-c.closeCh:
		c.pending.Delete(frame.RequestID)
		close(respCh)
		return nil, ErrClosed
	case <-time.After(c.opts.WriteTimeout):
		c.pending.Delete(frame.RequestID)
		close(respCh)
		return nil, ErrTimeout
	}
}

func (c *Client) writeLoop() {
	const (
		maxBatch = 50
		maxDelay = 50 * time.Microsecond
	)

	ticker := time.NewTicker(maxDelay)
	defer ticker.Stop()

	pendingWrites := 0

	for {
		select {
		case <-c.closeCh:
			return

		case data := <-c.writeCh:
			c.conn.SetWriteDeadline(time.Now().Add(c.opts.WriteTimeout))

			if _, err := c.writer.Write(data); err != nil {
				c.Close()
				return
			}

			pendingWrites++

			for i := 0; i < 10 && pendingWrites < maxBatch; i++ {
				select {
				case moreData := <-c.writeCh:
					c.writer.Write(moreData)
					pendingWrites++
				default:
					goto flush
				}
			}

		flush:
			if err := c.writer.Flush(); err != nil {
				c.Close()
				return
			}
			pendingWrites = 0

		case <-ticker.C:

			if pendingWrites > 0 {
				if err := c.writer.Flush(); err != nil {
					c.Close()
					return
				}
				pendingWrites = 0
			}
		}
	}
}

func (c *Client) readLoop() {
	defer func() {
		c.Close()
		c.pending.Range(func(key, value interface{}) bool {
			if ch, ok := value.(chan *Frame); ok {
				close(ch)
			}
			c.pending.Delete(key)
			return true
		})
	}()

	reader := bufio.NewReaderSize(c.conn, c.opts.ReadBufferSize)
	headerBuf := make([]byte, HeaderSize)

	for {
		if c.closed.Load() {
			return
		}

		if err := c.conn.SetReadDeadline(time.Now().Add(c.opts.ReadTimeout)); err != nil {
			return
		}

		if _, err := io.ReadFull(reader, headerBuf); err != nil {
			if !c.closed.Load() {
				if err != io.EOF {
				}
			}
			return
		}
		frameLen := binary.BigEndian.Uint32(headerBuf[0:4])
		if frameLen < HeaderSize || frameLen > MaxPayloadSize+HeaderSize {
			// Invalid frame
			return
		}

		fullFrame := make([]byte, frameLen)
		copy(fullFrame, headerBuf)

		if frameLen > HeaderSize {
			if _, err := io.ReadFull(reader, fullFrame[HeaderSize:]); err != nil {
				return
			}
		}

		frame, err := DecodeFrame(fullFrame)
		if err != nil {
			return
		}

		if val, ok := c.pending.LoadAndDelete(frame.RequestID); ok {
			if ch, ok := val.(chan *Frame); ok {
				select {
				case ch <- frame:
				default:
				}
				close(ch)
			}
		}
	}
}

func (c *Client) waitResponse(ctx context.Context, respCh <-chan *Frame) (*Frame, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-c.closeCh:
		return nil, ErrClosed
	case frame, ok := <-respCh:
		if !ok {
			return nil, ErrClosed
		}
		return frame, nil
	}
}

func (c *Client) Set(ctx context.Context, token, bucket, key string, ttl int64, singleRead bool, value []byte) (bool, error) {
	reqID := c.nextRequestID()
	payload := EncodeSetPayload(token, bucket, key, ttl, singleRead, value)
	frame := NewFrame(CmdSet, reqID, payload)

	respCh, err := c.sendFrame(frame)
	if err != nil {
		return false, err
	}

	resp, err := c.waitResponse(ctx, respCh)
	if err != nil {
		return false, err
	}

	if resp.Command == CmdError {
		status, msg, _ := ParseResponsePayload(resp.Payload)
		return false, fmt.Errorf("error %d: %s", status, util.BytesToString(msg))
	}
	status, _, err := ParseResponsePayload(resp.Payload)
	if err != nil {
		return false, err
	}

	return status == StatusCreated, nil
}

func (c *Client) Get(ctx context.Context, token, bucket, key string) ([]byte, error) {
	reqID := c.nextRequestID()
	payload := EncodeGetPayload(token, bucket, key)
	frame := NewFrame(CmdGet, reqID, payload)

	respCh, err := c.sendFrame(frame)
	if err != nil {
		return nil, err
	}

	resp, err := c.waitResponse(ctx, respCh)
	if err != nil {
		return nil, err
	}
	if resp.Command == CmdError {
		status, msg, _ := ParseResponsePayload(resp.Payload)
		return nil, fmt.Errorf("error %d: %s", status, util.BytesToString(msg))
	}

	status, data, err := ParseResponsePayload(resp.Payload)
	if err != nil {
		return nil, err
	}

	if status != StatusOK {
		return nil, fmt.Errorf("unexpected status: %d", status)
	}

	offset := 0

	if len(data) < offset+2 {
		return nil, errors.New("invalid response: missing key length")
	}
	keyLen := int(binary.BigEndian.Uint16(data[offset:]))
	offset += 2 + keyLen

	offset += 8 + 8 + 8 + 1

	if len(data) < offset+4 {
		return nil, errors.New("invalid response: missing value length")
	}
	valueLen := int(binary.BigEndian.Uint32(data[offset:]))
	offset += 4

	if len(data) < offset+valueLen {
		return nil, errors.New("invalid response: incomplete value")
	}
	value := make([]byte, valueLen)
	copy(value, data[offset:offset+valueLen])

	return value, nil
}

func (c *Client) Delete(ctx context.Context, token, bucket, key string) error {
	reqID := c.nextRequestID()
	payload := EncodeDeletePayload(token, bucket, key)
	frame := NewFrame(CmdDelete, reqID, payload)

	respCh, err := c.sendFrame(frame)
	if err != nil {
		return err
	}

	resp, err := c.waitResponse(ctx, respCh)
	if err != nil {
		return err
	}

	if resp.Command == CmdError {
		status, msg, _ := ParseResponsePayload(resp.Payload)
		return fmt.Errorf("error %d: %s", status, util.BytesToString(msg))
	}

	return nil
}

func (c *Client) Close() error {
	var err error
	c.closeOnce.Do(func() {
		c.closed.Store(true)
		c.cancelFunc()
		close(c.closeCh)
		err = c.conn.Close()
	})
	return err
}
