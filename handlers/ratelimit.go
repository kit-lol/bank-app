package handlers

import (
	"net/http"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

// ipLimiter хранит лимитер для конкретного IP-адреса
type ipLimiter struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

var (
	limiters = make(map[string]*ipLimiter)
	mu       sync.Mutex
)

// init запускает фоновую очистку старых записей каждые 3 минуты
func init() {
	go func() {
		for {
			time.Sleep(3 * time.Minute)
			mu.Lock()
			for ip, lim := range limiters {
				if time.Since(lim.lastSeen) > 5*time.Minute {
					delete(limiters, ip)
				}
			}
			mu.Unlock()
		}
	}()
}

// getLimiter возвращает rate limiter для данного IP
func getLimiter(ip string) *rate.Limiter {
	mu.Lock()
	defer mu.Unlock()

	if lim, exists := limiters[ip]; exists {
		lim.lastSeen = time.Now()
		return lim.limiter
	}

	// 5 запросов в минуту, burst до 10
	lim := rate.NewLimiter(rate.Every(time.Minute/5), 10)
	limiters[ip] = &ipLimiter{limiter: lim, lastSeen: time.Now()}
	return lim
}

// RateLimitMiddleware ограничивает количество запросов с одного IP
func RateLimitMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ip := r.RemoteAddr
		if !getLimiter(ip).Allow() {
			http.Error(w, "Слишком много запросов. Попробуйте позже.", http.StatusTooManyRequests)
			return
		}
		next(w, r)
	}
}
