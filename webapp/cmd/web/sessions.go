package main

import (
	"net/http"
	"time"

	"github.com/alexedwards/scs/v2"
)

func getSession() *scs.SessionManager {
	session := scs.New()
	session.Lifetime = 24 * time.Hour // session will last for 24 hours
	session.Cookie.Persist = true     // decides whether the cookies should persist
	session.Cookie.SameSite = http.SameSiteLaxMode
	session.Cookie.Secure = true // enables use of encrypted cookies in production
	return session
}
