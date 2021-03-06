package grpcpool

import (
	"time"
)

type Option interface {
	apply(*poolOptions)
}

type OptionFunc func(opts *poolOptions)

func (o OptionFunc) apply(opts *poolOptions) {
	o(opts)
}

// WithIdleTimeout returns a function which sets idle duration.
func WithIdleTimeout(d time.Duration) OptionFunc {
	return func(opts *poolOptions) {
		opts.idleTimeout = d
	}
}

// WithExpireTimeout returns a function which sets max life time.
func WithExpireTimeout(d time.Duration) OptionFunc {
	return func(opts *poolOptions) {
		opts.expireTimeout = d
	}
}

// WithMaxRequest returns a function which sets max request count.
func WithMaxRequestCount(c uint32) OptionFunc {
	return func(opts *poolOptions) {
		opts.maxRequestCount = c
	}
}

// WithPoolSize returns a function which sets pool size.
func WithPoolSize(size uint32) OptionFunc {
	return func(opts *poolOptions) {
		opts.poolSize = size
	}
}
