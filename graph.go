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


// Information about a node in the dependency graph: the file, whether
// it's dirty, mtime, etc.
type Node struct {
  Node(string path, uint64_t slash_bits)
      : path_(path),
        slash_bits_(slash_bits),
        mtime_(-1),
        exists_(ExistenceStatusUnknown),
        dirty_(false),
        dyndep_pending_(false),
        in_edge_(nil),
        id_(-1) {}

  // Return false on error.

  // Mark as not-yet-stat()ed and not dirty.

  // Mark the Node as already-stat()ed and missing.

  string path() const { return path_; }
  // Get |path()| but use slash_bits to convert back to original slash styles.
  uint64_t slash_bits() const { return slash_bits_; }

  TimeStamp mtime() const { return mtime_; }

  bool dirty() const { return dirty_; }
  void set_dirty(bool dirty) { dirty_ = dirty; }
  void MarkDirty() { dirty_ = true; }

  bool dyndep_pending() const { return dyndep_pending_; }
  void set_dyndep_pending(bool pending) { dyndep_pending_ = pending; }

  Edge* in_edge() const { return in_edge_; }
  void set_in_edge(Edge* edge) { in_edge_ = edge; }

  int id() const { return id_; }
  void set_id(int id) { id_ = id; }

  const vector<Edge*>& out_edges() const { return out_edges_; }
  void AddOutEdge(Edge* edge) { out_edges_.push_back(edge); }

  void Dump(string prefix="") const

  path_ string

  // Set bits starting from lowest for backslashes that were normalized to
  // forward slashes by CanonicalizePath. See |PathDecanonicalized|.
  slash_bits_ uint64

  // Possible values of mtime_:
  //   -1: file hasn't been examined
  //   0:  we looked, and file doesn't exist
  //   >0: actual file's mtime, or the latest mtime of its dependencies if it doesn't exist
  mtime_ TimeStamp

  exists_ ExistenceStatus

  // Dirty is true when the underlying file is out-of-date.
  // But note that Edge::outputs_ready_ is also used in judging which
  // edges to build.
  dirty_ bool

  // Store whether dyndep information is expected from this node but
  // has not yet been loaded.
  dyndep_pending_ bool

  // The Edge that produces this Node, or NULL when there is no
  // known edge to produce it.
  in_edge_ *Edge

  // All Edges that use this Node as an input.
  out_edges_ []*Edge

  // A dense integer id for the node, assigned and used by DepsLog.
  id_ int
}
  func (n *Node) StatIfNecessary(disk_interface *DiskInterface, err *string) bool {
    if status_known() {
      return true
    }
    return Stat(disk_interface, err)
  }
  func (n *Node) ResetState() {
    mtime_ = -1
    exists_ = ExistenceStatusUnknown
    dirty_ = false
  }
  func (n *Node) MarkMissing() {
    if mtime_ == -1 {
      mtime_ = 0
    }
    exists_ = ExistenceStatusMissing
  }
  func (n *Node) exists() bool {
    return exists_ == ExistenceStatusExists
  }
  func (n *Node) status_known() bool {
    return exists_ != ExistenceStatusUnknown
  }
  func (n *Node) PathDecanonicalized() string {
    return PathDecanonicalized(path_, slash_bits_)
  }
  enum ExistenceStatus {
    // The file hasn't been examined.
    ExistenceStatusUnknown,
    // The file doesn't exist. mtime_ will be the latest mtime of its dependencies.
    ExistenceStatusMissing,
    // The path is an actual file. mtime_ will be the file's mtime.
    ExistenceStatusExists
  }

// An edge in the dependency graph; links between Nodes using Rules.
type Edge struct {

  Edge()
      : rule_(nil), pool_(nil), dyndep_(nil), env_(nil), mark_(VisitNone),
        id_(0), outputs_ready_(false), deps_loaded_(false),
        deps_missing_(false), generated_by_dep_loader_(false),
        implicit_deps_(0), order_only_deps_(0), implicit_outs_(0) {}

  void Dump(string prefix="") const

  rule_ *Rule
  pool_ *Pool
  inputs_ []*Node
  outputs_ []*Node
  dyndep_ *Node
  env_ *BindingEnv
  mark_ VisitMark
  id_ uint
  outputs_ready_ bool
  deps_loaded_ bool
  deps_missing_ bool
  generated_by_dep_loader_ bool

  const Rule& rule() const { return *rule_; }
  Pool* pool() const { return pool_; }
  int weight() const { return 1; }
  bool outputs_ready() const { return outputs_ready_; }

  // There are three types of inputs.
  // 1) explicit deps, which show up as $in on the command line;
  // 2) implicit deps, which the target depends on implicitly (e.g. C headers),
  //                   and changes in them cause the target to rebuild;
  // 3) order-only deps, which are needed before the target builds but which
  //                     don't cause the target to rebuild.
  // These are stored in inputs_ in that order, and we keep counts of
  // #2 and #3 when we need to access the various subsets.
  implicit_deps_ int
  order_only_deps_ int

  // There are two types of outputs.
  // 1) explicit outs, which show up as $out on the command line;
  // 2) implicit outs, which the target generates but are not part of $out.
  // These are stored in outputs_ in that order, and we keep a count of
  // #2 to use when we need to access the various subsets.
  implicit_outs_ int

}
  enum VisitMark {
    VisitNone,
    VisitInStack,
    VisitDone
  }
  func (e *Edge) is_implicit(index uint) bool {
    return index >= inputs_.size() - order_only_deps_ - implicit_deps_ &&
        !is_order_only(index)
  }
  func (e *Edge) is_order_only(index uint) bool {
    return index >= inputs_.size() - order_only_deps_
  }
  func (e *Edge) is_implicit_out(index uint) bool {
    return index >= outputs_.size() - implicit_outs_
  }

type EdgeCmp struct {
  bool operator()(const Edge* a, const Edge* b) const {
    return a.id_ < b.id_
  }
}

typedef set<Edge*, EdgeCmp> EdgeSet

// ImplicitDepLoader loads implicit dependencies, as referenced via the
// "depfile" attribute in build files.
type ImplicitDepLoader struct {
  ImplicitDepLoader(State* state, DepsLog* deps_log, DiskInterface* disk_interface, DepfileParserOptions const* depfile_parser_options)
      : state_(state), disk_interface_(disk_interface), deps_log_(deps_log),
        depfile_parser_options_(depfile_parser_options) {}

  state_ *State
  disk_interface_ *DiskInterface
  deps_log_ *DepsLog
  depfile_parser_options_ *DepfileParserOptions const
}
  func (i *ImplicitDepLoader) deps_log() *DepsLog {
    return deps_log_
  }

// DependencyScan manages the process of scanning the files in a graph
// and updating the dirty/outputs_ready state of all the nodes and edges.
type DependencyScan struct {
  DependencyScan(State* state, BuildLog* build_log, DepsLog* deps_log, DiskInterface* disk_interface, DepfileParserOptions const* depfile_parser_options)
      : build_log_(build_log),
        disk_interface_(disk_interface),
        dep_loader_(state, deps_log, disk_interface, depfile_parser_options),
        dyndep_loader_(state, disk_interface) {}

  build_log_ *BuildLog
  disk_interface_ *DiskInterface
  dep_loader_ ImplicitDepLoader
  dyndep_loader_ DyndepLoader
}
  func (d *DependencyScan) build_log() *BuildLog {
    return build_log_
  }
  func (d *DependencyScan) set_build_log(log *BuildLog) {
    build_log_ = log
  }
  func (d *DependencyScan) deps_log() *DepsLog {
    return dep_loader_.deps_log()
  }


// Return false on error.
func (n *Node) Stat(disk_interface *DiskInterface, err *string) bool {
  METRIC_RECORD("node stat")
  mtime_ = disk_interface.Stat(path_, err)
  if mtime_ == -1 {
    return false
  }
  exists_ = (mtime_ != 0) ? ExistenceStatusExists : ExistenceStatusMissing
  return true
}

// If the file doesn't exist, set the mtime_ from its dependencies
func (n *Node) UpdatePhonyMtime(mtime TimeStamp) {
  if !exists() {
    mtime_ = max(mtime_, mtime)
  }
}

// Update the |dirty_| state of the given node by inspecting its input edge.
// Examine inputs, outputs, and command lines to judge whether an edge
// needs to be re-run, and update outputs_ready_ and each outputs' |dirty_|
// state accordingly.
// Returns false on failure.
func (d *DependencyScan) RecomputeDirty(node *Node, err *string) bool {
  var stack []*Node
  return RecomputeDirty(node, &stack, err)
}

// Update the |dirty_| state of the given node by inspecting its input edge.
// Examine inputs, outputs, and command lines to judge whether an edge
// needs to be re-run, and update outputs_ready_ and each outputs' |dirty_|
// state accordingly.
// Returns false on failure.
func (d *DependencyScan) RecomputeDirty(node *Node, stack *vector<Node*>, err *string) bool {
  edge := node.in_edge()
  if edge == nil {
    // If we already visited this leaf node then we are done.
    if node.status_known() {
      return true
    }
    // This node has no in-edge; it is dirty if it is missing.
    if !node.StatIfNecessary(disk_interface_, err) {
      return false
    }
    if !node.exists() {
      EXPLAIN("%s has no in-edge and is missing", node.path())
    }
    node.set_dirty(!node.exists())
    return true
  }

  // If we already finished this edge then we are done.
  if edge.mark_ == Edge::VisitDone {
    return true
  }

  // If we encountered this edge earlier in the call stack we have a cycle.
  if !VerifyDAG(node, stack, err) {
    return false
  }

  // Mark the edge temporarily while in the call stack.
  edge.mark_ = Edge::VisitInStack
  stack.push_back(node)

  dirty := false
  edge.outputs_ready_ = true
  edge.deps_missing_ = false

  if !edge.deps_loaded_ {
    // This is our first encounter with this edge.
    // If there is a pending dyndep file, visit it now:
    // * If the dyndep file is ready then load it now to get any
    //   additional inputs and outputs for this and other edges.
    //   Once the dyndep file is loaded it will no longer be pending
    //   if any other edges encounter it, but they will already have
    //   been updated.
    // * If the dyndep file is not ready then since is known to be an
    //   input to this edge, the edge will not be considered ready below.
    //   Later during the build the dyndep file will become ready and be
    //   loaded to update this edge before it can possibly be scheduled.
    if edge.dyndep_ && edge.dyndep_.dyndep_pending() {
      if !RecomputeDirty(edge.dyndep_, stack, err) {
        return false
      }

      if !edge.dyndep_.in_edge() || edge.dyndep_.in_edge().outputs_ready() {
        // The dyndep file is ready, so load it now.
        if !LoadDyndeps(edge.dyndep_, err) {
          return false
        }
      }
    }
  }

  // Load output mtimes so we can compare them to the most recent input below.
  for o := edge.outputs_.begin(); o != edge.outputs_.end(); o++ {
    if !(*o).StatIfNecessary(disk_interface_, err) {
      return false
    }
  }

  if !edge.deps_loaded_ {
    // This is our first encounter with this edge.  Load discovered deps.
    edge.deps_loaded_ = true
    if !dep_loader_.LoadDeps(edge, err) {
      if len(err) != 0 {
        return false
      }
      // Failed to load dependency info: rebuild to regenerate it.
      // LoadDeps() did EXPLAIN() already, no need to do it here.
      dirty = edge.deps_missing_ = true
    }
  }

  // Visit all inputs; we're dirty if any of the inputs are dirty.
  most_recent_input := nil
  for i := edge.inputs_.begin(); i != edge.inputs_.end(); i++ {
    // Visit this input.
    if !RecomputeDirty(*i, stack, err) {
      return false
    }

    // If an input is not ready, neither are our outputs.
    if Edge* in_edge = (*i).in_edge() {
      if !in_edge.outputs_ready_ {
        edge.outputs_ready_ = false
      }
    }

    if !edge.is_order_only(i - edge.inputs_.begin()) {
      // If a regular input is dirty (or missing), we're dirty.
      // Otherwise consider mtime.
      if (*i).dirty() {
        EXPLAIN("%s is dirty", (*i).path())
        dirty = true
      } else {
        if !most_recent_input || (*i).mtime() > most_recent_input.mtime() {
          most_recent_input = *i
        }
      }
    }
  }

  // We may also be dirty due to output state: missing outputs, out of
  // date outputs, etc.  Visit all outputs and determine whether they're dirty.
  if dirty == nil {
    if !RecomputeOutputsDirty(edge, most_recent_input, &dirty, err) {
      return false
    }
  }

  // Finally, visit each output and update their dirty state if necessary.
  for o := edge.outputs_.begin(); o != edge.outputs_.end(); o++ {
    if dirty != nil {
      (*o).MarkDirty()
    }
  }

  // If an edge is dirty, its outputs are normally not ready.  (It's
  // possible to be clean but still not be ready in the presence of
  // order-only inputs.)
  // But phony edges with no inputs have nothing to do, so are always
  // ready.
  if dirty && !(edge.is_phony() && edge.inputs_.empty()) {
    edge.outputs_ready_ = false
  }

  // Mark the edge as finished during this walk now that it will no longer
  // be in the call stack.
  edge.mark_ = Edge::VisitDone
  if !stack.back() == node { panic("oops") }
  stack.pop_back()

  return true
}

func (d *DependencyScan) VerifyDAG(node *Node, stack *vector<Node*>, err *string) bool {
  edge := node.in_edge()
  if !edge != nil { panic("oops") }

  // If we have no temporary mark on the edge then we do not yet have a cycle.
  if edge.mark_ != Edge::VisitInStack {
    return true
  }

  // We have this edge earlier in the call stack.  Find it.
  start := stack.begin()
  while start != stack.end() && (*start).in_edge() != edge {
    start++
  }
  if !start != stack.end() { panic("oops") }

  // Make the cycle clear by reporting its start as the node at its end
  // instead of some other output of the starting edge.  For example,
  // running 'ninja b' on
  //   build a b: cat c
  //   build c: cat a
  // should report a -> c -> a instead of b -> c -> a.
  *start = node

  // Construct the error message rejecting the cycle.
  *err = "dependency cycle: "
  for i := start; i != stack.end(); i++ {
    err.append((*i).path())
    err.append(" . ")
  }
  err.append((*start).path())

  if (start + 1) == stack.end() && edge.maybe_phonycycle_diagnostic() {
    // The manifest parser would have filtered out the self-referencing
    // input if it were not configured to allow the error.
    err.append(" [-w phonycycle=err]")
  }

  return false
}

// Recompute whether any output of the edge is dirty, if so sets |*dirty|.
// Returns false on failure.
func (d *DependencyScan) RecomputeOutputsDirty(edge *Edge, most_recent_input *Node, outputs_dirty *bool, err *string) bool {
  string command = edge.EvaluateCommand(/*incl_rsp_file=*/true)
  for o := edge.outputs_.begin(); o != edge.outputs_.end(); o++ {
    if RecomputeOutputDirty(edge, most_recent_input, command, *o) {
      *outputs_dirty = true
      return true
    }
  }
  return true
}

// Recompute whether a given single output should be marked dirty.
// Returns true if so.
func (d *DependencyScan) RecomputeOutputDirty(edge *Edge, most_recent_input *Node, command string, output *Node) bool {
  if edge.is_phony() {
    // Phony edges don't write any output.  Outputs are only dirty if
    // there are no inputs and we're missing the output.
    if edge.inputs_.empty() && !output.exists() {
      EXPLAIN("output %s of phony edge with no inputs doesn't exist", output.path())
      return true
    }

    // Update the mtime with the newest input. Dependents can thus call mtime()
    // on the fake node and get the latest mtime of the dependencies
    if most_recent_input {
      output.UpdatePhonyMtime(most_recent_input.mtime())
    }

    // Phony edges are clean, nothing to do
    return false
  }

  entry := 0

  // Dirty if we're missing the output.
  if !output.exists() {
    EXPLAIN("output %s doesn't exist", output.path())
    return true
  }

  // Dirty if the output is older than the input.
  if most_recent_input && output.mtime() < most_recent_input.mtime() {
    output_mtime := output.mtime()

    // If this is a restat rule, we may have cleaned the output with a restat
    // rule in a previous run and stored the most recent input mtime in the
    // build log.  Use that mtime instead, so that the file will only be
    // considered dirty if an input was modified since the previous run.
    used_restat := false
    if edge.GetBindingBool("restat") && build_log() && (entry = build_log().LookupByOutput(output.path())) {
      output_mtime = entry.mtime
      used_restat = true
    }

    if output_mtime < most_recent_input.mtime() {
      EXPLAIN("%soutput %s older than most recent input %s " "(%" PRId64 " vs %" PRId64 ")", used_restat ? "restat of " : "", output.path(), most_recent_input.path(), output_mtime, most_recent_input.mtime())
      return true
    }
  }

  if build_log() {
    bool generator = edge.GetBindingBool("generator")
    if entry || (entry = build_log().LookupByOutput(output.path())) {
      if !generator && BuildLog::LogEntry::HashCommand(command) != entry.command_hash {
        // May also be dirty due to the command changing since the last build.
        // But if this is a generator rule, the command changing does not make us
        // dirty.
        EXPLAIN("command line changed for %s", output.path())
        return true
      }
      if most_recent_input && entry.mtime < most_recent_input.mtime() {
        // May also be dirty due to the mtime in the log being older than the
        // mtime of the most recent input.  This can occur even when the mtime
        // on disk is newer if a previous run wrote to the output file but
        // exited with an error or was interrupted.
        EXPLAIN("recorded mtime of %s older than most recent input %s (%" PRId64 " vs %" PRId64 ")", output.path(), most_recent_input.path(), entry.mtime, most_recent_input.mtime())
        return true
      }
    }
    if !entry && !generator {
      EXPLAIN("command line not found in log for %s", output.path())
      return true
    }
  }

  return false
}

// Load a dyndep file from the given node's path and update the
// build graph with the new information.  One overload accepts
// a caller-owned 'DyndepFile' object in which to store the
// information loaded from the dyndep file.
func (d *DependencyScan) LoadDyndeps(node *Node, err *string) bool {
  return dyndep_loader_.LoadDyndeps(node, err)
}

// Load a dyndep file from the given node's path and update the
// build graph with the new information.  One overload accepts
// a caller-owned 'DyndepFile' object in which to store the
// information loaded from the dyndep file.
func (d *DependencyScan) LoadDyndeps(node *Node, ddf *DyndepFile, err *string) bool {
  return dyndep_loader_.LoadDyndeps(node, ddf, err)
}

// Return true if all inputs' in-edges are ready.
func (e *Edge) AllInputsReady() bool {
  for i := inputs_.begin(); i != inputs_.end(); i++ {
    if (*i).in_edge() && !(*i).in_edge().outputs_ready() {
      return false
    }
  }
  return true
}

// An Env for an Edge, providing $in and $out.
type EdgeEnv struct {

  EdgeEnv(const Edge* const edge, const EscapeKind escape)
      : edge_(edge), escape_in_out_(escape), recursive_(false) {}

  lookups_ []string
  edge_ Edge* const
  escape_in_out_ EscapeKind
  recursive_ bool
}
  enum EscapeKind { kShellEscape, kDoNotEscape }

func (e *EdgeEnv) LookupVariable(var string) string {
  if var == "in" || var == "in_newline" {
    int explicit_deps_count = edge_.inputs_.size() - edge_.implicit_deps_ -
      edge_.order_only_deps_
