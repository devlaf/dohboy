package doh

import (
	"strings"

	"golang.org/x/time/rate"
)

type RateLimiter interface {
	Please(ip string, userKey string) bool
}

type NoopRateLimiter struct{}

func (n *NoopRateLimiter) Please(a string, b string) bool {
	return true
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

func toSet(commaSeparated string) set {
	retval := NewSet()

	for _, key := range strings.Split(commaSeparated, ",") {
		retval.Add(strings.TrimSpace(key))
	}

	return *retval
}

func NewRateLimiter(config Config) RateLimiter {

	if config.Server.IPRateLimit.Enabled == false {
		return &NoopRateLimiter{}
	}

	return &IPRateLimiter{
		userKeyWhitelist:     toSet(config.Server.IPRateLimit.KeyWhitelist),
		ipLimits:             make(map[string]*rate.Limiter),
		recoverXTokensPerSec: rate.Limit(config.Server.IPRateLimit.RecoverXTokensPerSec),
		maxTokens:            config.Server.IPRateLimit.MaxTokens,
	}
}
