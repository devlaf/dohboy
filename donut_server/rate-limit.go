package donut

import (
	"net/http"
	"strings"
	"sync"

	"golang.org/x/time/rate"
)

type rateLimiter interface {
	please(ip string, userKey string) bool
	getIP(request *http.Request) string
}

type noopRateLimiter struct{}

func (n *noopRateLimiter) please(a string, b string) bool {
	return true
}

func (n *noopRateLimiter) getIP(request *http.Request) string {
	return ""
}

type iPRateLimiter struct {
	userKeyWhitelist     set
	ipLimits             map[string]*rate.Limiter
	ipLimitsMu           sync.RWMutex
	recoverXTokensPerSec rate.Limit
	maxTokens            int
	allowIPFromHeader    bool
}

func (rl *iPRateLimiter) please(ip string, userKey string) bool {
	if rl.userKeyWhitelist.Contains(userKey) {
		return true
	}

	rl.ipLimitsMu.RLock()
	limiter, exists := rl.ipLimits[ip]
	rl.ipLimitsMu.RUnlock()

	if !exists {
		limiter = rate.NewLimiter(rl.recoverXTokensPerSec, rl.maxTokens)
		rl.ipLimitsMu.Lock()
		rl.ipLimits[ip] = limiter
		rl.ipLimitsMu.Unlock()
	}

	return limiter.Allow()
}

func (rl *iPRateLimiter) getIP(request *http.Request) string {
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
	retval := newSet()

	for _, key := range strings.Split(commaSeparated, ",") {
		retval.Add(strings.TrimSpace(key))
	}

	return *retval
}

func newRateLimiter(config *Config) rateLimiter {

	if config.IPRateLimit.Enabled == false {
		return &noopRateLimiter{}
	}

	return &iPRateLimiter{
		userKeyWhitelist:     toSet(config.IPRateLimit.KeyWhitelist),
		ipLimits:             make(map[string]*rate.Limiter),
		recoverXTokensPerSec: rate.Limit(config.IPRateLimit.RecoverXTokensPerSec),
		maxTokens:            config.IPRateLimit.MaxTokens,
		allowIPFromHeader:    config.IPRateLimit.FetchIPFromHeaders,
	}
}
