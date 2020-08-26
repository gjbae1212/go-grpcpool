package grpcpool

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestWithIdleTimeout(t *testing.T) {
	assert := assert.New(t)

	tests := map[string]struct {
		input  time.Duration
		output time.Duration
	}{
		"success": {input: time.Duration(1 * time.Second), output: time.Duration(1 * time.Second)},
	}
	for _, t := range tests {
		opts := &poolOptions{}
		f := WithIdleTimeout(t.input)
		f(opts)
		assert.Equal(t.output, opts.idleTimeout)
	}
}

func TestWithExpireTimeout(t *testing.T) {
	assert := assert.New(t)

	tests := map[string]struct {
		input  time.Duration
		output time.Duration
	}{
		"success": {input: time.Duration(2 * time.Second), output: time.Duration(2 * time.Second)},
	}
	for _, t := range tests {
		opts := &poolOptions{}
		f := WithExpireTimeout(t.input)
		f(opts)
		assert.Equal(t.output, opts.expireTimeout)
	}
}

func TestWithMaxRequestCount(t *testing.T) {
	assert := assert.New(t)

	tests := map[string]struct {
		input  uint32
		output uint32
	}{
		"success": {input: 10, output: 10},
	}
	for _, t := range tests {
		opts := &poolOptions{}
		f := WithMaxRequestCount(t.input)
		f(opts)
		assert.Equal(t.output, opts.maxRequestCount)
	}
}

func TestWithPoolSize(t *testing.T) {
	assert := assert.New(t)

	tests := map[string]struct {
		input  uint32
		output uint32
	}{
		"success": {input: 10, output: 10},
	}
	for _, t := range tests {
		opts := &poolOptions{}
		f := WithPoolSize(t.input)
		f(opts)
		assert.Equal(t.output, opts.poolSize)
	}
}
