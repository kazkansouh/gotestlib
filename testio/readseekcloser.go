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
	"os"
	"time"
)

// Wraps up a function as a io.Reader
type RF func(p []byte) (int, error)

func (f RF) Read(p []byte) (int, error) {
	return f(p)
}

// Wraps up a function as a io.Seeker
type SF func(offset int64, whence int) (int64, error)

func (f SF) Seek(offset int64, whence int) (int64, error) {
	return f(offset, whence)
}

// Wraps up a function as a io.Closer
type CF func() error

func (f CF) Close() error {
	return f()
}

// interface of the os.File.Stat function, used by MockFile
type Stater interface {
	Stat() (os.FileInfo, error)
}

// Wraps up a function as a Stater
type StatF func() (os.FileInfo, error)

func (f StatF) Stat() (os.FileInfo, error) {
	return f()
}

// Implements the ReadSeekCloser interface. It is a collection of a
// reader, seeker and closer that can be implemented by providing
// objects or functions.
//
// When a member is nil, no operation is performed when its action is
// performed. (e.g. in the case of R == nil, Read will only return an
// EOF).
//
// Use the RF, SF, CF types to wrap a single function as an object.
type MockFile struct {
	// The underlying reader
	R io.Reader
	// The underlying seeker
	S io.Seeker
	// The underlying closer
	C io.Closer
	// The underlying stat function
	St Stater
}

// When R is not nil, R.Read is used.
// Otherwise (0, io.EOF) is returned
func (f *MockFile) Read(p []byte) (n int, err error) {
	if f.R != nil {
		return f.R.Read(p)
	}
	return 0, io.EOF
}

// When C is not nil, C.Close is used.
// Otherwise nil is returned
func (f *MockFile) Close() error {
	if f.C != nil {
		return f.C.Close()
	}
	return nil
}

// When S is not nil, S.Seek is used.
// Otherwise the requested offset is returned (assumes that whence is 0)
func (f *MockFile) Seek(offset int64, whence int) (int64, error) {
	if f.S != nil {
		return f.S.Seek(offset, whence)
	}
	return offset, nil
}

// When St is not nil, St.Stat is used.  Otherwise a dummy
// MockFileInfo is returned with length 0.
func (f *MockFile) Stat() (os.FileInfo, error) {
	if f.St != nil {
		return f.St.Stat()
	}
	return &MockFileInfo{FileName: "dummyfile", FileMode: 0644}, nil
}

// Mock type that trivially implements the os.FileInfo interface
type MockFileInfo struct {
	FileName    string      // base name of the file
	FileSize    int64       // length in bytes for regular files; system-dependent for others
	FileMode    os.FileMode // file mode bits
	FileModTime time.Time   // modification time
	FileIsDir   bool        // true if file is directory
}

func (mfi *MockFileInfo) Name() string       { return mfi.FileName }
func (mfi *MockFileInfo) Size() int64        { return mfi.FileSize }
func (mfi *MockFileInfo) Mode() os.FileMode  { return mfi.FileMode }
func (mfi *MockFileInfo) ModTime() time.Time { return mfi.FileModTime }
func (mfi *MockFileInfo) IsDir() bool        { return mfi.FileIsDir }
func (mfi *MockFileInfo) Sys() interface{}   { return nil }
