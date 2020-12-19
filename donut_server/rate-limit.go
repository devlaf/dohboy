package donut

import (
	"net/http"
	"strings"

	"golang.org/x/time/rate"
)

type RateLimiter interface {
	Please(ip string, userKey string) bool
	GetIP(request *http.Request) string
}

type NoopRateLimiter struct{}

func (n *NoopRateLimiter) Please(a string, b string) bool {
	return true
}

func (n *NoopRateLimiter) GetIP(request *http.Request) string {
	return ""
}

type IPRateLimiter struct {
	userKeyWhitelist     set
	ipLimits             map[string]*rate.Limiter
	recoverXTokensPerSec rate.Limit
	maxTokens            int
	allowIPFromHeader    bool
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

func (rl *IPRateLimiter) GetIP(request *http.Request) string {
	if rl.allowIPFromHeader {
		if forwarded := request.Header.Get("X-FORWARDED-FOR"); forwarded != "" {
			return forwarded
		}
		if real := request.Header.Get("X-Real-IP"); real != "" {
			return real
		}
	}
	return request.RemoteAddr
}

func toSet(commaSeparated string) set {
	retval := NewSet()

	for _, key := range strings.Split(commaSeparated, ",") {
		retval.Add(strings.TrimSpace(key))
	}

	return *retval
}

func NewRateLimiter(config *Config) RateLimiter {

	if config.IPRateLimit.Enabled == false {
		return &NoopRateLimiter{}
	}

	return &IPRateLimiter{
		userKeyWhitelist:     toSet(config.IPRateLimit.KeyWhitelist),
		ipLimits:             make(map[string]*rate.Limiter),
		recoverXTokensPerSec: rate.Limit(config.IPRateLimit.RecoverXTokensPerSec),
		maxTokens:            config.IPRateLimit.MaxTokens,
		allowIPFromHeader:    config.IPRateLimit.FetchIPFromHeaders,
	}
}
