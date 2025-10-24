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
	"math"
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

func (p *Proof) UnmarshalCBOR(data []byte) error {
	var tmpSteps []cbor.RawMessage
	if _, err := cbor.Decode(data, &tmpSteps); err != nil {
		return err
	}
	p.steps = p.steps[:0]
	for idx, stepData := range tmpSteps {
		var tmpStep ProofStep
		if err := tmpStep.UnmarshalCBOR(stepData); err != nil {
			return fmt.Errorf("failed to decode proof step %d: %w", idx, err)
		}
		p.steps = append(p.steps, tmpStep)
	}
	return nil
}

type ProofStep struct {
	stepType     ProofStepType
	prefixLength int
	neighbors    []Hash
	neighbor     ProofStepNeighbor
}

func (s *ProofStep) UnmarshalCBOR(data []byte) error {
	*s = ProofStep{}
	var tmpConstructor cbor.Constructor
	if _, err := cbor.Decode(data, &tmpConstructor); err != nil {
		return err
	}
	fields := tmpConstructor.Fields()
	switch tmpConstructor.Constructor() {
	case 0:
		if len(fields) != 2 {
			return errors.New("branch proof step missing fields")
		}
		prefixLen, err := toInt(fields[0])
		if err != nil {
			return fmt.Errorf("invalid branch prefix length: %w", err)
		}
		neighborsBytes, err := bytesFromField(fields[1])
		if err != nil {
			return fmt.Errorf("invalid branch neighbors type: %w", err)
		}
		if len(neighborsBytes)%HashSize != 0 {
			return fmt.Errorf("branch neighbor data incorrect length: %d", len(neighborsBytes))
		}
		neighbors := make([]Hash, 0, len(neighborsBytes)/HashSize)
		for i := 0; i < len(neighborsBytes); i += HashSize {
			neighborHash, err := hashFromBytes(neighborsBytes[i : i+HashSize])
			if err != nil {
				return err
			}
			neighbors = append(neighbors, neighborHash)
		}
		s.stepType = ProofStepTypeBranch
		s.prefixLength = prefixLen
		s.neighbors = neighbors
	case 1:
		if len(fields) != 2 {
			return errors.New("fork proof step missing fields")
		}
		prefixLen, err := toInt(fields[0])
		if err != nil {
			return fmt.Errorf("invalid fork prefix length: %w", err)
		}
		neighborConstructor, ok := fields[1].(cbor.Constructor)
		if !ok {
			return fmt.Errorf("invalid fork neighbor type: %T", fields[1])
		}
		if neighborConstructor.Constructor() != 0 {
			return fmt.Errorf("unexpected fork neighbor constructor: %d", neighborConstructor.Constructor())
		}
		neighborFields := neighborConstructor.Fields()
		if len(neighborFields) != 3 {
			return errors.New("fork neighbor missing fields")
		}
		neighborIdx, err := toInt(neighborFields[0])
		if err != nil {
			return fmt.Errorf("invalid fork neighbor index: %w", err)
		}
		prefixBytes, err := bytesFromField(neighborFields[1])
		if err != nil {
			return fmt.Errorf("invalid fork neighbor prefix: %w", err)
		}
		rootBytes, err := bytesFromField(neighborFields[2])
		if err != nil {
			return fmt.Errorf("invalid fork neighbor root: %w", err)
		}
		neighborRoot, err := hashFromBytes(rootBytes)
		if err != nil {
			return fmt.Errorf("invalid fork neighbor root: %w", err)
		}
		if neighborIdx < 0 || neighborIdx > 0xf {
			return fmt.Errorf("fork neighbor index out of range: %d", neighborIdx)
		}
		s.stepType = ProofStepTypeFork
		s.prefixLength = prefixLen
		s.neighbor = ProofStepNeighbor{
			prefix: bytesToNibbles(prefixBytes),
			nibble: Nibble(neighborIdx),
			root:   neighborRoot,
		}
	case 2:
		if len(fields) != 3 {
			return errors.New("leaf proof step missing fields")
		}
		prefixLen, err := toInt(fields[0])
		if err != nil {
			return fmt.Errorf("invalid leaf prefix length: %w", err)
		}
		keyBytes, err := bytesFromField(fields[1])
		if err != nil {
			return fmt.Errorf("invalid leaf key: %w", err)
		}
		if len(keyBytes)%2 != 0 {
			return fmt.Errorf("leaf key has odd length: %d", len(keyBytes))
		}
		valueBytes, err := bytesFromField(fields[2])
		if err != nil {
			return fmt.Errorf("invalid leaf value: %w", err)
		}
		leafValue, err := hashFromBytes(valueBytes)
		if err != nil {
			return fmt.Errorf("invalid leaf value: %w", err)
		}
		s.stepType = ProofStepTypeLeaf
		s.prefixLength = prefixLen
		s.neighbor = ProofStepNeighbor{
			key:   bytesToNibbles(keyBytes),
			value: leafValue,
		}
	default:
		return fmt.Errorf("unknown proof step constructor: %d", tmpConstructor.Constructor())
	}
	return nil
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

func toInt(value any) (int, error) {
	switch v := value.(type) {
	case int:
		return v, nil
	case uint:
		if uint64(v) > uint64(math.MaxInt) {
			return 0, fmt.Errorf("value out of range: %d", v)
		}
		// #nosec G115 -- conversion is safe due to explicit range check above
		return int(int64(v)), nil
	case int64:
		if v < 0 {
			return 0, fmt.Errorf("negative value: %d", v)
		}
		if v > int64(math.MaxInt) {
			return 0, fmt.Errorf("value out of range: %d", v)
		}
		return int(v), nil
	case uint64:
		if v > uint64(math.MaxInt) {
			return 0, fmt.Errorf("value out of range: %d", v)
		}
		return int(v), nil
	default:
		return 0, fmt.Errorf("unexpected numeric type: %T", value)
	}
}

func hashFromBytes(data []byte) (Hash, error) {
	if len(data) != HashSize {
		return Hash{}, fmt.Errorf("expected %d bytes for hash, got %d", HashSize, len(data))
	}
	var ret Hash
	copy(ret[:], data)
	return ret, nil
}

func bytesFromField(value any) ([]byte, error) {
	switch v := value.(type) {
	case []byte:
		return v, nil
	case cbor.ByteString:
		return v.Bytes(), nil
	default:
		return nil, fmt.Errorf("unexpected bytestring type: %T", value)
	}
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
