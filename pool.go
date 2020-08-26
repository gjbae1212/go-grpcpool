package grpcpool

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"google.golang.org/grpc"
)

var (
	defaultOpts = []Option{
		WithPoolSize(10),
		WithIdleTimeout(5 * time.Minute),
		WithExpireTimeout(1 * time.Hour),
		WithMaxRequestCount(1 << 17),
	}
)

type connector func() (*grpc.ClientConn, error)

type Pool interface {
	GetConn() (*grpc.ClientConn, error)
	Size() uint32
	Close() error
}

type poolOptions struct {
	poolSize        uint32
	idleTimeout     time.Duration
	expireTimeout   time.Duration
	maxRequestCount uint32
}

type pool struct {
	sync.RWMutex
	idx       uint64
	conns     []*clientConn
	connector connector
	opts      *poolOptions
}

// Conn returns valid grpc connection.
func (p *pool) GetConn() (*grpc.ClientConn, error) {
	p.RLock()
	defer p.RUnlock()

	if len(p.conns) == 0 {
		return nil, fmt.Errorf("[err] Conn %w", ErrorPoolEmpty)
	}

	i := atomic.AddUint64(&p.idx, 1)
	conn := p.conns[i%uint64(len(p.conns))]

	return conn.getConn()
}

// Size returns pool size.
func (p *pool) Size() uint32 {
	return uint32(len(p.conns))
}

// Close closes connections.
func (p *pool) Close() error {
	p.Lock()
	defer p.Unlock()
	for _, conn := range p.conns {
		conn.pool = nil
		if conn.current != nil {
			conn.current.Close()
		}
	}
	p.conns = nil
	return nil
}

type clientConn struct {
	sync.RWMutex
	pool         *pool
	current      *grpc.ClientConn
	initial      time.Time
	latest       time.Time
	requestCount uint32
}

func (c *clientConn) getConn() (*grpc.ClientConn, error) {
	c.Lock()
	defer c.Unlock()

	now := time.Now()
	if c.current == nil || c.initial.Add(c.pool.opts.expireTimeout).Before(now) ||
		c.latest.Add(c.pool.opts.idleTimeout).Before(now) ||
		c.requestCount > c.pool.opts.maxRequestCount {

		//After 1.minute, close outdated connection.
		go func(conn *grpc.ClientConn) {
			select {
			case <-time.After(time.Minute):
				conn.Close()
			}
		}(c.current)

		// initialize connection and options.
		c.current = nil
		c.initial = now
		c.requestCount = 0

		// get new connection.
		conn, err := c.pool.connector()
		if err != nil {
			return nil, fmt.Errorf("[err] getConn create connection %w", err)
		}
		c.current = conn
	}

	c.requestCount += 1
	c.latest = now
	return c.current, nil
}

// NewPool returns GRPC pool interface.
func NewPool(f func() (*grpc.ClientConn, error), opts ...Option) (Pool, error) {
	if f == nil {
		return nil, fmt.Errorf("[err] NewPool %w", ErrorInvalidParams)
	}

	var mergedOpt []Option
	mergedOpt = append(mergedOpt, defaultOpts...)
	mergedOpt = append(mergedOpt, opts...)

	po := &poolOptions{}
	for _, opt := range mergedOpt {
		opt.apply(po)
	}

	if po.poolSize == 0 || po.maxRequestCount == 0 {
		return nil, fmt.Errorf("[err] NewPool %w", ErrorInvalidParams)
	}

	p := &pool{connector: f, opts: po}
	for i := 0; i < int(p.opts.poolSize); i++ {
		conn, err := p.connector()
		if err != nil {
			return nil, fmt.Errorf("[err] NewPool %w", err)
		}
		wraper := &clientConn{initial: time.Now(), latest: time.Now(), current: conn, pool: p}
		p.conns = append(p.conns, wraper)
	}

	return p, nil
}
