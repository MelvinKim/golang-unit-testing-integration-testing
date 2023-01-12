package main

import (
	"context"
	"fmt"
	"net"
	"net/http"
)

type contextKey string

const contextUserKey contextKey = "user_ip"

func (app *application) ipFromContext(ctx context.Context) string {
	return ctx.Value(contextUserKey).(string)
}

func (app *application) addIpToContext(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// we need to have a context to work with
		var ctx = context.Background()

		// get the ip (as accurately as possible)
		// ip := r.RemoteAddr, returns the IP address of the request , not always accurate especially when the client is behind a proxy
		ip, err := getIP(r)
		if err != nil {
			ip, _, _ := net.SplitHostPort(r.RemoteAddr)
			if len(ip) == 0 {
				ip = "unknown"
			}

			// add the IP value to the context
			ctx = context.WithValue(r.Context(), contextUserKey, ip)
		} else {
			// add the IP value to the context
			ctx = context.WithValue(r.Context(), contextUserKey, ip)
		}
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func getIP(r *http.Request) (string, error) {
	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return "unknown", err
	}

	userIP := net.ParseIP(ip)
	if userIP == nil {
		return "", fmt.Errorf("userip: %q is not IP:port", r.RemoteAddr)
	}

	// check if the IP  is coming from a proxy
	forward := r.Header.Get("X-Forwarded-For")
	if len(forward) > 0 {
		ip = forward
	}

	if len(ip) == 0 {
		ip = "forward"
	}

	return ip, nil
}
