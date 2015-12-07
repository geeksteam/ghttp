package bruteforce

import (
	"fmt"
	"net/http"
	"time"

	"github.com/geekbros/ghttp/sessions"
)

var (
	// Timeouts: api request and timeout between next run in seconds.
	timeouts = map[string]int64{
		"/api/messages/send":          10,
		"/api/filemanager/pack":       10,
		"/api/filemanager/getarchive": 10,
		"/api/support/bugreport":      300,
	}
)

// CheckTimeout Check timeout for concrete handler
func CheckTimeout(r *http.Request, s *sessions.Sessions) error {
	// Check if current uri in map of timeouts
	timeout, ok := timeouts[r.URL.Path]
	if !ok {
		return fmt.Errorf("No info about timeout of %v", r.URL.Path)
	}
	sess, _ := s.Get(r)
	lastRequest, ok := sess.LastHandlers[r.RequestURI]
	if !ok {
		return fmt.Errorf("No info about last request of %v", r.RequestURI)
	}
	if time.Now().Unix()-lastRequest <= timeout {
		return fmt.Errorf(fmt.Sprint("One request '"+r.URL.Path+"' per ", timeout, " seconds limit."))
	}
	return nil
}
