package ghttp

import (
	"sync"

	"github.com/geekbros/ghttp/sessions"
	"github.com/gorilla/mux"
)

// rhandler is a running handler struct representation.
type rhandler struct {
	id        uint64 // Handler ID
	URI       string // URI handler ( /backups )
	IP        string // Client IP
	Username  string // Client username
	StartTime string // Date time of handler's start
	SessionID string // Session id if exist for this hanfdler
}

// Router is a custom gorilla's Router wrapper.
type Router struct {
	curID      uint64              // Counter total handlers done
	handlers   map[uint64]rhandler // List of running handlers
	Sessions   *sessions.Sessions  // User's sessions
	mutex      sync.RWMutex
	mux.Router // Include mux router composition
}
