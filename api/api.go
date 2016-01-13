package api

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/geeksteam/GoTools/executils"
)

var (
	cfg API

	// Trigger is a main entity for making api calls.
	Trigger = apiTrigger{make(map[string]int), &sync.Mutex{}}

	errNoTrigger = errors.New("No such trigger in api ")
	errCmdPipe   = errors.New("Error while getting cmd pipe.")
	errCmdStart  = errors.New("Error while starting cmd.")
)

func SetConfig(c API) {
	cfg = c
}

// Read reads api path directory and adds all possible api endpoints to it's
// internal Triggers list.
func (t *apiTrigger) Read() {
	t.Lock()
	defer t.Unlock()
	triggers := readDir(cfg.ApiPath, true)
	for _, v := range triggers {
		t.Triggers[v] = 0
	}
}

// GetTriggersInfo returns all registered api triggers.
func (t apiTrigger) GetTriggersInfo() []TriggerInfo {
	t.Lock()
	defer t.Unlock()
	triggers := []TriggerInfo{}
	for k, v := range t.Triggers {
		triggers = append(triggers, TriggerInfo{k, v})
	}
	return triggers
}

// MakeAPICall is a middleware function, which is used in all ghttp.Router.HandleFuncRegistered
// calls. Checks if current call is an API call, if true, makes this call.
func (t *apiTrigger) Call(sessionParams map[string]string, w http.ResponseWriter, r *http.Request) {
	// Get actual api endpoint.
	path := strings.Split(r.RequestURI, "?")[0]
	path = strings.TrimPrefix(path, "/api/")
	path = strings.TrimPrefix(path, "/")

	// Check if the corresponding api file exists.
	if !t.hasTrigger(path) {
		return
	}

	log.Println("Triggered " + path)
	t.Triggers[path]++

	fullpath := filepath.Join(cfg.ApiPath, path)
	currentDir, _ := os.Getwd()

	// Running a file.
	cmd := executils.Command(fullpath)
	pipe, err := cmd.StdinPipe()
	if err != nil {
		log.Println(errCmdPipe)
		return
	}

	// Executing api file.
	err = cmd.StartWithDefaultTimeout()
	if err != nil {
		log.Println(errCmdStart)
		return
	}

	// Getting base request data: get, post query parameters and session data.
	call := newAPICall(sessionParams, r)

	// Marshaling api call struct.
	resultingJSON, _ := json.Marshal(call)

	// Send resulting JSON api call to pipe.
	pipe.Write(resultingJSON)
	pipe.Close()

	// Go back to original working directory.
	os.Chdir(currentDir)
}

// hasTrigger checks if internal triggers list containts given trigger.
func (t *apiTrigger) hasTrigger(trigger string) bool {
	t.Lock()
	defer t.Unlock()
	for k := range t.Triggers {
		if k == trigger {
			return true
		}
	}
	return false
}

func readDir(path string, isRoot bool) []string {
	var pathPrefix = path + "/"
	endpoints := []string{}

	if dir, err := os.Open(path); err == nil {
		if fi, err := dir.Readdir(0); err == nil {
			for _, v := range fi {
				if !v.IsDir() {
					if filepath.Ext(v.Name()) != "" {
						continue
					}
					endpoints = append(endpoints, pathPrefix+v.Name())
				} else {
					endpoints = append(endpoints, readDir(pathPrefix+v.Name(), false)...)
				}
			}
		}
	}

	if isRoot {
		for i, val := range endpoints {
			endpoints[i] = strings.TrimPrefix(val, pathPrefix)
		}
	}

	return endpoints
}

// newAPICall constructs APICall struct.
func newAPICall(sessionsParams map[string]string, r *http.Request) Call {
	r.ParseForm()
	getParams := getQueryParamsMap(r.Form)
	postParams := getQueryParamsMap(r.PostForm)

	// Getting a "stdin" call section from request's body.
	var stdin map[string]interface{}
	content, err := ioutil.ReadAll(r.Body)
	if err == nil {
		json.Unmarshal(content, &stdin)
	}

	// Creating a corresponding api call struct.
	call := Call{
		Session: sessionsParams,
		Get:     getParams,
		Post:    postParams,
		Stdin:   stdin,
	}
	return call
}

// getQueryParamsMap simply converts map[string][]string to map[string]string.
// Supposed to be used to get more suitable representation of Form and PostForm
// values.
func getQueryParamsMap(params map[string][]string) map[string]string {
	result := make(map[string]string)
	for k, v := range params {
		var newValue string
		for _, i := range v {
			newValue += i
		}
		result[k] = newValue
	}
	return result
}
