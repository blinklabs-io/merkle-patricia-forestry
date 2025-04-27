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

func prependTag(tag byte, encodedValue []byte) []byte {
	// d8 is the indicator for a 1-byte tag value
	return append([]byte{0xd8, tag}, encodedValue...)
}

func (s *ProofStep) MarshalCBOR() ([]byte, error) {
	var valueToTag any
	var tagValue byte
	var err error

	switch s.stepType {
	case ProofStepTypeBranch:
		tagValue = 121 // 0x79
		tmpNeighbors := []byte{}
		for _, neighbor := range s.neighbors {
			if neighbor == (Hash{}) {
				tmpNeighbors = append(tmpNeighbors, NullHash[:]...)
			} else {
				tmpNeighbors = append(tmpNeighbors, neighbor.Bytes()...)
			}
		}
		if len(tmpNeighbors) != 128 {
			return nil, fmt.Errorf("proof step branch: expected 128 bytes for neighbors, got %d", len(tmpNeighbors))
		}
		valueToTag = cbor.IndefLengthList{
			uint64(s.prefixLength),
			cbor.IndefLengthByteString{
				tmpNeighbors[0:64],
				tmpNeighbors[64:128],
			},
		}

	case ProofStepTypeFork:
		tagValue = 122
		prefixBytes := nibblesToBytes(s.neighbor.prefix)

		// The inner value (neighbor details) is ALSO tagged with 121
		neighborData := cbor.IndefLengthList{
			uint64(s.neighbor.nibble),
			prefixBytes,
			s.neighbor.root,
		}
		// encode inner list first
		encodedNeighborData, err := cbor.Encode(&neighborData)
		if err != nil {
			return nil, fmt.Errorf("failed to encode fork neighbor data: %w", err)
		}

		taggedNeighborBytes := prependTag(121, encodedNeighborData)

		// The value for the outer tag (122) is a list containing skip and the already-tagged neighbor bytes
		// Use RawMessage to include the pre-tagged bytes without re-encoding
		valueToTag = cbor.IndefLengthList{
			uint64(s.prefixLength),
			cbor.RawMessage(taggedNeighborBytes),
		}

	case ProofStepTypeLeaf:
		tagValue = 123
		keyBytes := nibblesToBytes(s.neighbor.key)
		valueToTag = cbor.IndefLengthList{
			uint64(s.prefixLength),
			keyBytes,
			s.neighbor.value,
		}

	default:
		return nil, errors.New("unknown proof step type")
	}

	encodedValue, err := cbor.Encode(&valueToTag)
	if err != nil {
		return nil, fmt.Errorf("failed to encode proof step value for tag %d: %w", tagValue, err)
	}

	return prependTag(tagValue, encodedValue), nil
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
