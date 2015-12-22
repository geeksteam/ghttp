package api

import "sync"

// apiTrigger is a main entity through which all api calls are made.
type apiTrigger struct {
	Triggers map[string]int
	*sync.Mutex
}

// Call represents api call, sent to a corresponding "/api/*" file
// as JSON.
type Call struct {
	Session map[string]string
	Get     map[string]string
	Post    map[string]string
	Stdin   interface{}
}

// TriggerInfo holds info about Trigger's calls count.
type TriggerInfo struct {
	Trigger string `json:"Trigger"`
	Count   int    `json:"Count"`
}
