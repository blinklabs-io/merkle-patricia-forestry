// Copyright 2025 Blink Labs Software
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

package mpf

import (
	"errors"
	"fmt"
	"slices"

	"github.com/blinklabs-io/gouroboros/cbor"
)

type ProofStepType int

const (
	ProofStepTypeLeaf   ProofStepType = 1
	ProofStepTypeFork   ProofStepType = 2
	ProofStepTypeBranch ProofStepType = 3
)

func (p ProofStepType) String() string {
	switch p {
	case ProofStepTypeLeaf:
		return "leaf"
	case ProofStepTypeFork:
		return "fork"
	case ProofStepTypeBranch:
		return "branch"
	default:
		return "unknown"
	}
}

type Proof struct {
	path  []Nibble
	value []byte
	steps []ProofStep
}

func newProof(path []Nibble, value []byte) *Proof {
	p := &Proof{}
	if path != nil {
		p.path = append(p.path, path...)
	}
	if value != nil {
		p.value = append(p.value, value...)
	}
	return p
}

func (p *Proof) Rewind(targetIdx int, prefixLen int, neighbors []Node) {
	nonEmptyNeighbors := []Node{}
	var nonEmptyNeighborIdx int
	for idx, neighbor := range neighbors {
		if neighbor == nil {
			continue
		}
		if idx == targetIdx {
			continue
		}
		nonEmptyNeighbors = append(nonEmptyNeighbors, neighbor)
		nonEmptyNeighborIdx = idx
	}
	if len(nonEmptyNeighbors) == 1 {
		neighbor := nonEmptyNeighbors[0]
		switch n := neighbor.(type) {
		case *Leaf:
			step := ProofStep{
				stepType:     ProofStepTypeLeaf,
				prefixLength: prefixLen,
				neighbor: ProofStepNeighbor{
					key:   keyToPath(n.key),
					value: HashValue(n.value),
				},
			}
			p.steps = slices.Insert(p.steps, 0, step)
		case *Branch:
			step := ProofStep{
				stepType:     ProofStepTypeFork,
				prefixLength: prefixLen,
				neighbor: ProofStepNeighbor{
					prefix: n.prefix,
					nibble: Nibble(nonEmptyNeighborIdx),
					root:   merkleRoot(n.children[:]),
				},
			}
			p.steps = slices.Insert(p.steps, 0, step)
		default:
			panic(
				fmt.Sprintf(
					"unknown Node type %T...this should never happen",
					neighbor,
				),
			)
		}
	} else {
		step := ProofStep{
			stepType:     ProofStepTypeBranch,
			prefixLength: prefixLen,
			neighbors:    merkleProof(neighbors, targetIdx),
		}
		p.steps = slices.Insert(p.steps, 0, step)
	}
}

func (p *Proof) MarshalCBOR() ([]byte, error) {
	tmpSteps := make([]any, 0, len(p.steps))
	for _, step := range p.steps {
		tmpSteps = append(tmpSteps, &step)
	}
	tmpData := cbor.IndefLengthList(tmpSteps)
	return cbor.Encode(&tmpData)
}

type ProofStep struct {
	stepType     ProofStepType
	prefixLength int
	neighbors    []Hash
	neighbor     ProofStepNeighbor
}

func (s *ProofStep) MarshalCBOR() ([]byte, error) {
	switch s.stepType {
	case ProofStepTypeBranch:
		tmpNeighbors := []byte{}
		for _, neighbor := range s.neighbors {
			tmpNeighbors = append(tmpNeighbors, neighbor.Bytes()...)
		}
		tmpData := cbor.NewConstructor(
			0,
			cbor.IndefLengthList{
				s.prefixLength,
				cbor.IndefLengthByteString{
					tmpNeighbors[0:64],
					tmpNeighbors[64:],
				},
			},
		)
		return cbor.Encode(&tmpData)

	case ProofStepTypeFork:
		prefixBytes := nibblesToBytes(s.neighbor.prefix)
		tmpData := cbor.NewConstructor(
			1,
			cbor.IndefLengthList{
				s.prefixLength,
				cbor.NewConstructor(
					0,
					cbor.IndefLengthList{
						int(s.neighbor.nibble),
						prefixBytes,
						s.neighbor.root,
					},
				),
			},
		)
		return cbor.Encode(&tmpData)

	case ProofStepTypeLeaf:
		tmpData := cbor.NewConstructor(
			2,
			cbor.IndefLengthList{
				s.prefixLength,
				nibblesToBytes(s.neighbor.key),
				s.neighbor.value,
			},
		)
		return cbor.Encode(&tmpData)

	default:
		return nil, errors.New("unknown proof step type")
	}
}

type ProofStepNeighbor struct {
	key    []Nibble
	value  Hash
	prefix []Nibble
	nibble Nibble
	root   Hash
}

func merkleProof(nodes []Node, myIdx int) []Hash {
	var ret []Hash
	pivot := 8
	n := 8
	for n >= 1 {
		if myIdx < pivot {
			ret = append(
				ret,
				merkleRoot(
					nodes[pivot:pivot+n],
				),
			)
			pivot -= (n >> 1)
		} else {
			ret = append(
				ret,
				merkleRoot(
					nodes[pivot-n:pivot],
				),
			)
			pivot += (n >> 1)
		}
		n = n >> 1
	}
	return ret
}
