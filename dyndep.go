// Copyright 2015 Google Inc. All Rights Reserved.
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

// Store dynamically-discovered dependency information for one edge.
type Dyndeps struct {
	used            bool
	restat          bool
	implicitInputs  []*Node
	implicitOutputs []*Node
}

func NewDyndeps() *Dyndeps {
	return &Dyndeps{}
}

func (d *Dyndeps) String() string {
	out := "Dyndeps{in:"
	for i, n := range d.implicitInputs {
		if i != 0 {
			out += ","
		}
		out += n.Path
	}
	out += "; out:"
	for i, n := range d.implicitOutputs {
		if i != 0 {
			out += ","
		}
		out += n.Path
	}
	return out + "}"
}

// Store data loaded from one dyndep file.  Map from an edge
// to its dynamically-discovered dependency information.
// This is a struct rather than a typedef so that we can
// forward-declare it in other headers.
type DyndepFile map[*Edge]*Dyndeps

// DyndepLoader loads dynamically discovered dependencies, as
// referenced via the "dyndep" attribute in build files.
type DyndepLoader struct {
	state *State
	di    DiskInterface
}

func NewDyndepLoader(state *State, di DiskInterface) DyndepLoader {
	return DyndepLoader{
		state: state,
		di:    di,
	}
}

// Load a dyndep file from the given node's path and update the
// build graph with the new information.  One overload accepts
// a caller-owned 'DyndepFile' object in which to store the
// information loaded from the dyndep file.
func (d *DyndepLoader) LoadDyndeps(node *Node, ddf DyndepFile, err *string) bool {
	// We are loading the dyndep file now so it is no longer pending.
	node.DyndepPending = false

	// Load the dyndep information from the file.
	Explain("loading dyndep file '%s'", node.Path)
	if !d.LoadDyndepFile(node, ddf, err) {
		return false
	}

	// Update each edge that specified this node as its dyndep binding.
	outEdges := node.OutEdges
	for _, oe := range outEdges {
		edge := oe
		if edge.Dyndep != node {
			continue
		}

		ddi, ok := ddf[edge]
		if !ok {
			*err = ("'" + edge.Outputs[0].Path + "' not mentioned in its dyndep file '" + node.Path + "'")
			return false
		}

		ddi.used = true
		dyndeps := ddi
		if !d.UpdateEdge(edge, dyndeps, err) {
			return false
		}
	}

	// Reject extra outputs in dyndep file.
	for edge, oe := range ddf {
		if !oe.used {
			*err = ("dyndep file '" + node.Path + "' mentions output '" + edge.Outputs[0].Path + "' whose build statement does not have a dyndep binding for the file")
			return false
		}
	}

	return true
}

func (d *DyndepLoader) UpdateEdge(edge *Edge, dyndeps *Dyndeps, err *string) bool {
	// Add dyndep-discovered bindings to the edge.
	// We know the edge already has its own binding
	// scope because it has a "dyndep" binding.
	if dyndeps.restat {
		edge.Env.Bindings["restat"] = "1"
	}

	// Add the dyndep-discovered outputs to the edge.
	edge.Outputs = append(edge.Outputs, dyndeps.implicitOutputs...)
	edge.ImplicitOuts += int32(len(dyndeps.implicitOutputs))

	// Add this edge as incoming to each new output.
	for _, i := range dyndeps.implicitOutputs {
		if oldInEdge := i.InEdge; oldInEdge != nil {
			// This node already has an edge producing it.  Fail with an error
			// unless the edge was generated by ImplicitDepLoader, in which
			// case we can replace it with the now-known real producer.
			if !oldInEdge.GeneratedByDepLoader {
				*err = "multiple rules generate " + i.Path
				return false
			}
			oldInEdge.Outputs = nil
		}
		i.InEdge = edge
	}

	// Add the dyndep-discovered inputs to the edge.
	old := edge.Inputs
	offset := len(edge.Inputs) - int(edge.OrderOnlyDeps)
	edge.Inputs = make([]*Node, len(edge.Inputs)+len(dyndeps.implicitInputs))
	copy(edge.Inputs, old[:offset])
	copy(edge.Inputs[offset:], dyndeps.implicitInputs)
	copy(edge.Inputs[offset+len(dyndeps.implicitInputs):], old[offset:])
	edge.ImplicitDeps += int32(len(dyndeps.implicitInputs))

	// Add this edge as outgoing from each new input.
	for _, n := range dyndeps.implicitInputs {
		n.OutEdges = append(n.OutEdges, edge)
	}
	return true
}

func (d *DyndepLoader) LoadDyndepFile(file *Node, ddf DyndepFile, err *string) bool {
	parser := NewDyndepParser(d.state, d.di, ddf)
	return parser.Load(file.Path, err, nil)
}
