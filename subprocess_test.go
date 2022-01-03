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

package nin

import (
	"os"
	"os/signal"
	"runtime"
	"testing"
)

func testCommand() string {
	if runtime.GOOS == "windows" {
		return "cmd /c dir \\"
	}
	return "ls /"
}

func NewSubprocessSetTest(t *testing.T) SubprocessSet {
	s := NewSubprocessSet()
	t.Cleanup(s.Clear)
	return s
}

// Run a command that fails and emits to stderr.
func TestSubprocessTest_BadCommandStderr(t *testing.T) {
	subprocs_ := NewSubprocessSetTest(t)
	cmd := "bash -c foo"
	if runtime.GOOS == "windows" {
		cmd = "cmd /c ninja_no_such_command"
	}
	subproc := subprocs_.Add(cmd, false)
	if nil == subproc {
		t.Fatal("expected different")
	}

	for !subproc.Done() {
		// Pretend we discovered that stderr was ready for writing.
		subprocs_.DoWork()
	}

	// ExitFailure
	want := 127
	if runtime.GOOS == "windows" {
		want = 1
	}
	if got := subproc.Finish(); got != want {
		t.Fatal(got)
	}
	if got := subproc.GetOutput(); got == "" {
		t.Fatal("expected error output")
	}
}

// Run a command that does not exist
func TestSubprocessTest_NoSuchCommand(t *testing.T) {
	subprocs_ := NewSubprocessSetTest(t)
	subproc := subprocs_.Add("ninja_no_such_command", false)
	if nil == subproc {
		t.Fatal("expected different")
	}

	for !subproc.Done() {
		// Pretend we discovered that stderr was ready for writing.
		subprocs_.DoWork()
	}

	// ExitFailure
	// 127 on posix, -1 on Windows.
	want := 127 // Generated by /bin/sh.
	if runtime.GOOS == "windows" {
		want = -1
	}
	if got := subproc.Finish(); got != want {
		t.Fatal(got)
	}
	/*
		if got := subproc.GetOutput(); got != "" {
			t.Fatalf("%q", got)
		}
		if runtime.GOOS == "windows" {
			if "CreateProcess failed: The system cannot find the file specified.\n" != subproc.GetOutput() {
				t.Fatal()
			}
		}
	*/
}

func TestSubprocessTest_InterruptChild(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("can't run on Windows")
	}
	subprocs_ := NewSubprocessSetTest(t)
	subproc := subprocs_.Add("kill -INT $$", false)
	if nil == subproc {
		t.Fatal("expected different")
	}

	for !subproc.Done() {
		subprocs_.DoWork()
	}

	// ExitInterrupted
	if got := subproc.Finish(); got != -1 {
		t.Fatal(got)
	}
}

func TestSubprocessTest_InterruptParent(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("can't run on Windows")
	}
	t.Skip("TODO")
	subprocs_ := NewSubprocessSetTest(t)
	c := make(chan os.Signal, 1)
	go func() {
		<-c
		subprocs_.Clear()
	}()
	signal.Notify(c, os.Interrupt)
	defer signal.Reset(os.Interrupt)
	subproc := subprocs_.Add("kill -INT $PPID ; sleep 1", false)
	if nil == subproc {
		t.Fatal("expected different")
	}

	for !subproc.Done() {
		if subprocs_.DoWork() {
			return
		}
	}

	t.Fatal("We should have been interrupted")
}

func TestSubprocessTest_InterruptChildWithSigTerm(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("can't run on Windows")
	}
	subprocs_ := NewSubprocessSetTest(t)
	subproc := subprocs_.Add("kill -TERM $$", false)
	if nil == subproc {
		t.Fatal("expected different")
	}

	for !subproc.Done() {
		subprocs_.DoWork()
	}

	// TODO(maruel): ExitInterrupted
	if got := subproc.Finish(); got != -1 {
		t.Fatal(got)
	}
}

func TestSubprocessTest_InterruptParentWithSigTerm(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("can't run on Windows")
	}
	t.Skip("TODO")
	subprocs_ := NewSubprocessSetTest(t)
	subproc := subprocs_.Add("kill -TERM $PPID ; sleep 1", false)
	if nil == subproc {
		t.Fatal("expected different")
	}

	for !subproc.Done() {
		if subprocs_.DoWork() {
			return
		}
	}

	t.Fatal("We should have been interrupted")
}

func TestSubprocessTest_InterruptChildWithSigHup(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("can't run on Windows")
	}
	subprocs_ := NewSubprocessSetTest(t)
	subproc := subprocs_.Add("kill -HUP $$", false)
	if nil == subproc {
		t.Fatal("expected different")
	}

	for !subproc.Done() {
		subprocs_.DoWork()
	}

	// TODO(maruel): ExitInterrupted
	if got := subproc.Finish(); got != -1 {
		t.Fatal(got)
	}
}

func TestSubprocessTest_InterruptParentWithSigHup(t *testing.T) {
	t.Skip("TODO")
	if runtime.GOOS == "windows" {
		t.Skip("can't run on Windows")
	}
	subprocs_ := NewSubprocessSetTest(t)
	subproc := subprocs_.Add("kill -HUP $PPID ; sleep 1", false)
	if nil == subproc {
		t.Fatal("expected different")
	}

	for !subproc.Done() {
		if subprocs_.DoWork() {
			return
		}
	}

	t.Fatal("We should have been interrupted")
}

func TestSubprocessTest_Console(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("can't run on Windows")
	}
	t.Skip("TODO")
	/*
		// Skip test if we don't have the console ourselves.
		// TODO(maruel): Sub-run with a fake pty?
		if !isatty(0) || !isatty(1) || !isatty(2) {
			t.Skip("need a real console to run this test")
		}
	*/
	subprocs_ := NewSubprocessSetTest(t)
	// useConsole = true
	subproc := subprocs_.Add("test -t 0 -a -t 1 -a -t 2", true)
	if nil == subproc {
		t.Fatal("expected different")
	}

	for !subproc.Done() {
		subprocs_.DoWork()
	}

	if got := subproc.Finish(); got != ExitSuccess {
		t.Fatal(got)
	}
}

func TestSubprocessTest_SetWithSingle(t *testing.T) {
	subprocs_ := NewSubprocessSetTest(t)
	subproc := subprocs_.Add(testCommand(), false)
	if subproc == nil {
		t.Fatal("expected different")
	}

	for !subproc.Done() {
		subprocs_.DoWork()
	}
	if subproc.Finish() != ExitSuccess {
		t.Fatal("expected equal")
	}
	if subproc.GetOutput() == "" {
		t.Fatal("expected different")
	}

	if got := subprocs_.Finished(); got != 1 {
		t.Fatal(got)
	}
}

func TestSubprocessTest_SetWithMulti(t *testing.T) {
	processes := [3]Subprocess{}
	commands := []string{testCommand()}
	if runtime.GOOS == "windows" {
		commands = append(commands, "cmd /c echo hi", "cmd /c time /t")
	} else {
		commands = append(commands, "id -u", "pwd")
	}

	subprocs_ := NewSubprocessSetTest(t)
	for i := 0; i < 3; i++ {
		processes[i] = subprocs_.Add(commands[i], false)
		if processes[i] == nil {
			t.Fatal("expected different")
		}
	}

	if subprocs_.Running() != 3 {
		t.Fatal("expected equal")
	}
	/* The expectations with the C++ code is different.
	for i := 0; i < 3; i++ {
		if processes[i].Done() {
			t.Fatal("expected false")
		}
		if got := processes[i].GetOutput(); got != "" {
			t.Fatalf("%q", got)
		}
	}
	*/

	for !processes[0].Done() || !processes[1].Done() || !processes[2].Done() {
		if subprocs_.Running() <= 0 {
			t.Fatal("expected greater")
		}
		subprocs_.DoWork()
	}

	if subprocs_.Running() != 0 {
		t.Fatal("expected equal")
	}
	if subprocs_.Finished() != 3 {
		t.Fatal("expected equal")
	}

	for i := 0; i < 3; i++ {
		if processes[i].Finish() != ExitSuccess {
			t.Fatal("expected equal")
		}
		if processes[i].GetOutput() == "" {
			t.Fatal("expected different")
		}
	}
}

func TestSubprocessTest_SetWithLots(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("skipped on windows")
	}

	// Arbitrary big number; needs to be over 1024 to confirm we're no longer
	// hostage to pselect.
	const numProcs = 1025

	subprocessTestFixUlimit(t, numProcs)
	cmd := "/bin/echo"

	subprocs_ := NewSubprocessSetTest(t)
	var procs []Subprocess
	for i := 0; i < numProcs; i++ {
		subproc := subprocs_.Add(cmd, false)
		if nil == subproc {
			t.Fatal("expected different")
		}
		procs = append(procs, subproc)
	}
	for subprocs_.Running() != 0 {
		subprocs_.DoWork()
	}
	for i := 0; i < len(procs); i++ {
		if got := procs[i].Finish(); got != ExitSuccess {
			t.Fatal(got)
		}
		if procs[i].GetOutput() == "" {
			t.Fatal("expected different")
		}
	}
	if numProcs != subprocs_.Finished() {
		t.Fatal("expected equal")
	}
}

// TODO: this test could work on Windows, just not sure how to simply
// read stdin.
// Verify that a command that attempts to read stdin correctly thinks
// that stdin is closed.
func TestSubprocessTest_ReadStdin(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Has to be ported")
	}
	subprocs_ := NewSubprocessSetTest(t)
	subproc := subprocs_.Add("cat -", false)
	for !subproc.Done() {
		subprocs_.DoWork()
	}
	if subproc.Finish() != ExitSuccess {
		t.Fatal("expected equal")
	}
	if subprocs_.Finished() != 1 {
		t.Fatal("expected equal")
	}
}
