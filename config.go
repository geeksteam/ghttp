package ghttp

import (
	"github.com/geekbros/ghttp/api"
	"github.com/geekbros/ghttp/bruteforce"
	"github.com/geekbros/ghttp/journal"
	"github.com/geekbros/ghttp/sessions"
	"github.com/geekbros/ghttp/utemplates"
)

type Config struct {
	// CacheLifetime int `default:"0" comment:"Cache lifetime in days for static files (images,css, etc)"`
	//
	// SessionIDKey       string `default:"sessionID" comment:"Key of session id in cookies map, which generates randomly."`
	// SessionIDKeyLength int    `default:"24" comment:"Length of session id key for random generation."`
	//
	// SessionLifeTime    int `default:"1800" comment:"Lifetime of a session. Seconds."`
	// MaxHandlersForUser int `default:"30" comment:"Max allowed number of simultaneous queries for single user."`
	MaxHandlersForUser int `default:"30" comment:"Max allowed number of simultaneous queries for single user."`
	bruteforce.BruteForce
	journal.Journal
	api.API
	sessions.SessionsConf
	utemplates.Utemplates
}
