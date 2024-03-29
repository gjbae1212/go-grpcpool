package grpcpool

import (
	"context"
	"fmt"
	"net"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
	grpchealth "google.golang.org/grpc/health"
	health_pb "google.golang.org/grpc/health/grpc_health_v1"
)

func TestNewPool(t *testing.T) {
	assert := assert.New(t)

	tests := map[string]struct {
		inputOpts  []Option
		inputF     func() (*grpc.ClientConn, error)
		outputOpts *poolOptions
		isErr      bool
	}{
		"fail": {isErr: true},
		"default": {
			inputF:     func() (*grpc.ClientConn, error) { return nil, nil },
			outputOpts: &poolOptions{poolSize: 10, idleTimeout: 5 * time.Minute, expireTimeout: time.Hour, maxRequestCount: 1 << 17},
		},
		"pass": {
			inputF:     func() (*grpc.ClientConn, error) { return nil, nil },
			inputOpts:  []Option{WithPoolSize(1), WithIdleTimeout(time.Second), WithExpireTimeout(time.Second), WithMaxRequestCount(1)},
			outputOpts: &poolOptions{poolSize: 1, idleTimeout: time.Second, expireTimeout: time.Second, maxRequestCount: 1},
		},
		"lazy-loading": {
			inputF:     func() (*grpc.ClientConn, error) { return nil, nil },
			inputOpts:  []Option{WithLazyLoading(true)},
			outputOpts: &poolOptions{poolSize: 10, idleTimeout: 5 * time.Minute, expireTimeout: time.Hour, maxRequestCount: 1 << 17, lazyLoading: true},
		},
	}

	for _, t := range tests {
		p, err := NewPool(t.inputF, t.inputOpts...)
		assert.Equal(t.isErr, err != nil)
		if err == nil {
			assert.Equal(p.(*pool).opts.maxRequestCount, t.outputOpts.maxRequestCount)
			assert.Equal(p.(*pool).opts.idleTimeout, t.outputOpts.idleTimeout)
			assert.Equal(p.(*pool).opts.expireTimeout, t.outputOpts.expireTimeout)
			assert.Equal(p.(*pool).opts.poolSize, t.outputOpts.poolSize)
			assert.Equal(p.(*pool).opts.lazyLoading, t.outputOpts.lazyLoading)
			if p.(*pool).opts.lazyLoading {
				assert.Len(p.(*pool).conns, 0)
			} else {
				assert.Len(p.(*pool).conns, int(p.(*pool).opts.poolSize))
			}
		}
	}

}

func TestPool_Size(t *testing.T) {
	assert := assert.New(t)

	p1, err := NewPool(func() (*grpc.ClientConn, error) {
		return nil, nil
	})
	assert.NoError(err)

	p2, err := NewPool(func() (*grpc.ClientConn, error) {
		return nil, nil
	}, WithLazyLoading(true))
	assert.NoError(err)

	tests := map[string]struct {
		input  Pool
		output uint32
	}{
		"default":      {input: p1, output: 10},
		"lazy-loading": {input: p2, output: 0},
	}

	for _, t := range tests {
		result := t.input.Size()
		assert.Equal(t.output, result)
	}

}

func TestPool_Close(t *testing.T) {
	assert := assert.New(t)

	p, err := NewPool(func() (*grpc.ClientConn, error) {
		return nil, nil
	})
	assert.NoError(err)

	tests := map[string]struct {
		isErr bool
	}{
		"success": {isErr: false},
	}

	for _, t := range tests {
		err := p.Close()
		assert.Equal(t.isErr, err != nil)
		assert.Equal(len(p.(*pool).conns), 0)
	}

}

func TestPool_GetConn(t *testing.T) {
	assert := assert.New(t)

	server := grpc.NewServer()
	health_pb.RegisterHealthServer(server, grpchealth.NewServer())
	// create tcp
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", 9999))
	if err != nil {
		panic(err)
	}
	defer lis.Close()
	go server.Serve(lis)

	time.Sleep(1 * time.Second)

	p1, err := NewPool(func() (*grpc.ClientConn, error) {
		return grpc.Dial(fmt.Sprintf("localhost:%s", "9999"), grpc.WithInsecure(), grpc.WithBlock())
	}, WithMaxRequestCount(3000), WithExpireTimeout(100*time.Second), WithIdleTimeout(100*time.Second))
	assert.NoError(err)

	p2, err := NewPool(func() (*grpc.ClientConn, error) {
		return grpc.Dial(fmt.Sprintf("localhost:%s", "9999"), grpc.WithInsecure(), grpc.WithBlock())
	}, WithMaxRequestCount(3000), WithExpireTimeout(100*time.Second), WithIdleTimeout(100*time.Second), WithLazyLoading(true))
	assert.NoError(err)

	w := sync.WaitGroup{}
	for i := 0; i < 10; i++ {
		w.Add(1)
		go func() {
			defer w.Done()
			for j := 0; j < 10000; j++ {
				conn, err := p1.GetConn()
				assert.NoError(err)
				client := health_pb.NewHealthClient(conn)
				_, err = client.Check(context.Background(), &health_pb.HealthCheckRequest{Service: ""})
				assert.NoError(err)

				conn, err = p2.GetConn()
				assert.NoError(err)
				client = health_pb.NewHealthClient(conn)
				_, err = client.Check(context.Background(), &health_pb.HealthCheckRequest{Service: ""})
				assert.NoError(err)
			}
		}()
	}
	w.Wait()

	fmt.Println("p1 pool ", len(p1.(*pool).conns))
	fmt.Println("p2 pool ", len(p2.(*pool).conns))
}

func BenchmarkPool_GetConn(b *testing.B) {
	server := grpc.NewServer()
	health_pb.RegisterHealthServer(server, grpchealth.NewServer())
	// create tcp
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", 9999))
	if err != nil {
		panic(err)
	}
	defer lis.Close()
	go server.Serve(lis)

	time.Sleep(1 * time.Second)

	p, err := NewPool(func() (*grpc.ClientConn, error) {
		return grpc.Dial(fmt.Sprintf("localhost:%s", "9999"), grpc.WithInsecure(), grpc.WithBlock())
	}, WithMaxRequestCount(1000000), WithExpireTimeout(20*time.Second), WithIdleTimeout(20*time.Second), WithLazyLoading(true))

	for i := 0; i < b.N; i++ {
		conn, _ := p.GetConn()
		client := health_pb.NewHealthClient(conn)
		_, err = client.Check(context.Background(), &health_pb.HealthCheckRequest{Service: ""})
	}
}
