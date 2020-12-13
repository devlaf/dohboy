package doh

import (
	"golang.org/x/time/rate"
)

type RateLimiter interface {
	Please(val string) bool
}

type NoopRateLimiter struct{}

func (n *NoopRateLimiter) Please(whatever string) bool {
	return true
}

type IPRateLimitConfig struct {
	WhitelistIPs set
	Rate         rate.Limit
	BucketSize   int
}

type IPRateLimiter struct {
	ipWhitelist set
	ipLimits    map[string]*rate.Limiter
	rate        rate.Limit
	bucketSize  int
}

func (rl *IPRateLimiter) Please(ip string) bool {
	if rl.ipWhitelist.Contains(ip) {
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
		ipWhitelist: config.WhitelistIPs,
		ipLimits:    make(map[string]*rate.Limiter),
		rate:        config.Rate,
		bucketSize:  config.BucketSize,
	}
}
