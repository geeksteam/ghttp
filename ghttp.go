package ghttp

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/geeksteam/GoTools/logger"
	"github.com/geeksteam/GoTools/shutdown"
	"github.com/geeksteam/GoTools/stringutils"
	"github.com/geeksteam/SHM-Backend/core/users"
	"github.com/geeksteam/SHM-Backend/panicerr"
	"github.com/geeksteam/SHM-Backend/plugins"
	"github.com/geeksteam/ghttp/bruteforce"
	"github.com/geeksteam/ghttp/journal"
	"github.com/geeksteam/ghttp/moduleutils"
	"github.com/geeksteam/ghttp/sessions"
	"github.com/getsentry/raven-go"
	"github.com/gorilla/mux"
)

var (
	// cfg = config.Get().GHttp
	cfg Config
)

// SetConfig Global config settings
func SetConfig(c Config) {
	cfg = c
	bruteforce.SetConfig(bruteforce.BruteForce{
		BlockAttempts: cfg.BruteForce.BlockAttempts,
		BanTime:       cfg.BruteForce.BanTime,
		DataEncoding:  cfg.BruteForce.DataEncoding,
	})
	journal.SetConfig(journal.Journal{
		BoltDB:              cfg.Journal.BoltDB,
		BucketForOperations: cfg.Journal.BucketForOperations,
		Capacity:            cfg.Journal.Capacity,
		DataEncoding:        cfg.Journal.DataEncoding,
	})
	sessions.SetConfig(sessions.SessionsConf{
		SessionIDKey:       cfg.SessionsConf.SessionIDKey,
		SessionIDKeyLength: cfg.SessionsConf.SessionIDKeyLength,
		SessionLifeTime:    cfg.SessionsConf.SessionLifeTime,
		StrictIP:           cfg.SessionsConf.StrictIP,
	})
}

// NewRouter constructs Router instances
func NewRouter() *Router {
	sessions.NewSessions()
	return &Router{
		curID:    0,
		handlers: map[uint64]rhandler{},
		mutex:    sync.RWMutex{},
		Router:   *mux.NewRouter(),
	}
}

// Handlers returns copy of an internal Router's handlers list.
func (r *Router) Handlers() map[uint64]rhandler {
	// Make it atomic
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	// Make a copy of handlers slice
	result := make(map[uint64]rhandler, len(r.handlers))
	for k, v := range r.handlers {
		result[k] = v
	}
	return result
}

// Delete handler from slice
func (r *Router) deleteHandler(id uint64) {
	delete(r.handlers, id)
}

// HandleInternalFunc is a gorilla Router's wrapper function.
// This handles standart modules functions.
func (router *Router) HandleInternalFunc(path string, f func(http.ResponseWriter, *http.Request, *sessions.Sessions)) *mux.Route {
	routerFunc := func(w http.ResponseWriter, r *http.Request) {
		// Trigger starting of new process
		if !isIgnored(r.RequestURI) {
			if err := shutdown.DefaultWatcher.Start(); err != nil {
				http.Error(w, http.StatusText(http.StatusServiceUnavailable), http.StatusServiceUnavailable)
				return
			}
			defer shutdown.DefaultWatcher.Finish()
		}

		/*
			# Pre run
		*/
		// 1. Clean expired sessions
		sessions.SessionsStorage.CleanExpired()

		// 2. Define empty handler
		var handler rhandler

		// 3. Set headers
		setHeaderNoCache(w)

		// 4. Defered run and catch panics
		defer func() {
			// Remove handler from running list
			router.mutex.Lock()
			router.deleteHandler(handler.id)
			router.mutex.Unlock()

			// Catch all panics from handler
			if rec := recover(); rec != nil {
				switch rec := rec.(type) {
				// Panicerr catched
				case panicerr.Error:
					logger.Info(fmt.Sprintf("%v [ %v ] Error catched: Code '%v' Text '%v'", strings.Split(r.RemoteAddr, ":")[0], r.RequestURI, rec.Code, rec.Err))
					//Send response with json error description
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(500)
					w.Write([]byte(rec.ToJSONString()))
				// Unknown panic
				default:
					raven.CaptureError(fmt.Errorf("%v", rec), nil)
					logger.Error(fmt.Sprintf("%v [ %v ] Unknown Error catched: '%v'", strings.Split(r.RemoteAddr, ":")[0], r.RequestURI, rec))
					http.Error(w, http.StatusText(500), 500)
					panic(rec)
				}
			}
		}()

		/*
			# Permissions checks
		*/
		// 1. Prevent bruteforce of sessionID
		ok, duration := bruteforce.Check(strings.Split(r.RemoteAddr, ":")[0])
		if !ok {
			logger.Warning("IP " + strings.Split(r.RemoteAddr, ":")[0] + " banned by bruteforce (no session) for " + strconv.FormatInt(duration, 10) + " sec.")
			http.Error(w, http.StatusText(429), 429)
			return
		}

		// 2. Check if session started and getting session info
		sess, err := sessions.SessionsStorage.Get(r)
		if err != nil {
			logger.Warning(err.Error())
			http.Error(w, http.StatusText(401), 401)
			return
		}

		// 3. Clear IP in bruteforce check
		bruteforce.Clean(strings.Split(r.RemoteAddr, ":")[0])

		// 4. Check for timeout before actions for particular handlers
		if err := bruteforce.CheckTimeout(r, sessions.SessionsStorage); err != nil {
			http.Error(w, http.StatusText(429), 429)
			log.Println("Timeout error: ", err)
			return
		}

		// 5. Register session activity for sessions timeout
		sessions.SessionsStorage.RegisterActivity(r)

		// 6. Check module access permisions
		if sess.Username != "root" {
			userInfo := users.Get(sess.Username)
			if userInfo == nil {
				panicerr.Core.Auth("Can't get template" + sess.Username)
			}

			allowedModules := userInfo.GetTemplate().Modules
			if err != nil || !hasPermissions(r.RequestURI, allowedModules) {
				http.Error(w, http.StatusText(403), 403)
				logger.Warning(fmt.Sprintf("Permission denied to access '%v' for %v as user %v", r.RequestURI, strings.Split(r.RemoteAddr, ":")[0], sess.Username))
				return
			}
		}

		// 7. Check for simultaneous connections from a single user
		router.CheckNumConnection(sess.Username)

		/*
			# Append handler to running list for tracking
		*/
		func() {
			// Make it atomic
			router.mutex.Lock()
			defer router.mutex.Unlock()
			// Increment handler ID
			router.curID++
			// create handler struct
			handler = rhandler{
				id:        router.curID,
				URI:       r.RequestURI,
				Username:  sess.Username,
				IP:        r.RemoteAddr,
				StartTime: time.Now().Format(time.StampMilli),
				SessionID: sess.ID,
			}
			// Append new handler to list
			router.handlers[router.curID] = handler
		}()

		/*
			# Add action to journal
		*/

		journal.Add(journal.Operation{
			SessionID: sess.ID,
			Date:      time.Now().Format(journal.TimeLayout),
			Username:  sess.Username,
			Operation: moduleutils.GetCurrentModule(r.RequestURI),
			Content:   r.RequestURI,
			//Extra:
		})

		/*
			# Run handler's function
		*/
		f(w, r, sessions.SessionsStorage)

		/*
			# Make api trigger call
		*/
		plugins.DefaultManager.Trigger(w, r, sess)
	}
	// Insert func to gorilla/mux router
	return router.HandleFunc(path, routerFunc)
}

// HandleLoginFunc is uniq handler for Authorization and create new session only
func (router *Router) HandleLoginFunc(path string, f func(http.ResponseWriter, *http.Request, *sessions.Sessions)) *mux.Route {
	routerFunc := func(w http.ResponseWriter, r *http.Request) {
		/*
			Set headers
		*/
		setHeaderNoCache(w)
		/*
			Defered run and catch panics
		*/
		defer func() {
			// Catch all panics from handler
			if rec := recover(); rec != nil {
				switch rec := rec.(type) {
				// Panicerr catched
				case panicerr.Error:
					logger.Info(fmt.Sprintf("%v [ %v ] Error catched: Code '%v' Text '%v'", strings.Split(r.RemoteAddr, ":")[0], r.RequestURI, rec.Code, rec.Err))
					//Send response with json error description
					w.WriteHeader(500)
					w.Write([]byte(rec.ToJSONString()))
				// Unknown panic
				default:
					raven.CaptureError(fmt.Errorf("%v", rec), nil)
					logger.Error(fmt.Sprintf("%v [ %v ] Unknown Error catched: '%v'", strings.Split(r.RemoteAddr, ":")[0], r.RequestURI, rec))
					http.Error(w, http.StatusText(500), 500)
					panic(rec)
				}
			}
		}()

		/*
			Run handler's function
		*/
		f(w, r, sessions.SessionsStorage)

		/*
			Make api trigger call
		*/
		plugins.DefaultManager.Trigger(w, r, nil)
	}
	// Insert func to gorilla/mux router
	return router.HandleFunc(path, routerFunc)
}

// CheckNumConnection - Checking for number of simultaneous requests for user
// panicerr if exceeded
func (router *Router) CheckNumConnection(username string) {
	handlersCount := 0

	router.mutex.RLock()
	defer router.mutex.RUnlock()
	for _, handler := range router.handlers {
		if handler.Username == username {
			handlersCount++
			if handlersCount > cfg.MaxHandlersForUser {
				panicerr.Handlers.RequestsExceeded(fmt.Sprint("Exceeded the number of simultaneous requests for user (", cfg.MaxHandlersForUser, ")"))
			}
		}
	}
}

// Check for user permissions to module for /uri
func hasPermissions(path string, modules []string) bool {
	// If no modules allow access
	if len(modules) == 0 {
		return true
	}
	// Iterate and check for each module in grant list
	for _, userModule := range modules {
		if userModule == moduleutils.GetCurrentModule(path) {
			return true
		}
	}
	return false
}

// Set http headers to no-cache, content json
func setHeaderNoCache(w http.ResponseWriter) {
	w.Header().Set("Server", cfg.WebServerName)
	w.Header().Set("Version", cfg.Version)
	w.Header().Set("Cache-Control", "post-check=0, pre-check=0, no-store, no-cache, must-revalidate, max-age=0")
	w.Header().Set("Pragma", "no-cache")
	w.Header().Set("Expires", "-1")
	w.Header().Set("Content-Type", "application/json")
}

func isIgnored(path string) bool {
	var ignored = []string{
		"/api/info/actualizer/",
		"/api/shell/console",
		"/api/core/livesysstat",
	}

	return stringutils.Contains(ignored, path)
}
