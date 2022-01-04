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
	"bufio"
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"io/ioutil"
	"os"
)

// Reading (startup-time) interface.
type Deps struct {
	MTime TimeStamp
	Nodes []*Node
}

func NewDeps(mtime TimeStamp, nodeCount int) *Deps {
	return &Deps{
		MTime: mtime,
		Nodes: make([]*Node, nodeCount),
	}
}

// As build commands run they can output extra dependency information
// (e.g. header dependencies for C source) dynamically.  DepsLog collects
// that information at build time and uses it for subsequent builds.
//
// The on-disk format is based on two primary design constraints:
// - it must be written to as a stream (during the build, which may be
//   interrupted);
// - it can be read all at once on startup.  (Alternative designs, where
//   it contains indexing information, were considered and discarded as
//   too complicated to implement; if the file is small than reading it
//   fully on startup is acceptable.)
// Here are some stats from the Windows Chrome dependency files, to
// help guide the design space.  The total text in the files sums to
// 90mb so some compression is warranted to keep load-time fast.
// There's about 10k files worth of dependencies that reference about
// 40k total paths totalling 2mb of unique strings.
//
// Based on these stats, here's the current design.
// The file is structured as version header followed by a sequence of records.
// Each record is either a path string or a dependency list.
// Numbering the path strings in file order gives them dense integer ids.
// A dependency list maps an output id to a list of input ids.
//
// Concretely, a record is:
//    four bytes record length, high bit indicates record type
//      (but max record sizes are capped at 512kB)
//    path records contain the string name of the path, followed by up to 3
//      padding bytes to align on 4 byte boundaries, followed by the
//      one's complement of the expected index of the record (to detect
//      concurrent writes of multiple ninja processes to the log).
//    dependency records are an array of 4-byte integers
//      [output path id,
//       output path mtime (lower 4 bytes), output path mtime (upper 4 bytes),
//       input path id, input path id...]
//      (The mtime is compared against the on-disk output path mtime
//      to verify the stored data is up-to-date.)
// If two records reference the same output the latter one in the file
// wins, allowing updates to just be appended to the file.  A separate
// repacking step can run occasionally to remove dead records.
type DepsLog struct {
	needsRecompaction bool
	file              *os.File
	buf               *bufio.Writer
	filePath          string

	// Maps id -> Node.
	Nodes []*Node
	// Maps id -> Deps of that id.
	Deps []*Deps
}

func NewDepsLog() DepsLog {
	return DepsLog{}
}

// The version is stored as 4 bytes after the signature and also serves as a
// byte order mark. Signature and version combined are 16 bytes long.
const (
	DepsLogFileSignature  = "# ninjadeps\n"
	DepsLogCurrentVersion = uint32(4)
)

// Record size is currently limited to less than the full 32 bit, due to
// internal buffers having to have this size.
const maxRecordSize = (1 << 19) - 1

// Writing (build-time) interface.
func (d *DepsLog) OpenForWrite(path string, err *string) bool {
	if d.needsRecompaction {
		if !d.Recompact(path, err) {
			return false
		}
	}

	if d.file != nil {
		panic("M-A")
	}
	// we don't actually open the file right now, but will do
	// so on the first write attempt
	d.filePath = path
	return true
}

func (d *DepsLog) RecordDeps(node *Node, mtime TimeStamp, nodes []*Node) bool {
	nodeCount := len(nodes)
	// Track whether there's any new data to be recorded.
	madeChange := false

	// Assign ids to all nodes that are missing one.
	if node.ID < 0 {
		if !d.RecordID(node) {
			return false
		}
		madeChange = true
	}
	for i := 0; i < nodeCount; i++ {
		if nodes[i].ID < 0 {
			if !d.RecordID(nodes[i]) {
				return false
			}
			madeChange = true
		}
	}

	// See if the new data is different than the existing data, if any.
	if !madeChange {
		deps := d.GetDeps(node)
		if deps == nil || deps.MTime != mtime || len(deps.Nodes) != nodeCount {
			madeChange = true
		} else {
			for i := 0; i < nodeCount; i++ {
				if deps.Nodes[i] != nodes[i] {
					madeChange = true
					break
				}
			}
		}
	}

	// Don't write anything if there's no new info.
	if !madeChange {
		return true
	}

	// Update on-disk representation.
	size := uint32(4 * (1 + 2 + nodeCount))
	if size > maxRecordSize {
		//errno = ERANGE
		return false
	}

	if !d.OpenForWriteIfNeeded() {
		return false
	}
	size |= 0x80000000 // Deps record: set high bit.

	if err := binary.Write(d.buf, binary.LittleEndian, size); err != nil {
		return false
	}
	if err := binary.Write(d.buf, binary.LittleEndian, uint32(node.ID)); err != nil {
		return false
	}
	if err := binary.Write(d.buf, binary.LittleEndian, mtime); err != nil {
		return false
	}
	for i := 0; i < nodeCount; i++ {
		if err := binary.Write(d.buf, binary.LittleEndian, uint32(nodes[i].ID)); err != nil {
			return false
		}
	}
	if err := d.buf.Flush(); err != nil {
		return false
	}

	// Update in-memory representation.
	deps := NewDeps(mtime, nodeCount)
	for i := 0; i < nodeCount; i++ {
		deps.Nodes[i] = nodes[i]
	}
	d.UpdateDeps(node.ID, deps)

	return true
}

func (d *DepsLog) Close() error {
	// create the file even if nothing has been recorded
	// TODO(maruel): Error handling.
	d.OpenForWriteIfNeeded()
	var err error
	if d.file != nil {
		if err2 := d.buf.Flush(); err2 != nil {
			err = err2
		}
		if err2 := d.file.Close(); err2 != nil {
			err = err2
		}
	}
	d.buf = nil
	d.file = nil
	return err
}

// Load loads a .ninja_deps to accelerate incremental build.
//
// Note: For version differences, this should migrate to the new format.
// But the v1 format could sometimes (rarely) end up with invalid data, so
// don't migrate v1 to v3 to force a rebuild. (v2 only existed for a few days,
// and there was no release with it, so pretend that it never happened.)
func (d *DepsLog) Load(path string, state *State, err *string) LoadStatus {
	defer metricRecord(".ninja_deps load")()
	data, err2 := ioutil.ReadFile(path)
	if err2 != nil {
		if os.IsNotExist(err2) {
			return LoadNotFound
		}
		*err = err2.Error()
		return LoadError
	}
	// Read the file all at once then use a buffer.
	validHeader := false
	version := uint32(0)
	if len(data) >= len(DepsLogFileSignature)+4 && unsafeString(data[:len(DepsLogFileSignature)]) == DepsLogFileSignature {
		version = binary.LittleEndian.Uint32(data[len(DepsLogFileSignature):])
		validHeader = version == DepsLogCurrentVersion
	}
	if !validHeader {
		if version == 1 {
			*err = "deps log version change; rebuilding"
		} else {
			l := bytes.IndexByte(data[:], 0)
			if l <= 0 {
				*err = "bad deps log signature or version; starting over"
			} else {
				*err = fmt.Sprintf("bad deps log signature %q or version %d; starting over", data[:l], version)
			}
		}
		_ = os.Remove(path)
		// Don't report this as a failure.  An empty deps log will cause
		// us to rebuild the outputs anyway.
		return LoadSuccess
	}

	offset := len(DepsLogFileSignature) + 4
	// TODO(maruel): Use binary.LittleEndian.Uint32() directly instead of using a
	// io.Reader.
	f := bytes.NewReader(data[offset:])
	buf := [maxRecordSize + 1]byte{}
	readFailed := false
	uniqueDepRecordCount := 0
	totalDepRecordCount := 0
	for {
		var size uint32
		if err2 := binary.Read(f, binary.LittleEndian, &size); err2 != nil {
			if err2 != io.EOF {
				readFailed = true
				*err = err2.Error()
			}
			break
		}
		isDeps := size&0x80000000 != 0
		size = size & 0x7FFFFFFF
		if size > maxRecordSize {
			readFailed = true
			// TODO(maruel): Make it a real error.
			break
		}
		if _, err2 := f.Read(buf[:size]); err2 != nil {
			readFailed = true
			*err = err2.Error()
			break
		}

		if isDeps {
			if size%4 != 0 || size < 12 {
				// TODO(maruel): Make it a real error.
				break
			}
			// TODO(maruel): Not super efficient but looking for correctness now.
			// Optimize later.
			r := bytes.NewReader(buf[:size])
			outID := int32(0)
			if err2 := binary.Read(r, binary.LittleEndian, &outID); err2 != nil {
				panic(err2)
			}
			// TODO(maruel): It seems like it's registering invalid IDs.
			if outID < 0 || outID >= 0x1000000 {
				// That's a lot of nodes.
				readFailed = true
				// TODO(maruel): Make it a real error.
				break
			}
			var mtime TimeStamp
			if err2 := binary.Read(r, binary.LittleEndian, &mtime); err2 != nil {
				// TODO(maruel): Make it a real error.
				panic(err2)
			}
			depsCount := int(size)/4 - 3

			deps := NewDeps(mtime, depsCount)
			for i := 0; i < depsCount; i++ {
				v := uint32(0)
				if err2 := binary.Read(r, binary.LittleEndian, &v); err2 != nil {
					panic(err2)
				}
				if int(v) >= len(d.Nodes) {
					panic("M-A")
				}
				if d.Nodes[v] == nil {
					panic("M-A")
				}
				deps.Nodes[i] = d.Nodes[v]
			}

			totalDepRecordCount++
			if !d.UpdateDeps(outID, deps) {
				uniqueDepRecordCount++
			}
		} else {
			pathSize := size - 4
			if pathSize <= 0 {
				panic("M-A")
			} // CanonicalizePath() rejects empty paths.
			// There can be up to 3 bytes of padding.
			if buf[pathSize-1] == '\x00' {
				pathSize--
			}
			if buf[pathSize-1] == '\x00' {
				pathSize--
			}
			if buf[pathSize-1] == '\x00' {
				pathSize--
			}
			subpath := string(buf[:pathSize])
			// It is not necessary to pass in a correct slashBits here. It will
			// either be a Node that's in the manifest (in which case it will already
			// have a correct slashBits that GetNode will look up), or it is an
			// implicit dependency from a .d which does not affect the build command
			// (and so need not have its slashes maintained).
			node := state.GetNode(subpath, 0)

			// Check that the expected index matches the actual index. This can only
			// happen if two ninja processes write to the same deps log concurrently.
			// (This uses unary complement to make the checksum look less like a
			// dependency record entry.)
			checksum := uint32(0) // *reinterpretCast<unsigned*>(buf + size - 4)
			if err2 := binary.Read(bytes.NewReader(buf[size-4:]), binary.LittleEndian, &checksum); err2 != nil {
				panic(err2)
			}
			expectedID := ^checksum
			id := int32(len(d.Nodes))
			if id != int32(expectedID) {
				readFailed = true
				// TODO(maruel): Make it a real error.
				break
			}

			if node.ID >= 0 {
				panic("M-A")
			}
			node.ID = id
			d.Nodes = append(d.Nodes, node)
		}
	}

	if readFailed {
		// An error occurred while loading; try to recover by truncating the
		// file to the last fully-read record.
		if *err == "" {
			*err = "premature end of file"
		}

		// TODO(maruel): Not exactly right.
		offset, _ := f.Seek(0, io.SeekCurrent)
		if err := os.Truncate(path, offset); err != nil {
			return LoadError
		}

		// The truncate succeeded; we'll just report the load error as a
		// warning because the build can proceed.
		*err += "; recovering"
		return LoadSuccess
	}

	// Rebuild the log if there are too many dead records.
	const minCompactionEntryCount = 1000
	kCompactionRatio := 3
	if totalDepRecordCount > minCompactionEntryCount && totalDepRecordCount > uniqueDepRecordCount*kCompactionRatio {
		d.needsRecompaction = true
	}
	return LoadSuccess
}

func (d *DepsLog) GetDeps(node *Node) *Deps {
	// Abort if the node has no id (never referenced in the deps) or if
	// there's no deps recorded for the node.
	if node.ID < 0 || int(node.ID) >= len(d.Deps) {
		return nil
	}
	return d.Deps[node.ID]
}

func (d *DepsLog) GetFirstReverseDepsNode(node *Node) *Node {
	for id := 0; id < len(d.Deps); id++ {
		deps := d.Deps[id]
		if deps == nil {
			continue
		}
		for _, n := range deps.Nodes {
			if n == node {
				return d.Nodes[id]
			}
		}
	}
	return nil
}

// Rewrite the known log entries, throwing away old data.
func (d *DepsLog) Recompact(path string, err *string) bool {
	defer metricRecord(".ninja_deps recompact")()

	_ = d.Close()
	tempPath := path + ".recompact"

	// OpenForWrite() opens for append.  Make sure it's not appending to a
	// left-over file from a previous recompaction attempt that crashed somehow.
	if err2 := os.Remove(tempPath); err2 != nil && !os.IsNotExist(err2) {
		*err = err2.Error()
		return false
	}

	newLog := NewDepsLog()
	if !newLog.OpenForWrite(tempPath, err) {
		return false
	}

	// Clear all known ids so that new ones can be reassigned.  The new indices
	// will refer to the ordering in newLog, not in the current log.
	for _, i := range d.Nodes {
		i.ID = -1
	}

	// Write out all deps again.
	for oldID := 0; oldID < len(d.Deps); oldID++ {
		deps := d.Deps[oldID]
		if deps == nil { // If nodes[oldID] is a leaf, it has no deps.
			continue
		}

		if !d.IsDepsEntryLiveFor(d.Nodes[oldID]) {
			continue
		}

		if !newLog.RecordDeps(d.Nodes[oldID], deps.MTime, deps.Nodes) {
			_ = newLog.Close()
			return false
		}
	}

	_ = newLog.Close()

	// All nodes now have ids that refer to newLog, so steal its data.
	d.Deps = newLog.Deps
	d.Nodes = newLog.Nodes

	if err2 := os.Remove(path); err2 != nil {
		*err = err2.Error()
		return false
	}

	if err2 := os.Rename(tempPath, path); err2 != nil {
		*err = err2.Error()
		return false
	}
	return true
}

// Returns if the deps entry for a node is still reachable from the manifest.
//
// The deps log can contain deps entries for files that were built in the
// past but are no longer part of the manifest.  This function returns if
// this is the case for a given node.  This function is slow, don't call
// it from code that runs on every build.
func (d *DepsLog) IsDepsEntryLiveFor(node *Node) bool {
	// Skip entries that don't have in-edges or whose edges don't have a
	// "deps" attribute. They were in the deps log from previous builds, but
	// the the files they were for were removed from the build and their deps
	// entries are no longer needed.
	// (Without the check for "deps", a chain of two or more nodes that each
	// had deps wouldn't be collected in a single recompaction.)
	return node.InEdge != nil && node.InEdge.GetBinding("deps") != ""
}

// Updates the in-memory representation.  Takes ownership of |deps|.
// Returns true if a prior deps record was deleted.
func (d *DepsLog) UpdateDeps(outID int32, deps *Deps) bool {
	if n := int(outID) + 1 - len(d.Deps); n > 0 {
		d.Deps = append(d.Deps, make([]*Deps, n)...)
	}
	existed := d.Deps[outID] != nil
	d.Deps[outID] = deps
	return existed
}

// Write a node name record, assigning it an id.
func (d *DepsLog) RecordID(node *Node) bool {
	pathSize := len(node.Path)
	padding := (4 - pathSize%4) % 4 // Pad path to 4 byte boundary.

	size := uint32(pathSize + padding + 4)
	if size > maxRecordSize {
		// TODO(maruel): Make it a real error.
		//errno = ERANGE
		return false
	}

	if !d.OpenForWriteIfNeeded() {
		// TODO(maruel): Make it a real error.
		return false
	}
	if err := binary.Write(d.buf, binary.LittleEndian, size); err != nil {
		// TODO(maruel): Make it a real error.
		return false
	}
	if _, err := d.buf.WriteString(node.Path); err != nil {
		if node.Path == "" {
			panic("M-A")
		}
		// TODO(maruel): Make it a real error.
		return false
	}
	if padding != 0 {
		if _, err := d.buf.Write(make([]byte, padding)); err != nil {
			// TODO(maruel): Make it a real error.
			return false
		}
	}
	id := int32(len(d.Nodes))
	checksum := ^uint32(id)
	if err := binary.Write(d.buf, binary.LittleEndian, checksum); err != nil {
		// TODO(maruel): Make it a real error.
		return false
	}
	if err := d.buf.Flush(); err != nil {
		return false
	}
	node.ID = id
	d.Nodes = append(d.Nodes, node)

	return true
}

// Should be called before using file. When false is returned, errno will
// be set.
func (d *DepsLog) OpenForWriteIfNeeded() bool {
	if d.filePath == "" {
		return true
	}
	if d.file != nil {
		panic("surprising state")
	}
	d.file, _ = os.OpenFile(d.filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o666)
	if d.file == nil {
		// TODO(maruel): Make it a real error.
		return false
	}
	// Set the buffer size to this and flush the file buffer after every record
	// to make sure records aren't written partially.
	d.buf = bufio.NewWriterSize(d.file, maxRecordSize+1)
	//SetCloseOnExec(fileno(d.file))

	// Opening a file in append mode doesn't set the file pointer to the file's
	// end on Windows. Do that explicitly.
	offset, err := d.file.Seek(0, os.SEEK_END)

	if err != nil {
		// TODO(maruel): Make it a real error.
		return false
	}

	if offset == 0 {
		if _, err := d.buf.WriteString(DepsLogFileSignature); err != nil {
			// TODO(maruel): Return the real error.
			return false
		}
		if err := binary.Write(d.buf, binary.LittleEndian, DepsLogCurrentVersion); err != nil {
			// TODO(maruel): Return the real error.
			return false
		}
	}
	if err := d.buf.Flush(); err != nil {
		return false
	}
	d.filePath = ""
	return true
}
