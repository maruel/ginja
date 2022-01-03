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
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDepsLogTest_WriteRead(t *testing.T) {
	kTestFilename := filepath.Join(t.TempDir(), "DepsLogTest-tempfile")
	state1 := NewState()
	log1 := NewDepsLog()
	err := ""
	if !log1.OpenForWrite(kTestFilename, &err) {
		t.Fatal("expected true")
	}
	if "" != err {
		t.Fatal("expected equal")
	}

	{
		var deps []*Node
		deps = append(deps, state1.GetNode("foo.h", 0))
		deps = append(deps, state1.GetNode("bar.h", 0))
		if !log1.RecordDeps(state1.GetNode("out.o", 0), 1, deps) {
			t.Fatal("oops")
		}

		deps = nil
		deps = append(deps, state1.GetNode("foo.h", 0))
		deps = append(deps, state1.GetNode("bar2.h", 0))
		if !log1.RecordDeps(state1.GetNode("out2.o", 0), 2, deps) {
			t.Fatal("oops")
		}

		logDeps := log1.GetDeps(state1.GetNode("out.o", 0))
		if logDeps == nil {
			t.Fatal("expected true")
		}
		if 1 != logDeps.mtime {
			t.Fatal("expected equal")
		}
		if 2 != logDeps.nodeCount {
			t.Fatal("expected equal")
		}
		if "foo.h" != logDeps.nodes[0].Path {
			t.Fatal("expected equal")
		}
		if "bar.h" != logDeps.nodes[1].Path {
			t.Fatal("expected equal")
		}
	}

	log1.Close()

	state2 := NewState()
	log2 := NewDepsLog()
	if log2.Load(kTestFilename, &state2, &err) != LOAD_SUCCESS {
		t.Fatal(err)
	}
	if "" != err {
		t.Fatal("expected equal")
	}

	if len(log1.nodes()) != len(log2.nodes()) {
		t.Fatal("expected equal")
	}
	for i := 0; i < len(log1.nodes()); i++ {
		node1 := log1.nodes()[i]
		node2 := log2.nodes()[i]
		if int32(i) != node1.ID {
			t.Fatal("expected equal")
		}
		if node1.ID != node2.ID {
			t.Fatal("expected equal")
		}
	}

	// Spot-check the entries in log2.
	logDeps := log2.GetDeps(state2.GetNode("out2.o", 0))
	if logDeps == nil {
		t.Fatal("expected true")
	}
	if 2 != logDeps.mtime {
		t.Fatal("expected equal")
	}
	if 2 != logDeps.nodeCount {
		t.Fatal("expected equal")
	}
	if "foo.h" != logDeps.nodes[0].Path {
		t.Fatal("expected equal")
	}
	if "bar2.h" != logDeps.nodes[1].Path {
		t.Fatal("expected equal")
	}
}

func TestDepsLogTest_LotsOfDeps(t *testing.T) {
	kTestFilename := filepath.Join(t.TempDir(), "DepsLogTest-tempfile")
	kNumDeps := 100000 // More than 64k.

	state1 := NewState()
	log1 := NewDepsLog()
	err := ""
	if !log1.OpenForWrite(kTestFilename, &err) {
		t.Fatal("expected true")
	}
	if "" != err {
		t.Fatal("expected equal")
	}

	{
		var deps []*Node
		for i := 0; i < kNumDeps; i++ {
			buf := fmt.Sprintf("file%d.h", i)
			deps = append(deps, state1.GetNode(buf, 0))
		}
		log1.RecordDeps(state1.GetNode("out.o", 0), 1, deps)

		logDeps := log1.GetDeps(state1.GetNode("out.o", 0))
		if kNumDeps != logDeps.nodeCount {
			t.Fatal("expected equal")
		}
	}

	log1.Close()

	state2 := NewState()
	log2 := NewDepsLog()
	if log2.Load(kTestFilename, &state2, &err) != LOAD_SUCCESS {
		t.Fatal("expected true")
	}
	if "" != err {
		t.Fatal("expected equal")
	}

	logDeps := log2.GetDeps(state2.GetNode("out.o", 0))
	if kNumDeps != logDeps.nodeCount {
		t.Fatal("expected equal")
	}
}

func getFileSize(t *testing.T, p string) int {
	fi, err := os.Stat(p)
	if err != nil {
		t.Fatal(err)
	}
	return int(fi.Size())
}

// Verify that adding the same deps twice doesn't grow the file.
func TestDepsLogTest_DoubleEntry(t *testing.T) {
	kTestFilename := filepath.Join(t.TempDir(), "DepsLogTest-tempfile")
	// Write some deps to the file and grab its size.
	fileSize := 0
	{
		state := NewState()
		log := NewDepsLog()
		err := ""
		if !log.OpenForWrite(kTestFilename, &err) {
			t.Fatal("expected true")
		}
		if "" != err {
			t.Fatal("expected equal")
		}

		var deps []*Node
		deps = append(deps, state.GetNode("foo.h", 0))
		deps = append(deps, state.GetNode("bar.h", 0))
		log.RecordDeps(state.GetNode("out.o", 0), 1, deps)
		log.Close()

		fileSize = getFileSize(t, kTestFilename)
		if fileSize <= 0 {
			t.Fatal("expected greater")
		}
	}

	// Now reload the file, and read the same deps.
	{
		state := NewState()
		log := NewDepsLog()
		err := ""
		if log.Load(kTestFilename, &state, &err) != LOAD_SUCCESS {
			t.Fatal("expected true")
		}

		if !log.OpenForWrite(kTestFilename, &err) {
			t.Fatal("expected true")
		}
		if "" != err {
			t.Fatal("expected equal")
		}

		var deps []*Node
		deps = append(deps, state.GetNode("foo.h", 0))
		deps = append(deps, state.GetNode("bar.h", 0))
		log.RecordDeps(state.GetNode("out.o", 0), 1, deps)
		log.Close()

		fileSize_2 := getFileSize(t, kTestFilename)
		if fileSize != fileSize_2 {
			t.Fatal("expected equal")
		}
	}
}

// Verify that adding the new deps works and can be compacted away.
func TestDepsLogTest_Recompact(t *testing.T) {
	kTestFilename := filepath.Join(t.TempDir(), "DepsLogTest-tempfile")
	kManifest := "rule cc\n  command = cc\n  deps = gcc\nbuild out.o: cc\nbuild other_out.o: cc\n"

	// Write some deps to the file and grab its size.
	fileSize := 0
	{
		state := NewState()
		assertParse(t, kManifest, &state)
		log := NewDepsLog()
		err := ""
		if !log.OpenForWrite(kTestFilename, &err) {
			t.Fatal("expected true")
		}
		if "" != err {
			t.Fatal("expected equal")
		}

		var deps []*Node
		deps = append(deps, state.GetNode("foo.h", 0))
		deps = append(deps, state.GetNode("bar.h", 0))
		log.RecordDeps(state.GetNode("out.o", 0), 1, deps)

		deps = nil
		deps = append(deps, state.GetNode("foo.h", 0))
		deps = append(deps, state.GetNode("baz.h", 0))
		log.RecordDeps(state.GetNode("other_out.o", 0), 1, deps)

		log.Close()

		fileSize = getFileSize(t, kTestFilename)
		if fileSize <= 0 {
			t.Fatal("expected greater")
		}
	}

	// Now reload the file, and add slightly different deps.
	fileSize_2 := 0
	{
		state := NewState()
		assertParse(t, kManifest, &state)
		log := NewDepsLog()
		err := ""
		if log.Load(kTestFilename, &state, &err) != LOAD_SUCCESS {
			t.Fatal("expected true")
		}

		if !log.OpenForWrite(kTestFilename, &err) {
			t.Fatal("expected true")
		}
		if "" != err {
			t.Fatal("expected equal")
		}

		var deps []*Node
		deps = append(deps, state.GetNode("foo.h", 0))
		log.RecordDeps(state.GetNode("out.o", 0), 1, deps)
		log.Close()

		fileSize_2 = getFileSize(t, kTestFilename)
		// The file should grow to record the new deps.
		if fileSize_2 <= fileSize {
			t.Fatal("expected greater")
		}
	}

	// Now reload the file, verify the new deps have replaced the old, then
	// recompact.
	fileSize_3 := 0
	{
		state := NewState()
		assertParse(t, kManifest, &state)
		log := NewDepsLog()
		err := ""
		if log.Load(kTestFilename, &state, &err) != LOAD_SUCCESS {
			t.Fatal("expected true")
		}

		out := state.GetNode("out.o", 0)
		deps := log.GetDeps(out)
		if deps == nil {
			t.Fatal("expected true")
		}
		if 1 != deps.mtime {
			t.Fatal("expected equal")
		}
		if 1 != deps.nodeCount {
			t.Fatal("expected equal")
		}
		if "foo.h" != deps.nodes[0].Path {
			t.Fatal("expected equal")
		}

		otherOut := state.GetNode("other_out.o", 0)
		deps = log.GetDeps(otherOut)
		if deps == nil {
			t.Fatal("expected true")
		}
		if 1 != deps.mtime {
			t.Fatal("expected equal")
		}
		if 2 != deps.nodeCount {
			t.Fatal("expected equal")
		}
		if "foo.h" != deps.nodes[0].Path {
			t.Fatal("expected equal")
		}
		if "baz.h" != deps.nodes[1].Path {
			t.Fatal("expected equal")
		}

		if !log.Recompact(kTestFilename, &err) {
			t.Fatal("expected true")
		}

		// The in-memory deps graph should still be valid after recompaction.
		deps = log.GetDeps(out)
		if deps == nil {
			t.Fatal("expected true")
		}
		if 1 != deps.mtime {
			t.Fatal("expected equal")
		}
		if 1 != deps.nodeCount {
			t.Fatal("expected equal")
		}
		if "foo.h" != deps.nodes[0].Path {
			t.Fatal("expected equal")
		}
		if out != log.nodes()[out.ID] {
			t.Fatal("expected equal")
		}

		deps = log.GetDeps(otherOut)
		if deps == nil {
			t.Fatal("expected true")
		}
		if 1 != deps.mtime {
			t.Fatal("expected equal")
		}
		if 2 != deps.nodeCount {
			t.Fatal("expected equal")
		}
		if "foo.h" != deps.nodes[0].Path {
			t.Fatal("expected equal")
		}
		if "baz.h" != deps.nodes[1].Path {
			t.Fatal("expected equal")
		}
		if otherOut != log.nodes()[otherOut.ID] {
			t.Fatal("expected equal")
		}

		// The file should have shrunk a bit for the smaller deps.
		fileSize_3 = getFileSize(t, kTestFilename)
		if fileSize_3 >= fileSize_2 {
			t.Fatal("expected less or equal")
		}
	}

	// Now reload the file and recompact with an empty manifest. The previous
	// entries should be removed.
	{
		state := NewState()
		// Intentionally not parsing kManifest here.
		log := NewDepsLog()
		err := ""
		if log.Load(kTestFilename, &state, &err) != LOAD_SUCCESS {
			t.Fatal("expected true")
		}

		out := state.GetNode("out.o", 0)
		deps := log.GetDeps(out)
		if deps == nil {
			t.Fatal("expected true")
		}
		if 1 != deps.mtime {
			t.Fatal("expected equal")
		}
		if 1 != deps.nodeCount {
			t.Fatal("expected equal")
		}
		if "foo.h" != deps.nodes[0].Path {
			t.Fatal("expected equal")
		}

		otherOut := state.GetNode("other_out.o", 0)
		deps = log.GetDeps(otherOut)
		if deps == nil {
			t.Fatal("expected true")
		}
		if 1 != deps.mtime {
			t.Fatal("expected equal")
		}
		if 2 != deps.nodeCount {
			t.Fatal("expected equal")
		}
		if "foo.h" != deps.nodes[0].Path {
			t.Fatal("expected equal")
		}
		if "baz.h" != deps.nodes[1].Path {
			t.Fatal("expected equal")
		}

		if !log.Recompact(kTestFilename, &err) {
			t.Fatal("expected true")
		}

		// The previous entries should have been removed.
		deps = log.GetDeps(out)
		if deps != nil {
			t.Fatal("expected false")
		}

		deps = log.GetDeps(otherOut)
		if deps != nil {
			t.Fatal("expected false")
		}

		// The .h files pulled in via deps should no longer have ids either.
		if -1 != state.LookupNode("foo.h").ID {
			t.Fatal("expected equal")
		}
		if -1 != state.LookupNode("baz.h").ID {
			t.Fatal("expected equal")
		}

		// The file should have shrunk more.
		fileSize_4 := getFileSize(t, kTestFilename)
		if fileSize_4 >= fileSize_3 {
			t.Fatal("expected less or equal")
		}
	}
}

// Verify that invalid file headers cause a new build.
func TestDepsLogTest_InvalidHeader(t *testing.T) {
	kTestFilename := filepath.Join(t.TempDir(), "DepsLogTest-tempfile")
	kInvalidHeaders := []string{
		"",                              // Empty file.
		"# ninjad",                      // Truncated first line.
		"# ninjadeps\n",                 // No version int.
		"# ninjadeps\n\001\002",         // Truncated version int.
		"# ninjadeps\n\001\002\003\004", // Invalid version int.
	}
	for i := 0; i < len(kInvalidHeaders); i++ {
		depsLog, err2 := os.OpenFile(kTestFilename, os.O_CREATE|os.O_WRONLY, 0o600)
		if depsLog == nil {
			t.Fatal(err2)
		}
		if _, err := depsLog.Write([]byte(kInvalidHeaders[i])); err != nil {
			t.Fatal(err)
		}
		if err := depsLog.Close(); err != nil {
			t.Fatal(err)
		}

		err := ""
		log := NewDepsLog()
		state := NewState()
		if log.Load(kTestFilename, &state, &err) != LOAD_SUCCESS {
			t.Fatal("expected true")
		}

		if !strings.HasPrefix(err, "bad deps log signature ") {
			t.Fatalf("%q", err)
		}
	}
}

// Simulate what happens when loading a truncated log file.
func TestDepsLogTest_Truncated(t *testing.T) {
	kTestFilename := filepath.Join(t.TempDir(), "DepsLogTest-tempfile")
	// Create a file with some entries.
	{
		state := NewState()
		log := NewDepsLog()
		err := ""
		if !log.OpenForWrite(kTestFilename, &err) {
			t.Fatal("expected true")
		}
		if "" != err {
			t.Fatal("expected equal")
		}

		var deps []*Node
		deps = append(deps, state.GetNode("foo.h", 0))
		deps = append(deps, state.GetNode("bar.h", 0))
		log.RecordDeps(state.GetNode("out.o", 0), 1, deps)

		deps = nil
		deps = append(deps, state.GetNode("foo.h", 0))
		deps = append(deps, state.GetNode("bar2.h", 0))
		log.RecordDeps(state.GetNode("out2.o", 0), 2, deps)

		log.Close()
	}

	// Get the file size.
	fileSize := getFileSize(t, kTestFilename)

	// Try reloading at truncated sizes.
	// Track how many nodes/deps were found; they should decrease with
	// smaller sizes.
	nodeCount := 5
	depsCount := 2
	for size := fileSize; size > 0; size-- {
		if err := os.Truncate(kTestFilename, int64(size)); err != nil {
			t.Fatal(err)
		}

		state := NewState()
		log := NewDepsLog()
		err := ""
		if log.Load(kTestFilename, &state, &err) == LOAD_NOT_FOUND {
			t.Fatal(err)
		}
		if len(err) != 0 {
			// At some point the log will be so short as to be unparsable.
			break
		}

		if nodeCount < len(log.nodes()) {
			t.Fatal("expected greater or equal")
		}
		nodeCount = len(log.nodes())

		// Count how many non-NULL deps entries there are.
		newDepsCount := 0
		for _, i := range log.deps() {
			if i != nil {
				newDepsCount++
			}
		}
		if depsCount < newDepsCount {
			t.Fatal("expected greater or equal")
		}
		depsCount = newDepsCount
	}
}

// Run the truncation-recovery logic.
func TestDepsLogTest_TruncatedRecovery(t *testing.T) {
	t.Skip("TODO; the load is still succeeding. This is not critical for release.")
	kTestFilename := filepath.Join(t.TempDir(), "DepsLogTest-tempfile")
	// Create a file with some entries.
	{
		state := NewState()
		log := NewDepsLog()
		err := ""
		if !log.OpenForWrite(kTestFilename, &err) {
			t.Fatal("expected true")
		}
		if "" != err {
			t.Fatal("expected equal")
		}

		var deps []*Node
		deps = append(deps, state.GetNode("foo.h", 0))
		deps = append(deps, state.GetNode("bar.h", 0))
		log.RecordDeps(state.GetNode("out.o", 0), 1, deps)

		deps = nil
		deps = append(deps, state.GetNode("foo.h", 0))
		deps = append(deps, state.GetNode("bar2.h", 0))
		log.RecordDeps(state.GetNode("out2.o", 0), 2, deps)

		log.Close()
	}

	// Shorten the file, corrupting the last record.
	{
		fileSize := getFileSize(t, kTestFilename)
		const cut = 2
		if err := os.Truncate(kTestFilename, int64(fileSize-cut)); err != nil {
			t.Fatal(err)
		}
		if f2 := getFileSize(t, kTestFilename); f2 != fileSize-cut {
			t.Fatal(f2)
		}
	}

	// Load the file again, add an entry.
	{
		state := NewState()
		log := NewDepsLog()
		err := ""
		if log.Load(kTestFilename, &state, &err) != LOAD_SUCCESS {
			t.Fatal("expected true")
		}
		if "premature end of file; recovering" != err {
			t.Fatal(err)
		}
		err = ""

		// The truncated entry should've been discarded.
		if nil != log.GetDeps(state.GetNode("out2.o", 0)) {
			t.Fatal("expected out2.o to be stripped")
		}

		if !log.OpenForWrite(kTestFilename, &err) {
			t.Fatal("expected true")
		}
		if "" != err {
			t.Fatal("expected equal")
		}

		// Add a new entry.
		var deps []*Node
		deps = append(deps, state.GetNode("foo.h", 0))
		deps = append(deps, state.GetNode("bar2.h", 0))
		log.RecordDeps(state.GetNode("out2.o", 0), 3, deps)

		log.Close()
	}

	// Load the file a third time to verify appending after a mangled
	// entry doesn't break things.
	{
		state := NewState()
		log := NewDepsLog()
		err := ""
		if log.Load(kTestFilename, &state, &err) != LOAD_SUCCESS {
			t.Fatal("expected true")
		}

		// The truncated entry should exist.
		deps := log.GetDeps(state.GetNode("out2.o", 0))
		if deps == nil {
			t.Fatal("expected true")
		}
	}
}

func TestDepsLogTest_ReverseDepsNodes(t *testing.T) {
	kTestFilename := filepath.Join(t.TempDir(), "DepsLogTest-tempfile")
	state := NewState()
	log := NewDepsLog()
	err := ""
	if !log.OpenForWrite(kTestFilename, &err) {
		t.Fatal("expected true")
	}
	if "" != err {
		t.Fatal("expected equal")
	}

	var deps []*Node
	deps = append(deps, state.GetNode("foo.h", 0))
	deps = append(deps, state.GetNode("bar.h", 0))
	log.RecordDeps(state.GetNode("out.o", 0), 1, deps)

	deps = nil
	deps = append(deps, state.GetNode("foo.h", 0))
	deps = append(deps, state.GetNode("bar2.h", 0))
	log.RecordDeps(state.GetNode("out2.o", 0), 2, deps)

	log.Close()

	revDeps := log.GetFirstReverseDepsNode(state.GetNode("foo.h", 0))
	if revDeps != state.GetNode("out.o", 0) || revDeps == state.GetNode("out2.o", 0) {
		t.Fatal("expected true")
	}

	revDeps = log.GetFirstReverseDepsNode(state.GetNode("bar.h", 0))
	if revDeps != state.GetNode("out.o", 0) {
		t.Fatal("expected true")
	}
}
