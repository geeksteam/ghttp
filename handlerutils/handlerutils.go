package handlerUtils

import (
	"compress/gzip"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/geeksteam/GoTools/sysutils/osuser"
	"github.com/geeksteam/SHM-Backend/panicerr"
	"github.com/geeksteam/ghttp/sessions"

	"github.com/gorilla/websocket"
)

// GetUID returns UID of current session user
func GetUID(r *http.Request, s *sessions.Sessions) (int, error) {
	//get os user which uses current session
	user, err := GetOsUser(r, s)
	if err != nil {
		return 0, err
	}

	return int(user.UID), nil
}

// GetOsUser returns os user info corresponding to session
func GetOsUser(r *http.Request, s *sessions.Sessions) (osuser.OsUser, error) {
	//get current session user
	username, err := GetSessionUser(r, s)
	if err != nil {
		return osuser.OsUser{}, err
	}

	//get os user which uses current session
	user, err := osuser.GetUserWithUsername(username)
	if err != nil {
		return osuser.OsUser{}, err
	}

	return user, nil

}

// GetSessionUser returns username  of current session user
func GetSessionUser(r *http.Request, s *sessions.Sessions) (string, error) {
	//get current session
	sess, err := s.Get(r)
	if err != nil {
		return "", err
	}
	return sess.Username, nil
}

// ParseJSONBody attempts to parse r's body concerning that it is  a valid json struct.
// If somethings goes wrong, it makes valid error response to client.
func ParseJSONBody(w http.ResponseWriter, r *http.Request, jsonStruct interface{}) {
	defer r.Body.Close()
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(jsonStruct)
	if err != nil {
		panicerr.JSON.ParsingError(err)
	}
}

// WriteJSONBody Marshal interface{} to JSON and write it to response
func WriteJSONBody(w http.ResponseWriter, jsonStruct interface{}) {
	js := json.NewEncoder(w)

	if err := js.Encode(jsonStruct); err != nil {
		panicerr.JSON.EncodingError(err)
	}
}

// SendOkStatus Sends Status 204 Accepted for successful requests
func SendOkStatus(w http.ResponseWriter) {
	// Send HTTP 204 everything ok but no content
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusNoContent)
}

// AddExpiresHeaderAndServe Writes header for cache control
func AddExpiresHeaderAndServe(h http.Handler, cacheLifeTime int) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		// Send Expires -1 if cache 0
		if cacheLifeTime == 0 {
			w.Header().Set("Cache-Control", "post-check=0, pre-check=0, no-store, no-cache, must-revalidate, max-age=0")
			w.Header().Set("Pragma", "no-cache")
			w.Header().Set("Expires", "-1")
		} else {
			// Calculate expires date time
			cacheSince := time.Now().UTC().Format(http.TimeFormat)
			cacheUntil := time.Now().AddDate(0, 0, cacheLifeTime).UTC().Format(http.TimeFormat)
			w.Header().Set("Cache-Control", "max-age:290304000, public")
			w.Header().Set("Last-Modified", cacheSince)
			w.Header().Set("Expires", cacheUntil)
		}
		// Serve with the actual handler.
		h.ServeHTTP(w, r)
	}
}

type gzipResponseWriter struct {
	io.Writer
	http.ResponseWriter
	sniffDone bool
}

func (w *gzipResponseWriter) Write(b []byte) (int, error) {
	if !w.sniffDone {
		if w.Header().Get("Content-Type") == "" {
			w.Header().Set("Content-Type", http.DetectContentType(b))
		}
		w.sniffDone = true
	}
	return w.Writer.Write(b)
}

// MakeGzipHandler Wrap a http.Handler to support transparent gzip encoding.
func MakeGzipHandler(h http.Handler, cacheLifeTime int) http.Handler {
	return AddExpiresHeaderAndServe(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Vary", "Accept-Encoding")
		if !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
			h.ServeHTTP(w, r)
			return
		}
		w.Header().Set("Content-Encoding", "gzip")
		gz := gzip.NewWriter(w)
		defer gz.Close()
		h.ServeHTTP(&gzipResponseWriter{Writer: gz, ResponseWriter: w}, r)
	}), cacheLifeTime)
}

func GetWebSocketConnection(w http.ResponseWriter, r *http.Request) (*websocket.Conn, error) {
	upgrader := websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return nil, err
	}

	return conn, nil
}
