// MFP - Miulti-Function Printers and scanners toolkit
// WSD device discovery
//
// Copyright (C) 2024 and up by Alexander Pevzner (pzz@apevzner.com)
// See LICENSE for license terms and conditions
//
// Multicast messaging scheduler

package wsdd

import (
	"sync"
	"time"

	"github.com/alexpevzner/mfp/internal/random"
)

// Scheduler parameters:
//
// The following diagram will help to understand scheduler parameters:
//
//	 -- random pause, probeRetransmitDelayMin...probeRetransmitDelayMax
//	 |
//	 |      ------------------- probeFastSeriesDelay
//	 |      |             ----- probeInterSeriesDelay
//	 |      |             |
//	 V      V             V
//	1-1-1-1---2-2-2-2----------4-4-4-4---5-5-5-5----
//	<----->   <----->
//	   |
//	   `--------------- probeRetransmitSeriesLen
//
//	   |<------->|
//	        |
//	        ----------- probeFastSeriesLen (count of retretransmit
//	                                        series)
//
// Scheduler may run in either "browse" or "resolve" mode.
//
// In the "browse" mode it continuously sends messages, grouped by series,
// as explained below:
//
//  1. The same message is repeated (retransmitted) several times with
//     randomized pauses between retransmissions, to compensate possible
//     packet lost,  which becomes especially serious problem when
//     multicasting over WiFi. This is called "retransmit series".
//  2. Some retransmit series are repeated with small intervals between
//     them. This is called "fast series".
//  3. The fast series are continuously repeated with some delay
//     between them.
//
// The "resolve" mode has the following differences:
//  1. The fast series has unlimited length
//  2. The entire process has a timeout.
const (
	// The retransmit series parameters:
	schedRetransmitSeriesLen = 4
	schedRetransmitDelayMin  = 250 * time.Millisecond
	schedRetransmitDelayMax  = 500 * time.Millisecond

	// The fast series parameters:
	schedFastSeriesLen   = 2
	schedFastSeriesDelay = 1000 * time.Millisecond

	// Delay between continuously repeated fast series:
	schedInterSeriesDelay = 5000 * time.Millisecond

	// Resolve mode parameters:
	schedResolveMaxTime = 5000 * time.Millisecond
)

// sched is the multicast messaging scheduler.
//
// The scheduler can be used either for continuously sending
// probes ("browsing") or to find some particular peer ("resolving").
type sched struct {
	resolve bool            // Resolve mode
	timer   timer           // Underlying timer
	c       chan schedEvent // Event channel
	done    sync.WaitGroup  // For sched.Close synchronization
}

// schedEvent are events, generated by the scheduler
type schedEvent int

const (
	schedClosed     schedEvent = iota // Scheduler is closed
	schedNewMessage                   // Generate new message
	schedSend                         // Send current message
)

// newSched creates a new scheduler
func newSched(resolve bool) *sched {
	s := &sched{
		resolve: resolve,
		timer:   newTimer(),
		c:       make(chan schedEvent, 4),
	}

	s.done.Add(1)
	go s.proc()

	return s
}

// Close closes the scheduler.
func (s *sched) Close() {
	s.timer.Cancel()
	for len(s.c) > 0 {
		<-s.c
	}

	s.done.Wait()
}

// Chan returns scheduler's event channel.
//
// Scheduler closes this channel when scheduler is closed.
// Additionally, resolve-mode scheduler closes this channel,
// when resolve timeout is reached.
//
// When scheduler channel is closed, any attempt to read from
// it will return schedClosed, which is the zero value for the
// schedEvent.
func (s *sched) Chan() <-chan schedEvent {
	return s.c
}

// proc runs on its own goroutine and generates events
func (s *sched) proc() {
	defer s.done.Done()
	defer close(s.c)

	start := time.Now()

	for {
		// Run fast series
		for fastCnt := 0; fastCnt < schedFastSeriesLen; {
			s.c <- schedNewMessage

			// Run retransmit series
			for tx := 0; tx < schedRetransmitSeriesLen; tx++ {
				s.c <- schedSend

				delay := time.Duration(random.UintRange(
					uint(schedRetransmitDelayMin),
					uint(schedRetransmitDelayMax)))

				if !s.timer.Sleep(delay) {
					return
				}
			}

			if !s.timer.Sleep(schedFastSeriesDelay) {
				return
			}

			if !s.resolve {
				fastCnt++
			}

			// Check for resolve timeout.
			if s.resolveTimedOut(start) {
				return
			}
		}

		if !s.timer.Sleep(schedInterSeriesDelay) {
			return
		}
	}
}

// resolveTimedOut returns true, if scheduler runs in resolve
// mode and resolve max time reached.
func (s *sched) resolveTimedOut(start time.Time) bool {
	if !s.resolve {
		return false
	}

	now := time.Now()
	return now.Sub(start) >= schedResolveMaxTime
}
