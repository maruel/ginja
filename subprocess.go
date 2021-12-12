// Copyright 2012 Google Inc. All Rights Reserved.
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

package ginga

// Subprocess wraps a single async subprocess.  It is entirely
// passive: it expects the caller to notify it when its fds are ready
// for reading, as well as call Finish() to reap the child once done()
// is true.
type Subprocess struct {
	buf_            string
	child_          HANDLE
	pipe_           HANDLE
	overlapped_     OVERLAPPED
	overlapped_buf_ [4 << 10]byte
	is_reading_     bool
	fd_             int
	pid_            uint
	use_console_    bool
}

// SubprocessSet runs a ppoll/pselect() loop around a set of Subprocesses.
// DoWork() waits for any state change in subprocesses; finished_
// is a queue of subprocesses as they finish.
type SubprocessSet struct {
	running_  []*Subprocess
	finished_ []*Subprocess // queue<Subprocess*>

	//static HANDLE ioport_
	// Store the signal number that causes the interruption.
	// 0 if not interruption.
	//static int interrupted_

	//static bool IsInterrupted() { return interrupted_ != 0; }
	/*
	  struct sigaction old_int_act_
	  struct sigaction old_term_act_
	  struct sigaction old_hup_act_
	  sigset_t old_mask_
	*/
}
