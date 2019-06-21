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

package testio

import (
	"errors"
	"io"
	"strings"
	"testing"

	"gotest.tools/assert"
)

var anError = errors.New("An Error")

func TestHookedReader(t *testing.T) {
	type iter struct {
		read  int
		data  []byte
		error string
	}

	type test struct {
		name     string
		r        io.Reader
		n        int
		f        HookedF
		expected []iter
	}

	testfunc := func(test *test) func(*testing.T) {
		return func(t *testing.T) {
			c := make(chan int, 1)
			f := func(ctr int) (next int, err error) {
				c <- ctr
				return test.f(ctr)
			}

			r := NewHookedReader(test.r, test.n, f)
			assert.Assert(t, r != nil)

			ctr := 1

			for _, x := range test.expected {
				t.Run(test.name, func(t *testing.T) {
					remain := r.(*hookedR).n

					// read data
					buffer := make([]byte, x.read)
					n, err := r.Read(buffer)

					// if read data past callback, check
					// callback was called
					if remain >= 0 && remain <= n {
						select {
						case y := <-c:
							assert.Equal(t, ctr, y)
							ctr++
						default:
							t.Fatal("Callback was not called")

						}
					}

					// acceptance
					if x.error == "" {
						assert.NilError(t, err)
					} else {
						assert.ErrorContains(t, err, x.error)
					}
					assert.Equal(t, n, len(x.data))
					assert.DeepEqual(t, buffer[:n], x.data)
				})
			}
		}
	}

	tests := []test{
		test{
			name: "no-callback",
			r:    strings.NewReader("hello world"),
			n:    -1,
			f:    nil,
			expected: []iter{
				iter{
					read:  5,
					data:  []byte("hello"),
					error: "",
				},
				iter{
					read:  20,
					data:  []byte(" world"),
					error: "",
				},
				iter{
					read:  20,
					data:  []byte{},
					error: io.EOF.Error(),
				},
			},
		},
		test{
			name: "distant",
			r:    strings.NewReader("hello world"),
			n:    100,
			f: func(ctr int) (int, error) {
				return -1, anError
			},
			expected: []iter{
				iter{
					read:  20,
					data:  []byte("hello world"),
					error: "",
				},
				iter{
					read:  20,
					data:  []byte{},
					error: io.EOF.Error(),
				},
			},
		},
		test{
			name: "one-shot",
			r:    strings.NewReader("hello world"),
			n:    5,
			f: func(ctr int) (int, error) {
				return -1, anError
			},
			expected: []iter{
				iter{
					read:  10,
					data:  []byte("hello"),
					error: anError.Error(),
				},
				iter{
					read:  20,
					data:  []byte(" world"),
					error: "",
				},
				iter{
					read:  20,
					data:  []byte{},
					error: io.EOF.Error(),
				},
			},
		},
		test{
			name: "one-shot-at-eof",
			r:    strings.NewReader("hello world"),
			n:    11,
			f: func(ctr int) (int, error) {
				return -1, anError
			},
			expected: []iter{
				iter{
					read:  11,
					data:  []byte("hello world"),
					error: anError.Error(),
				},
				iter{
					read:  20,
					data:  []byte{},
					error: io.EOF.Error(),
				},
			},
		},
		test{
			name: "recurring-noerror",
			r:    strings.NewReader("hello world"),
			n:    5,
			f: func(ctr int) (int, error) {
				return 5, nil
			},
			expected: []iter{
				iter{
					read:  10,
					data:  []byte("hello"),
					error: "",
				},
				iter{
					read:  20,
					data:  []byte(" worl"),
					error: "",
				},
				iter{
					read:  20,
					data:  []byte("d"),
					error: "",
				},
				iter{
					read:  20,
					data:  []byte{},
					error: io.EOF.Error(),
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, testfunc(&test))
	}
}
