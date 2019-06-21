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
	"io"
	"testing"

	"gotest.tools/assert"
)

func TestMockFile(t *testing.T) {

	type test struct {
		name    string
		r, s, c bool
	}

	testfunc := func(test *test) func(*testing.T) {
		return func(t *testing.T) {
			mf := MockFile{}

			in_p := []byte{0x01, 0x02}
			in_offset := int64(65)
			in_whence := 23

			if test.r {
				rc := make(chan struct{}, 0)

				mf.R = RF(
					func(p []byte) (n int, err error) {
						assert.DeepEqual(t, p, in_p)
						close(rc)
						return 123, anError
					},
				)

				n, err := mf.Read(in_p)

				select {
				case <-rc:
				default:
					t.Fatal("Channel not closed")
				}

				assert.Equal(t, n, 123)
				assert.Assert(t, err == anError)

			} else {
				n, err := mf.Read(in_p)
				assert.Equal(t, n, 0)
				assert.Assert(t, err == io.EOF)
			}

			if test.s {
				sc := make(chan struct{}, 0)
				mf.S = SF(
					func(offset int64, whence int) (int64, error) {
						assert.Equal(t, offset, in_offset)
						assert.Equal(t, whence, in_whence)
						close(sc)
						return 456, anError
					},
				)

				o, err := mf.Seek(in_offset, in_whence)

				select {
				case <-sc:
				default:
					t.Fatal("Channel not closed")
				}

				assert.Equal(t, o, int64(456))
				assert.Assert(t, err == anError)

			} else {
				o, err := mf.Seek(123, 0)

				assert.Equal(t, o, int64(123))
				assert.NilError(t, err)
			}

			if test.c {
				cc := make(chan struct{}, 0)

				mf.C = CF(
					func() error {
						close(cc)
						return anError
					},
				)

				err := mf.Close()

				select {
				case <-cc:
				default:
					t.Fatal("Channel not closed")
				}

				assert.Assert(t, err == anError)
			} else {
				assert.NilError(t, mf.Close())
			}
		}
	}

	tests := []test{
		test{"nil", false, false, false},
		test{"funcs", true, true, true},
		test{"mix1", true, false, true},
		test{"mix2", false, true, false},
	}

	for _, test := range tests {
		t.Run(test.name, testfunc(&test))
	}
}
