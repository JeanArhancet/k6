/*
 *
 * k6 - a next-generation load testing tool
 * Copyright (C) 2021 Load Impact
 *
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Affero General Public License as
 * published by the Free Software Foundation, either version 3 of the
 * License, or (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU Affero General Public License for more details.
 *
 * You should have received a copy of the GNU Affero General Public License
 * along with this program.  If not, see <http://www.gnu.org/licenses/>.
 *
 */

package js

import (
	"context"
	"sync"

	"github.com/dop251/goja"
)

// an event loop
// TODO: DO NOT USE AS IT'S NOT DONE
type eventLoop struct {
	queueLock     sync.Mutex
	queue         []func()
	wakeupCh      chan struct{} // maybe use sync.Cond ?
	reservedCount int
}

func newEventLoop() *eventLoop {
	return &eventLoop{
		wakeupCh: make(chan struct{}, 1),
	}
}

func (e *eventLoop) RunOnLoop(f func()) {
	e.queueLock.Lock()
	e.queue = append(e.queue, f)
	e.queueLock.Unlock()
	select {
	case e.wakeupCh <- struct{}{}:
	default:
	}
}

func (e *eventLoop) Reserve() func(func()) {
	e.queueLock.Lock()
	e.reservedCount++
	e.queueLock.Unlock()

	return func(f func()) {
		e.queueLock.Lock()
		e.queue = append(e.queue, f)
		e.reservedCount--
		e.queueLock.Unlock()
	}
}

func (e *eventLoop) Start(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
		}
		e.queueLock.Lock()
		queue := e.queue
		e.queue = make([]func(), 0, len(queue))
		e.queueLock.Unlock()
		for _, f := range queue {
			f()
		}
		e.queueLock.Lock()
		l := len(e.queue)
		reserved := e.reservedCount != 0
		e.queueLock.Unlock()
		if l == 0 {
			if !reserved {
				return
			}
			select {
			case <-ctx.Done():
				return
			case <-e.wakeupCh:
			}
		}
	}
}

func (e *eventLoop) makeHandledPromise(rt *goja.Runtime) (*goja.Promise, func(interface{}), func(interface{})) {
	reserved := e.Reserve()
	p, resolve, reject := rt.NewPromise()
	return p, func(i interface{}) {
			// more stuff
			reserved(func() { resolve(i) })
		}, func(i interface{}) {
			// more stuff
			reserved(func() { reject(i) })
		}
}
