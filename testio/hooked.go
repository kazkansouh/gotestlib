/*
 * Copyright (c) 2019 Karim Kanso. All Rights Reserved.
 *
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with this program.  If not, see <http://www.gnu.org/licenses/>.
 */

// Utility functions useful for orchestrating io testing.
package testio

import (
	"io"
)

// Action to be called by hooked reader.
//
// The parameter passed to the function is a counter used to
// differentiate each time the callback is used within a given hooked
// reader, it starts at 1. This allows for different actions to be
// performed on each call.
//
// The return "next" indicates how many more bytes to read before
// executing the function again. A negative disables this.
//
// The return "err" is passed directly to the result of Read. It is
// typically used to inject errors.
//
//   For example, to create a one-shot EOF error it could return:
//     (-1, io.EOF)
//   Or a recurring EOF every time 20 bytes are consume, it could
//   return:
//     (20, io.EOF)
type HookedF func(ctr int) (next int, err error)

// Implements io.Reader interface. Can be used to hook into a reader
// to perform an action after reading n bytes, such as returning an
// error or sending sending data over a channel.
type hookedR struct {
	// Underlying reader
	r io.Reader
	// bytes until action to execute, negative disable f
	n int
	// action to execute, return value is reset value
	f func(ctr int) (int, error)

	// do no set, incremented before each call to f
	ctr int
}

// read n bytes then perform action f
func (r *hookedR) Read(p []byte) (int, error) {
	if r.n < 0 {
		return r.r.Read(p)
	}
	if len(p) < r.n {
		n, err := r.r.Read(p)
		r.n -= n
		return n, err
	}
	n, err := r.r.Read(p[:r.n])
	r.n -= n
	if err != nil {
		return n, err
	}
	r.ctr++

	r.n, err = r.f(r.ctr)
	return n, err
}

// A Hooked Reader is one that allows for hooking a callback function
// into the reader which is called after reading a number of
// bytes. This is useful for interacting with a program performing IO,
// e.g. injecting an error into the result of calling Read or sending
// an action over a channel to interact with the routine calling Read
// (i.e. a graceful shutdown).
//
// Create a new reader that will execute function f after reading n
// bytes from r.
//
// See description of HookedF for more information.
func NewHookedReader(r io.Reader, n int, f HookedF) io.Reader {
	return &hookedR{
		r:   r,
		n:   n,
		f:   f,
		ctr: 0,
	}
}
