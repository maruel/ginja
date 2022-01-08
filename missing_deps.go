// Copyright 2019 Google Inc. All Rights Reserved.
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

import "fmt"

// MissingDependencyScannerDelegate is a callback when a missing dependency is
// found.
type MissingDependencyScannerDelegate interface {
	OnMissingDep(node *Node, path string, generator *Rule)
}

// MissingDependencyScanner is a scanner for missing dependencies.
type MissingDependencyScanner struct {
	delegate            MissingDependencyScannerDelegate
	depsLog             *DepsLog
	state               *State
	di                  DiskInterface
	seen                map[*Node]struct{}
	nodesMissingDeps    map[*Node]struct{}
	generatedNodes      map[*Node]struct{}
	generatorRules      map[*Rule]struct{}
	missingDepPathCount int

	adjacencyMap map[*Edge]map[*Edge]bool
}

// HadMissingDeps return true if there were any missing dependencies found.
func (m *MissingDependencyScanner) HadMissingDeps() bool {
	return len(m.nodesMissingDeps) != 0
}

// ImplicitDepLoader variant that stores dep nodes into the given output
// without updating graph deps like the base loader does.
type nodeStoringImplicitDepLoader struct {
	implicitDepLoader
	depNodesOutput []*Node
}

func newNodeStoringImplicitDepLoader(state *State, depsLog *DepsLog, di DiskInterface, depNodesOutput []*Node) nodeStoringImplicitDepLoader {
	return nodeStoringImplicitDepLoader{
		implicitDepLoader: newImplicitDepLoader(state, depsLog, di),
		depNodesOutput:    depNodesOutput,
	}
}

func (n *nodeStoringImplicitDepLoader) ProcessDepfileDeps(edge *Edge, depfileIns []string, err *string) bool {
	for _, i := range depfileIns {
		node := n.state.GetNode(CanonicalizePathBits(i))
		n.depNodesOutput = append(n.depNodesOutput, node)
	}
	return true
}

//

// NewMissingDependencyScanner returns an initialized MissingDependencyScanner.
func NewMissingDependencyScanner(delegate MissingDependencyScannerDelegate, depsLog *DepsLog, state *State, di DiskInterface) MissingDependencyScanner {
	return MissingDependencyScanner{
		delegate:         delegate,
		depsLog:          depsLog,
		state:            state,
		di:               di,
		seen:             map[*Node]struct{}{},
		nodesMissingDeps: map[*Node]struct{}{},
		generatedNodes:   map[*Node]struct{}{},
		generatorRules:   map[*Rule]struct{}{},
		adjacencyMap:     map[*Edge]map[*Edge]bool{},
	}
}

// ProcessNode does something?
//
// TODO(maruel): Figure out.
func (m *MissingDependencyScanner) ProcessNode(node *Node) {
	if node == nil {
		return
	}
	edge := node.InEdge
	if edge == nil {
		return
	}
	if _, ok := m.seen[node]; ok {
		return
	}
	m.seen[node] = struct{}{}

	for _, in := range edge.Inputs {
		m.ProcessNode(in)
	}

	depsType := edge.GetBinding("deps")
	if len(depsType) != 0 {
		deps := m.depsLog.GetDeps(node)
		if deps != nil {
			m.processNodeDeps(node, deps.Nodes)
		}
	} else {
		var depfileDeps []*Node
		depLoader := newNodeStoringImplicitDepLoader(m.state, m.depsLog, m.di, depfileDeps)
		_, _ = depLoader.LoadDeps(edge)
		if len(depfileDeps) != 0 {
			m.processNodeDeps(node, depfileDeps)
		}
	}
}

func (m *MissingDependencyScanner) processNodeDeps(node *Node, depNodes []*Node) {
	edge := node.InEdge
	deplogEdges := map[*Edge]struct{}{}
	for i := 0; i < len(depNodes); i++ {
		deplogNode := depNodes[i]
		// Special exception: A dep on build.ninja can be used to mean "always
		// rebuild this target when the build is reconfigured", but build.ninja is
		// often generated by a configuration tool like cmake or gn. The rest of
		// the build "implicitly" depends on the entire build being reconfigured,
		// so a missing dep path to build.ninja is not an actual missing dependency
		// problem.
		if deplogNode.Path == "build.ninja" {
			return
		}
		deplogEdge := deplogNode.InEdge
		if deplogEdge != nil {
			deplogEdges[deplogEdge] = struct{}{}
		}
	}
	var missingDeps []*Edge
	for de := range deplogEdges {
		if !m.pathExistsBetween(de, edge) {
			missingDeps = append(missingDeps, de)
		}
	}

	if len(missingDeps) != 0 {
		missingDepsRuleNames := map[string]struct{}{}
		for _, ne := range missingDeps {
			if ne == nil {
				panic("M-A")
			}
			for i := 0; i < len(depNodes); i++ {
				if depNodes[i].InEdge == nil {
					panic("M-A")
				}
				if m.delegate == nil {
					panic("M-A")
				}
				if depNodes[i].InEdge == ne {
					m.generatedNodes[depNodes[i]] = struct{}{}
					m.generatorRules[ne.Rule] = struct{}{}
					missingDepsRuleNames[ne.Rule.Name] = struct{}{}
					m.delegate.OnMissingDep(node, depNodes[i].Path, ne.Rule)
				}
			}
		}
		m.missingDepPathCount += len(missingDepsRuleNames)
		m.nodesMissingDeps[node] = struct{}{}
	}
}

// PrintStats prints statistics to stdout.
func (m *MissingDependencyScanner) PrintStats() {
	fmt.Printf("Processed %d nodes.\n", len(m.seen))
	if m.HadMissingDeps() {
		fmt.Printf("Error: There are %d missing dependency paths.\n", m.missingDepPathCount)
		fmt.Printf("%d targets had depfile dependencies on %d distinct generated inputs (from %d rules) without a non-depfile dep path to the generator.\n",
			len(m.nodesMissingDeps), len(m.generatedNodes), len(m.generatorRules))
		fmt.Printf("There might be build flakiness if any of the targets listed above are built alone, or not late enough, in a clean output directory.\n")
	} else {
		fmt.Printf("No missing dependencies on generated files found.\n")
	}
}

func (m *MissingDependencyScanner) pathExistsBetween(from *Edge, to *Edge) bool {
	it, ok := m.adjacencyMap[from]
	if ok {
		innerIt, ok := it[to]
		if ok {
			return innerIt
		}
	} else {
		it = map[*Edge]bool{}
		m.adjacencyMap[from] = it
	}
	found := false
	for i := 0; i < len(to.Inputs); i++ {
		e := to.Inputs[i].InEdge
		if e != nil && (e == from || m.pathExistsBetween(from, e)) {
			found = true
			break
		}
	}
	it[to] = found
	return found
}
