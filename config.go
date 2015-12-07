package ghttp

import (
	"github.com/geekbros/ghttp/api"
	"github.com/geekbros/ghttp/bruteforce"
	"github.com/geekbros/ghttp/journal"
	"github.com/geekbros/ghttp/sessions"
	"github.com/geekbros/ghttp/utemplates"
)

type Config struct {
	MaxHandlersForUser int    `default:"30" comment:"Max allowed number of simultaneous queries for single user."`
	Version            string `default:"0.1.1alpha"`
	WebServerName      string `default:"SHM API server"`

	bruteforce.BruteForce
	journal.Journal
	api.API
	sessions.SessionsConf
	utemplates.Utemplates
}
