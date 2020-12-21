package dohboy

import (
	"net"
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
		for _, key := range []string{"X-Forwarded-For", "X-Real-Ip"} {
			if val := request.Header.Get(key); val != "" {
				addrs := strings.Split(strings.Trim(val, ","), ",")
				last := addrs[len(addrs)-1]
				if ip := net.ParseIP(last); ip != nil {
					return ip.String()
				}
			}
		}
	}

	if ip, _, err := net.SplitHostPort(request.RemoteAddr); err == nil {
		return ip
	}
	return ""
}

func toSet(commaSeparated string) set {
	retval := newSet()

	if commaSeparated != "" {
		for _, key := range strings.Split(commaSeparated, ",") {
			retval.Add(strings.TrimSpace(key))
		}
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
