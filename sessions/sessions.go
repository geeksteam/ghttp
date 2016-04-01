package sessions

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/geeksteam/GoTools/deepcopy"
	"github.com/geeksteam/GoTools/stringutils"
)

var (
	// cfg = config.Get().GHttp
	cfg SessionsConf

	errNoSessionID     = errors.New("Session not found: no sessionID set in cookies.")
	errNoSessionWithID = errors.New("Session not found: no sessions with given sessionID found.")
	errIP              = errors.New("Session's IP and current user's IP are not equal.")
	errValue           = errors.New("Value not found.")
)

func SetConfig(c SessionsConf) {
	cfg = c
}

type (
	// TempFile represents temporary file which is uploaded by authorized user.
	TempFile struct {
		Filename    string `json:"Filename"`
		TmpFileName string `json:"TmpFileName"`
		Created     string `json:"Created"`
		Size        int64  `json:"Size"`
	}

	// TempFiles is a wrapper of []TempFile type, which provides additional functionality.
	TempFiles []TempFile

	// TODO:
	// comment
	ActualizeListener struct {
		MessageChan chan string
		CloseChan   chan bool
		IsListening bool
	}
)

// GetTempFileName returns actual temporary file name, which had given original name.
// Returns empty string if no files with given original name are found.
func (t TempFiles) GetTempFileName(originalName string) (tempName string) {
	for _, v := range t {
		if v.Filename == originalName {
			return v.TmpFileName
		}
	}
	return
}

// GetOriginalFileName returns actual file name of file with given temp name.
// Returns empty string if no files with given temp name are found.
func (t TempFiles) GetOriginalFileName(tempFileName string) (originalName string) {
	for _, v := range t {
		if v.TmpFileName == tempFileName {
			return v.Filename
		}
	}
	return
}

// Session stores info for user session
type Session struct {
	ID           string // ID сессии (сгенерированный из 256 символов)
	IP           string // IP адрес клиента
	Created      int64  // Дата и время создания в секундах UNIXTIME сейчас - UNIXTIME от 01/01/2015 года.
	LastActivity int64  // Дата последней активности пользователя в текущей сессии
	Username     string // Имя юзера под которым авторизирована сессия
	UserAgent    string // UserAgent пользователя
	// UserInfo     *users.UserInfo    `json:"-"` // Параметры юзера
	Theme      string
	Language   string
	Template   string
	Uploads    TempFiles          // Current sessions temp files list
	Actualizer *ActualizeListener `json:"-"`

	LastHandlers map[string]int64 // /handler and unixtime of last request
}

// Sessions is a general service, which handles sessions.
type Sessions struct {
	sessions map[string]Session // Список сессий
	sync.RWMutex
}

// NewSessions is a Sessions constructor.
func NewSessions() *Sessions {
	return &Sessions{make(map[string]Session), sync.RWMutex{}}
}

// Get attempts to get session from local sessions map.
func (s *Sessions) Get(r *http.Request) (*Session, error) {

	// Getting SessID from cookie
	cookie, err := r.Cookie(cfg.SessionIDKey)
	if err != nil {
		return nil, errNoSessionID
	}
	sessionID := cookie.Value

	// Make it atomic
	s.RLock()
	defer s.RUnlock()

	// Check for session exist
	sess, ok := s.sessions[sessionID]
	if !ok {
		return nil, errNoSessionWithID
	}

	// Check session IP and client IP if StrictIP on
	if cfg.StrictIP {
		if sess.IP != strings.Split(r.RemoteAddr, ":")[0] {
			return nil, errIP
		}
	}

	return deepcopy.Iface(&sess).(*Session), nil
}

// AddTempFile adds given TempFile struct to current session's internal tempfiles
// list.
func (s *Sessions) AddTempFile(r *http.Request, f TempFile) error {
	sess, err := s.Get(r)
	if err != nil {
		return err
	}
	sess.Uploads = append(sess.Uploads, f)
	err = s.Set(r, *sess)
	if err != nil {
		return err
	}
	return nil
}

// ClearTempFiles for remov
func (s *Sessions) ClearTempFiles(r *http.Request) error {
	sess, err := s.Get(r)
	if err != nil {
		return err
	}
	sess.Uploads = []TempFile{}
	err = s.Set(r, *sess)
	if err != nil {
		return err
	}
	return nil
}

// GetTempFiles returns all TempFiles, registered to current session.
func (s *Sessions) GetTempFiles(r *http.Request) (result TempFiles, err error) {
	sess, err := s.Get(r)
	if err != nil {
		return TempFiles{}, err
	}
	return sess.Uploads, nil
}

// IsExist Check if session for user which send given request exist. It returns true if
// request contains cookie with SessionIDKey and sessions map contains entry with key sessionID from cookie
func (s *Sessions) IsExist(r *http.Request) bool {
	// Getting cookie
	cookie, err := r.Cookie(cfg.SessionIDKey)
	// No sessionID in request's cookies.
	if err != nil {
		return false
	}
	sessionID := cookie.Value
	// Atomic Check if sssion with given id contains in sessions map
	s.RLock()
	_, ok := s.sessions[sessionID]
	s.RUnlock()

	return ok
}

// ListenActualizer Sets parameter Actualizer.IsListening to isListen
func (s *Sessions) ListenActualizer(r *http.Request, isListen bool) error {
	cookie, err := r.Cookie(cfg.SessionIDKey)
	if err != nil {
		return err
	}

	sessionID := cookie.Value

	// Atomic
	s.Lock()
	defer s.Unlock()

	sess, ok := s.sessions[sessionID]
	if !ok {
		return fmt.Errorf("No session with id %v found", sessionID)
	}

	sess.Actualizer.IsListening = isListen
	s.sessions[sessionID] = sess

	return nil
}

// RegisterActivity Update time of last user activity
func (s *Sessions) RegisterActivity(r *http.Request) error {
	cookie, err := r.Cookie(cfg.SessionIDKey)
	if err != nil {
		return err
	}

	sessionID := cookie.Value

	// Atomic
	s.Lock()
	defer s.Unlock()

	sess, ok := s.sessions[sessionID]
	if !ok {
		return fmt.Errorf("No session with id %v found", sessionID)
	}

	sess.LastActivity = time.Now().Unix()
	sess.LastHandlers[r.RequestURI] = time.Now().Unix()

	s.sessions[sessionID] = sess

	return nil
}

// StartNewSession Create new session for user, generate sessionID and write new cookie into response
// If session for user already exist it returns error
func (s *Sessions) StartNewSession(r *http.Request, w http.ResponseWriter, username string) *Session {

	sessionID := stringutils.GetRandomString(cfg.SessionIDKeyLength)
	http.SetCookie(w, &http.Cookie{Name: cfg.SessionIDKey, Value: sessionID, Path: "/"})

	sess := Session{
		ID:        sessionID,
		IP:        strings.Split(r.RemoteAddr, ":")[0],
		Username:  username,
		UserAgent: r.UserAgent(),
		// 	UserInfo:     users.Get(username),
		Created:      time.Now().Unix(),
		LastActivity: time.Now().Unix(),
		Actualizer: &ActualizeListener{
			MessageChan: make(chan string, 10),
			CloseChan:   make(chan bool, 10),
			IsListening: false,
		},

		LastHandlers: make(map[string]int64),
	}
	// Append new session to map
	// Atomic
	s.Lock()
	defer s.Unlock()

	s.sessions[sessionID] = sess
	return deepcopy.Iface(&sess).(*Session)
}

// Set attempts to reset current user's session struct to given session struct.
func (s *Sessions) Set(r *http.Request, session Session) error {
	s.RLock()
	defer s.RUnlock()
	sess, err := s.Get(r)
	if err != nil {
		return err
	}
	s.sessions[sess.ID] = session
	return nil
}

// GetAll fetches copy of a map of current active sessions.
func (s *Sessions) GetAll() map[string]Session {
	s.Lock()
	defer s.Unlock()
	return deepcopy.Iface(s.sessions).(map[string]Session)
}

// CleanExpired deletes all sessions, whose lifetime has already passed.
func (s *Sessions) CleanExpired() {
	sessForKill := []string{}

	s.RLock()
	for k, v := range s.sessions {
		// if time.Now().Unix()-v.LastActivity >= int64(cfg.SessionLifeTime) {
		// 	delete(s.sessions, k)
		// }
		if time.Now().After(time.Unix(v.LastActivity, 0).Add(time.Duration(cfg.SessionLifeTime) * time.Second)) {
			sessForKill = append(sessForKill, k)
		}
	}
	s.RUnlock()

	for _, sessID := range sessForKill {
		s.DelByID(sessID)
	}
}

// Del deletes session, which corresponds to given request.
func (s *Sessions) Del(r *http.Request, w http.ResponseWriter) {
	// Check for cookie
	cookie, err := r.Cookie(cfg.SessionIDKey)
	if err != nil {
		return
	}

	s.DelByID(cookie.Value)

	// Remove cookie
	cookie.MaxAge = -1
	http.SetCookie(w, cookie)
}

// DelByID Atomic delete session by ID
func (s *Sessions) DelByID(sessionID string) {
    // Lock for mutex
	s.Lock()
    defer s.Unlock()
	// check for session exist in map and delete
	_, ok := s.sessions[sessionID]
	if !ok {
        log.Println("Trying to remove unexistent sessionID:", sessionID)
        return
	}
    //close websocket
	s.sessions[sessionID].Actualizer.CloseChan <- true
    delete(s.sessions, sessionID)
}
func (s *Sessions) getUserSessions(username string) []Session {
	s.Lock()
	defer s.Unlock()

	sessions := []Session{}

	for _, v := range s.sessions {
		if v.Username == username {
			sessions = append(sessions, v)
		}
	}

	return sessions
}

// Actualize Send message to Actualizer chan. Uses for instant GUI updating through websocket
func (s *Sessions) Actualize(username, message string) {
	sessions := s.getUserSessions(username)

	for _, sess := range sessions {
		if sess.Actualizer.IsListening {

			sess.Actualizer.MessageChan <- message
		}
	}
}
