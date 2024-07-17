// MFP - Miulti-Function Printers and scanners toolkit
// Network state monitoring
//
// Copyright (C) 2024 and up by Alexander Pevzner (pzz@apevzner.com)
// See LICENSE for license terms and conditions
//
// Event queue

package netstate

// eventqueue is the queue of Event-s
type eventqueue struct {
	events []Event
}

// push adds Events to the queue.
func (eq *eventqueue) push(events ...Event) {
	eq.events = append(eq.events, events...)
}

// pull returns first Event from the queue or nil, if queue is empty.
func (eq *eventqueue) pull() (evnt Event) {
	if len(eq.events) > 0 {
		evnt = eq.events[0]
		copy(eq.events, eq.events[1:])
		eq.events = eq.events[:len(eq.events)-1]
	}

	return
}
