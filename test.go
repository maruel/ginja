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

//go:build nobuild

package ginja


// A tiny testing framework inspired by googletest, but much simpler and
// faster to compile. It supports most things commonly used from googltest. The
// most noticeable things missing: EXPECT_* and ASSERT_* don't support
// streaming notes to them with operator<<, and for failing tests the lhs and
// rhs are not printed. That's so that this header does not have to include
// sstream, which slows down building ninja_test almost 20%.
class Test {
  bool failed_
  int assertion_failures_
  Test() : failed_(false), assertion_failures_(0) {}
  virtual ~Test() {}
  func SetUp() {}
  func TearDown() {}

  func Failed() bool {
	return failed_
}
  func AssertionFailures() int {
	return assertion_failures_
}
  func AddAssertionFailure() {
	assertion_failures_++
}
}

void RegisterTest(testing::Test* (*)(), string)

//extern testing::Test* g_current_test

// Support utilities for tests.

// A base test fixture that includes a State object with a
// builtin "cat" rule.
type StateTestWithBuiltinRules struct {

  state_ State
}

void AssertParse(State* state, string input, ManifestParserOptions = ManifestParserOptions())

// An implementation of DiskInterface that uses an in-memory representation
// of disk state.  It also logs file accesses and directory creations
// so it can be used by tests to verify disk access patterns.
type VirtualFileSystem struct {

  directories_made_ []string
  files_read_ []string
  files_ FileMap
  files_removed_ map[string]struct{}
  files_created_ map[string]struct{}

  // A simple fake timestamp for file operations.
  now_ int
}
func NewVirtualFileSystem() VirtualFileSystem {
	return VirtualFileSystem{
		now_: 1,
	}
}
// Tick "time" forwards; subsequent file operations will be newer than
// previous ones.
func (v *VirtualFileSystem) Tick() int {
  return ++v.now_
}
// An entry for a single in-memory file.
type Entry struct {
  mtime int
  stat_error string  // If mtime is -1.
  contents string
  }
type FileMap map[string]Entry

type ScopedTempDir struct {

  // The temp directory containing our dir.
  start_dir_ string
  // The subdirectory name for our dir, or empty if it hasn't been set up.
  temp_dir_name_ string
}


//extern "C" {
        //extern char* mkdtemp(char* name_template)

namespace {

// Windows has no mkdtemp.  Implement it in terms of _mktemp_s.
func mkdtemp(name_template *char) *char {
  int err = _mktemp_s(name_template, strlen(name_template) + 1)
  if err < 0 {
    perror("_mktemp_s")
    return nil
  }

  err = _mkdir(name_template)
  if err < 0 {
    perror("mkdir")
    return nil
  }

  return name_template
}

func GetSystemTempDir() string {
  char buf[1024]
  if !GetTempPath(sizeof(buf), buf) {
    return ""
  }
  return buf
  string tempdir = getenv("TMPDIR")
  if tempdir != nil {
    return tempdir
  }
  return "/tmp"
}

}  // anonymous namespace

StateTestWithBuiltinRules::StateTestWithBuiltinRules() {
  AddCatRule(&state_)
}

// Add a "cat" rule to \a state.  Used by some tests; it's
// otherwise done by the ctor to state_.
func (s *StateTestWithBuiltinRules) AddCatRule(state *State) {
  AssertParse(state, "rule cat\n  command = cat $in > $out\n")
}

// Short way to get a Node by its path from state_.
func (s *StateTestWithBuiltinRules) GetNode(path string) *Node {
  if !strpbrk(path, "/\\") { t.FailNow() }
  return s.state_.GetNode(path, 0)
}

func (s *StateTestWithBuiltinRules) AssertParse(state *State, input string, opts ManifestParserOptions) {
  ManifestParser parser(state, nil, opts)
  err := ""
  if parser.ParseTest(input, &err) { t.FailNow() }
  if "" != err { t.FailNow() }
  VerifyGraph(*state)
}

func (s *StateTestWithBuiltinRules) AssertHash(expected string, actual uint64) {
  if BuildLog::LogEntry::HashCommand(expected) != actual { t.FailNow() }
}

func (s *StateTestWithBuiltinRules) VerifyGraph(state *State) {
  for e := state.edges_.begin(); e != state.edges_.end(); e++ {
    // All edges need at least one output.
    if !(*e).outputs_.empty() { t.FailNow() }
    // Check that the edge's inputs have the edge as out-edge.
    for in_node := (*e).inputs_.begin(); in_node != (*e).inputs_.end(); in_node++ {
      out_edges := (*in_node).out_edges()
      if find(out_edges.begin() == out_edges.end(), *e), out_edges.end() { t.FailNow() }
    }
    // Check that the edge's outputs have the edge as in-edge.
    for out_node := (*e).outputs_.begin(); out_node != (*e).outputs_.end(); out_node++ {
      if (*out_node).in_edge() != *e { t.FailNow() }
    }
  }

  // The union of all in- and out-edges of each nodes should be exactly edges_.
  var node_edge_set map[*Edge]struct{}
  for p := state.paths_.begin(); p != state.paths_.end(); p++ {
    n := p.second
    if n.in_edge() {
      node_edge_set.insert(n.in_edge())
    }
    node_edge_set.insert(n.out_edges().begin(), n.out_edges().end())
  }
  set<const Edge*> edge_set(state.edges_.begin(), state.edges_.end())
  if node_edge_set != edge_set { t.FailNow() }
}

// "Create" a file with contents.
func (v *VirtualFileSystem) Create(path string, contents string) {
  v.files_[path].mtime = v.now_
  v.files_[path].contents = contents
  v.files_created_.insert(path)
}

// DiskInterface
func (v *VirtualFileSystem) Stat(path string, err *string) TimeStamp {
  i := v.files_.find(path)
  if i != v.files_.end() {
    *err = i.second.stat_error
    return i.second.mtime
  }
  return 0
}

func (v *VirtualFileSystem) WriteFile(path string, contents string) bool {
  Create(path, contents)
  return true
}

func (v *VirtualFileSystem) MakeDir(path string) bool {
  v.directories_made_.push_back(path)
  return true  // success
}

func (v *VirtualFileSystem) ReadFile(path string, contents *string, err *string) FileReader::Status {
  v.files_read_.push_back(path)
  i := v.files_.find(path)
  if i != v.files_.end() {
    *contents = i.second.contents
    return Okay
  }
  *err = strerror(ENOENT)
  return NotFound
}

func (v *VirtualFileSystem) RemoveFile(path string) int {
  if find(v.directories_made_.begin(), v.directories_made_.end(), path) != v.directories_made_.end() {
    return -1
  }
  i := v.files_.find(path)
  if i != v.files_.end() {
    v.files_.erase(i)
    v.files_removed_.insert(path)
    return 0
  } else {
    return 1
  }
}

// Create a temporary directory and chdir into it.
func (s *ScopedTempDir) CreateAndEnter(name string) {
  // First change into the system temp dir and save it for cleanup.
  s.start_dir_ = GetSystemTempDir()
  if s.start_dir_.empty() {
    Fatal("couldn't get system temp dir")
  }
  if chdir(s.start_dir_) < 0 {
    Fatal("chdir: %s", strerror(errno))
  }

  // Create a temporary subdirectory of that.
  char name_template[1024]
  strcpy(name_template, name)
  strcat(name_template, "-XXXXXX")
  tempname := mkdtemp(name_template)
  if tempname == nil {
    Fatal("mkdtemp: %s", strerror(errno))
  }
  s.temp_dir_name_ = tempname

  // chdir into the new temporary directory.
  if chdir(s.temp_dir_name_) < 0 {
    Fatal("chdir: %s", strerror(errno))
  }
}

// Clean up the temporary directory.
func (s *ScopedTempDir) Cleanup() {
  if s.temp_dir_name_.empty() {
    return  // Something went wrong earlier.
  }

  // Move out of the directory we're about to clobber.
  if chdir(s.start_dir_) < 0 {
    Fatal("chdir: %s", strerror(errno))
  }

  string command = "rmdir /s /q " + s.temp_dir_name_
  string command = "rm -rf " + s.temp_dir_name_
  if system(command) < 0 {
    Fatal("system: %s", strerror(errno))
  }

  s.temp_dir_name_ = nil
}

