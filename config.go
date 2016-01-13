package ghttp

import (
	"github.com/geeksteam/ghttp/api"
	"github.com/geeksteam/ghttp/bruteforce"
	"github.com/geeksteam/ghttp/journal"
	"github.com/geeksteam/ghttp/sessions"
	"github.com/geeksteam/ghttp/utemplates"
)

type Config struct {
	MaxHandlersForUser int    `default:"30" comment:"Max allowed number of simultaneous queries for single user."`
	Version            string `default:"0.1.1alpha"`
	WebServerName      string `default:"SHM API server"`
	CacheLifetime      int    `default:"0" comment:"Cache lifetime in days for static files (images,css, etc)"`

	bruteforce.BruteForce
	journal.Journal
	api.API
	sessions.SessionsConf
	utemplates.Utemplates
}
