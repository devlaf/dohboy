package doh

import (
	"golang.org/x/time/rate"
)

type RateLimiter interface {
	Please(ip string, token string) bool
}

type NoopRateLimiter struct{}

func (n *NoopRateLimiter) Please(a string, b string) bool {
	return true
}

type IPRateLimitConfig struct {
	tokenWhitelist set
	Rate           rate.Limit
	BucketSize     int
}

type IPRateLimiter struct {
	tokenWhitelist set
	ipLimits       map[string]*rate.Limiter
	rate           rate.Limit
	bucketSize     int
}

func (rl *IPRateLimiter) Please(ip string, token string) bool {
	if rl.tokenWhitelist.Contains(token) {
		return true
	}

	limiter, exists := rl.ipLimits[ip]

	if !exists {
		// don't care about the RC here
		rl.ipLimits[ip] = rate.NewLimiter(rl.rate, rl.bucketSize)
		return true
	}

	return limiter.Allow()
}

func NewRateLimiter(config *IPRateLimitConfig) RateLimiter {

	if config == nil {
		return &NoopRateLimiter{}
	}

	return &IPRateLimiter{
		tokenWhitelist: config.tokenWhitelist,
		ipLimits:       make(map[string]*rate.Limiter),
		rate:           config.Rate,
		bucketSize:     config.BucketSize,
	}
}
