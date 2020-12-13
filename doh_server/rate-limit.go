package doh

import (
	"golang.org/x/time/rate"
)

type RateLimiter interface {
	Please(ip string, userKey string) bool
}

type NoopRateLimiter struct{}

func (n *NoopRateLimiter) Please(a string, b string) bool {
	return true
}

type IPRateLimitConfig struct {
	userKeyWhitelist     set
	RecoverXTokensPerSec rate.Limit
	MaxTokens            int
}

type IPRateLimiter struct {
	userKeyWhitelist     set
	ipLimits             map[string]*rate.Limiter
	recoverXTokensPerSec rate.Limit
	maxTokens            int
}

func (rl *IPRateLimiter) Please(ip string, userKey string) bool {
	if rl.userKeyWhitelist.Contains(userKey) {
		return true
	}

	limiter, exists := rl.ipLimits[ip]

	if !exists {
		// don't care about the RC here
		rl.ipLimits[ip] = rate.NewLimiter(rl.recoverXTokensPerSec, rl.maxTokens)
		return true
	}

	return limiter.Allow()
}

func NewRateLimiter(config *IPRateLimitConfig) RateLimiter {

	if config == nil {
		return &NoopRateLimiter{}
	}

	return &IPRateLimiter{
		userKeyWhitelist:     config.userKeyWhitelist,
		ipLimits:             make(map[string]*rate.Limiter),
		recoverXTokensPerSec: config.RecoverXTokensPerSec,
		maxTokens:            config.MaxTokens,
	}
}
