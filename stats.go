// mgo - MongoDB driver for Go

package mgo

import (
	"sync"
	"time"
)

var stats *Stats
var statsMutex sync.Mutex

// SetStats enable database state monitoring
func SetStats(enabled bool) {
	statsMutex.Lock()
	if enabled {
		if stats == nil {
			stats = &Stats{}
		}
	} else {
		stats = nil
	}
	statsMutex.Unlock()
}

// GetStats return the current database state
func GetStats() (snapshot Stats) {
	statsMutex.Lock()
	snapshot = *stats
	statsMutex.Unlock()
	return
}

// ResetStats reset Stats to the previous database state
func ResetStats() {
	statsMutex.Lock()
	debug("Resetting stats")
	old := stats
	stats = &Stats{}
	// These are absolute values:
	stats.Clusters = old.Clusters
	stats.SocketsInUse = old.SocketsInUse
	stats.SocketsAlive = old.SocketsAlive
	stats.SocketRefs = old.SocketRefs
	statsMutex.Unlock()
	return
}

// Stats holds info on the database state
//
// Relevant documentation:
//
//    https://docs.mongodb.com/manual/reference/command/serverStatus/
//
// TODO outdated fields ?
type Stats struct {
	Clusters            int
	MasterConns         int
	SlaveConns          int
	SentOps             int
	ReceivedOps         int
	ReceivedDocs        int
	SocketsAlive        int
	SocketsInUse        int
	SocketRefs          int
	TimesSocketAcquired int
	TimesWaitedForPool  int
	TotalPoolWaitTime   time.Duration
	PoolTimeouts        int
}

func (stats *Stats) cluster(delta int) {
	if stats != nil {
		statsMutex.Lock()
		stats.Clusters += delta
		statsMutex.Unlock()
	}
}

func (stats *Stats) conn(delta int, master bool) {
	if stats != nil {
		statsMutex.Lock()
		if master {
			stats.MasterConns += delta
		} else {
			stats.SlaveConns += delta
		}
		statsMutex.Unlock()
	}
}

func (stats *Stats) sentOps(delta int) {
	if stats != nil {
		statsMutex.Lock()
		stats.SentOps += delta
		statsMutex.Unlock()
	}
}

func (stats *Stats) receivedOps(delta int) {
	if stats != nil {
		statsMutex.Lock()
		stats.ReceivedOps += delta
		statsMutex.Unlock()
	}
}

func (stats *Stats) receivedDocs(delta int) {
	if stats != nil {
		statsMutex.Lock()
		stats.ReceivedDocs += delta
		statsMutex.Unlock()
	}
}

func (stats *Stats) socketsInUse(delta int) {
	if stats != nil {
		statsMutex.Lock()
		stats.SocketsInUse += delta
		statsMutex.Unlock()
	}
}

func (stats *Stats) socketsAlive(delta int) {
	if stats != nil {
		statsMutex.Lock()
		stats.SocketsAlive += delta
		statsMutex.Unlock()
	}
}

func (stats *Stats) socketRefs(delta int) {
	if stats != nil {
		statsMutex.Lock()
		stats.SocketRefs += delta
		statsMutex.Unlock()
	}
}

func (stats *Stats) noticeSocketAcquisition(waitTime time.Duration) {
	if stats != nil {
		statsMutex.Lock()
		stats.TimesSocketAcquired++
		stats.TotalPoolWaitTime += waitTime
		if waitTime > 0 {
			stats.TimesWaitedForPool++
		}
		statsMutex.Unlock()
	}
}

func (stats *Stats) noticePoolTimeout(waitTime time.Duration) {
	if stats != nil {
		statsMutex.Lock()
		stats.TimesWaitedForPool++
		stats.PoolTimeouts++
		stats.TotalPoolWaitTime += waitTime
		statsMutex.Unlock()
	}
}
