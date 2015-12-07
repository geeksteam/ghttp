package bruteforce

import (
	"sync"
	"time"
)

var (
	cfg BruteForce
	// Local variables
	iPs   = make(map[string]BruteIP)
	mutex sync.RWMutex
)

func SetConfig(c BruteForce) {
	cfg = c
}

// Clean - set number of checks for IP to zero.
// Used after success login and etc.
func Clean(IP string) {
	// Make it atomic
	mutex.Lock()
	defer mutex.Unlock()

	iPs[IP] = BruteIP{
		Timestamp: time.Now().Unix(),
		Attempts:  1,
	}
}

// Check checks given IP according to it's activity history. Returns true if
// ip is still doesn't have to be banned.
func Check(IP string) (bool, int64) {

	// Clean all expired BruteIP instances.
	clean()
	// Make it atomic
	mutex.Lock()
	defer mutex.Unlock()

	bi, ok := iPs[IP]

	// No BruteIP instances for given IP - create new.
	if !ok {
		iPs[IP] = BruteIP{
			Timestamp: time.Now().Unix(),
			Attempts:  1,
		}
		return true, -1
	}
	// Ban condition: given IP made more attempts that are allowed and it's
	// ban has not expired.
	bi.Timestamp = time.Now().Unix()

	if bi.Attempts > cfg.BlockAttempts {
		iPs[IP] = bi
		return false, cfg.BanTime
	}
	// Ban attempts are not exceeded - update corresponding BruteIP instancesinfo.
	bi.Attempts++
	iPs[IP] = bi
	return true, -1

}

// Remove expired banned ips
func clean() {
	mutex.Lock()
	defer mutex.Unlock()
	for k, v := range iPs {
		// Check if current v (BruteIP) ban time has expired.
		if time.Now().Unix() > v.Timestamp+cfg.BanTime {
			delete(iPs, k)
		}
	}
}
