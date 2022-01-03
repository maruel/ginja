// Copyright 2011 Google Inc. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package nin

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"
)

// Interface for reading files from disk.  See DiskInterface for details.
// This base offers the minimum interface needed just to read files.
type FileReader interface {
	// Read and store in given string.  On success, return Okay and injects a
	// trailing 0 byte.
	// On error, return another Status and fill |err|.
	ReadFile(path string, contents *string, err *string) DiskStatus
}

// Result of ReadFile.
type DiskStatus int32

const (
	Okay DiskStatus = iota
	NotFound
	OtherError
)

// Interface for accessing the disk.
//
// Abstract so it can be mocked out for tests.  The real implementation
// is RealDiskInterface.
type DiskInterface interface {
	FileReader
	// stat() a file, returning the mtime, or 0 if missing and -1 on
	// other errors.
	Stat(path string, err *string) TimeStamp

	// Create a directory, returning false on failure.
	MakeDir(path string) bool

	// Create a file, with the specified name and contents
	// Returns true on success, false on failure
	WriteFile(path, contents string) bool

	// Remove the file named @a path. It behaves like 'rm -f path' so no errors
	// are reported if it does not exists.
	// @returns 0 if the file has been removed,
	//          1 if the file does not exist, and
	//          -1 if an error occurs.
	RemoveFile(path string) int
}

type DirCache map[string]TimeStamp
type Cache map[string]DirCache

func DirName(path string) string {
	return filepath.Dir(path)
	/*
		kPathSeparators := "\\/"
		kEnd := kPathSeparators + len(kPathSeparators) - 1

		slashPos := path.findLastOf(kPathSeparators)
		if slashPos == -1 {
			return "" // Nothing to do.
		}
		for slashPos > 0 && find(kPathSeparators, kEnd, path[slashPos-1]) != kEnd {
			slashPos--
		}
		return path[0:slashPos]
	*/
}

func MakeDir(path string) int {
	//return Mkdir(path)
	if err := os.Mkdir(path, 0o777); err != nil {
		return 1
	}
	return 0
}

/*
func TimeStampFromFileTime(filetime *FILETIME) TimeStamp {
	// FILETIME is in 100-nanosecond increments since the Windows epoch.
	// We don't much care about epoch correctness but we do want the
	// resulting value to fit in a 64-bit integer.
	mtime := (filetime.dwHighDateTime << 32) | filetime.dwLowDateTime
	// 1600 epoch -> 2000 epoch (subtract 400 years).
	return mtime - 12622770400*(1000000000/100)
}
*/

func StatSingleFile(path string, err *string) TimeStamp {
	/*
		var attrs WIN32_FILE_ATTRIBUTE_DATA
		if !GetFileAttributesExA(path, GetFileExInfoStandard, &attrs) {
			winErr := GetLastError()
			if winErr == ERROR_FILE_NOT_FOUND || winErr == ERROR_PATH_NOT_FOUND {
				return 0
			}
			*err = "GetFileAttributesEx(" + path + "): " + GetLastErrorString()
			return -1
		}
		return TimeStampFromFileTime(attrs.ftLastWriteTime)
	*/

	// This will obviously have to be optimized.
	s, err2 := os.Stat(path)
	if err2 != nil {
		// See TestDiskInterfaceTest_StatMissingFile for rationale for ENOTDIR
		// check.
		if os.IsNotExist(err2) || errors.Unwrap(err2) == syscall.ENOTDIR {
			return 0
		}
		*err = err2.Error()
		return -1
	}
	return TimeStamp(s.ModTime().UnixMicro())
}

/*
func IsWindows7OrLater() bool {
	versionInfo := OSVERSIONINFOEX{sizeof(OSVERSIONINFOEX), 6, 1, 0, 0, {0}, 0, 0, 0, 0, 0}
	comparison := 0
	VER_SET_CONDITION(comparison, VER_MAJORVERSION, VER_GREATER_EQUAL)
	VER_SET_CONDITION(comparison, VER_MINORVERSION, VER_GREATER_EQUAL)
	return VerifyVersionInfo(&versionInfo, VER_MAJORVERSION|VER_MINORVERSION, comparison)
}
*/

func StatAllFilesInDir(dir string, stamps map[string]TimeStamp, err *string) bool {
	/*
		// FindExInfoBasic is 30% faster than FindExInfoStandard.
		//canUseBasicInfo := IsWindows7OrLater()
		// This is not in earlier SDKs.
		//FINDEX_INFO_LEVELS
		kFindExInfoBasic := 1
		//FINDEX_INFO_LEVELS
		level := kFindExInfoBasic
		// FindExInfoStandard
		var ffd WIN32_FIND_DATAA
		findHandle := FindFirstFileExA((dir + "\\*"), level, &ffd, FindExSearchNameMatch, nil, 0)

		if findHandle == INVALID_HANDLE_VALUE {
			winErr := GetLastError()
			if winErr == ERROR_FILE_NOT_FOUND || winErr == ERROR_PATH_NOT_FOUND {
				return true
			}
			*err = "FindFirstFileExA(" + dir + "): " + GetLastErrorString()
			return false
		}
		for {
			lowername := ffd.cFileName
			if lowername == ".." {
				// Seems to just copy the timestamp for ".." from ".", which is wrong.
				// This is the case at least on NTFS under Windows 7.
				continue
			}
			lowername = strings.ToLower(lowername)
			stamps[lowername] = TimeStampFromFileTime(ffd.ftLastWriteTime)
			if !FindNextFileA(findHandle, &ffd) {
				break
			}
		}
		FindClose(findHandle)
		return true
	*/
	f, err2 := os.Open(dir)
	if err2 != nil {
		*err = err2.Error()
		return false
	}
	d, err2 := f.Readdir(0)
	if err2 != nil {
		*err = err2.Error()
		_ = f.Close()
		return false
	}
	for _, i := range d {
		if !i.IsDir() {
			stamps[i.Name()] = TimeStamp(i.ModTime().UnixMicro())
		}
	}
	_ = f.Close()
	return true
}

// Create all the parent directories for path; like mkdir -p
// `basename path`.
func MakeDirs(d DiskInterface, path string) bool {
	dir := DirName(path)
	if dir == path || dir == "." || dir == "" {
		return true // Reached root; assume it's there.
	}
	err := ""
	mtime := d.Stat(dir, &err)
	if mtime < 0 {
		errorf("%s", err)
		return false
	}
	if mtime > 0 {
		return true // Exists already; we're done.
	}

	// Directory doesn't exist.  Try creating its parent first.
	if !MakeDirs(d, dir) {
		return false
	}
	return d.MakeDir(dir)
}

//

// Implementation of DiskInterface that actually hits the disk.
type RealDiskInterface struct {
	// Whether stat information can be cached.
	useCache_ bool

	// TODO: Neither a map nor a hashmap seems ideal here.  If the statcache
	// works out, come up with a better data structure.
	cache_ Cache
}

func NewRealDiskInterface() RealDiskInterface {
	return RealDiskInterface{}
}

// MSDN: "Naming Files, Paths, and Namespaces"
// http://msdn.microsoft.com/en-us/library/windows/desktop/aa365247(v=vs.85).aspx
const maxPath = 260

func (r *RealDiskInterface) Stat(path string, err *string) TimeStamp {
	defer MetricRecord("node stat")()
	if runtime.GOOS == "windows" {
		if path != "" && path[0] != '\\' && len(path) >= maxPath {
			*err = fmt.Sprintf("Stat(%s): Filename longer than %d characters", path, maxPath)
			return -1
		}
		if !r.useCache_ {
			return StatSingleFile(path, err)
		}

		dir := DirName(path)
		o := 0
		if dir != "" {
			o = len(dir) + 1
		}
		base := path[o:]
		if base == ".." {
			// StatAllFilesInDir does not report any information for base = "..".
			base = "."
			dir = path
		}

		dir = strings.ToLower(dir)
		base = strings.ToLower(base)

		ci, ok := r.cache_[dir]
		if !ok {
			ci = DirCache{}
			r.cache_[dir] = ci
			s := "."
			if dir != "" {
				s = dir
			}
			if !StatAllFilesInDir(s, ci, err) {
				delete(r.cache_, dir)
				return -1
			}
		}
		return ci[base]
	}
	return StatSingleFile(path, err)
}

func (r *RealDiskInterface) WriteFile(path string, contents string) bool {
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o666)
	if err != nil {
		return false
	}
	_, err = f.WriteString(contents)
	if err1 := f.Close(); err == nil {
		err = err1
	}
	return err == nil
}

func (r *RealDiskInterface) MakeDir(path string) bool {
	/*
		if MakeDir(path) < 0 {
			if errno == EEXIST {
				return true
			}
			Error("mkdir(%s): %s", path, strerror(errno))
			return false
		}
		return true
	*/
	err := os.Mkdir(path, 0o777)
	return err == nil || os.IsExist(err)
}

func (r *RealDiskInterface) ReadFile(path string, contents *string, err *string) DiskStatus {
	c, err2 := ioutil.ReadFile(path)
	if err2 == nil {
		if len(c) == 0 {
			*contents = ""
		} else {
			// ioutil.ReadFile() is guaranteed to have an extra byte in the slice,
			// (ab)use it.
			*contents = string(c[:len(c)+1])
		}
		return Okay
	}
	*err = err2.Error()
	if os.IsNotExist(err2) {
		return NotFound
	}
	return OtherError
}

func (r *RealDiskInterface) RemoveFile(path string) int {
	if err := os.Remove(path); err != nil {
		// TODO: return -1?
		return 1
	}
	return 0
	/*
		attributes := GetFileAttributes(path)
		if attributes == INVALID_FILE_ATTRIBUTES {
			winErr := GetLastError()
			if winErr == ERROR_FILE_NOT_FOUND || winErr == ERROR_PATH_NOT_FOUND {
				return 1
			}
		} else if (attributes & FILE_ATTRIBUTE_READONLY) != 0 {
			// On non-Windows systems, remove() will happily delete read-only files.
			// On Windows Ninja should behave the same:
			//   https://github.com/ninja-build/ninja/issues/1886
			// Skip error checking.  If this fails, accept whatever happens below.
			SetFileAttributes(path, attributes & ^FILE_ATTRIBUTE_READONLY)
		}
		if (attributes & FILE_ATTRIBUTE_DIRECTORY) != 0 {
			// remove() deletes both files and directories. On Windows we have to
			// select the correct function (DeleteFile will yield Permission Denied when
			// used on a directory)
			// This fixes the behavior of ninja -t clean in some cases
			// https://github.com/ninja-build/ninja/issues/828
			if !RemoveDirectory(path) {
				winErr := GetLastError()
				if winErr == ERROR_FILE_NOT_FOUND || winErr == ERROR_PATH_NOT_FOUND {
					return 1
				}
				// Report remove(), not RemoveDirectory(), for cross-platform consistency.
				Error("remove(%s): %s", path, GetLastErrorString())
				return -1
			}
		} else {
			if !DeleteFile(path) {
				winErr := GetLastError()
				if winErr == ERROR_FILE_NOT_FOUND || winErr == ERROR_PATH_NOT_FOUND {
					return 1
				}
				// Report as remove(), not DeleteFile(), for cross-platform consistency.
				Error("remove(%s): %s", path, GetLastErrorString())
				return -1
			}
		}
		return 0
	*/
}

// Whether stat information can be cached.  Only has an effect on Windows.
func (r *RealDiskInterface) AllowStatCache(allow bool) {
	if runtime.GOOS == "windows" {
		r.useCache_ = allow
		if !r.useCache_ {
			r.cache_ = nil
		} else if r.cache_ == nil {
			r.cache_ = Cache{}
		}
	}
}
