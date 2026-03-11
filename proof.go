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

func (p *Proof) UnmarshalCBOR(data []byte) error {
	*p = Proof{}
	var tmpSteps []ProofStep
	bytesRead, err := cbor.Decode(data, &tmpSteps)
	if err != nil {
		return err
	}
	if bytesRead != len(data) {
		return fmt.Errorf(
			"trailing data after proof: %d bytes",
			len(data)-bytesRead,
		)
	}
	p.steps = tmpSteps
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
	var constructor cbor.ConstructorDecoder
	bytesRead, err := cbor.Decode(data, &constructor)
	if err != nil {
		return err
	}
	if bytesRead != len(data) {
		return fmt.Errorf(
			"trailing data after proof step: %d bytes",
			len(data)-bytesRead,
		)
	}
	switch constructor.Tag() {
	case 0:
		if err := s.unmarshalBranchStep(constructor); err != nil {
			return fmt.Errorf("decode branch proof step: %w", err)
		}
	case 1:
		if err := s.unmarshalForkStep(constructor); err != nil {
			return fmt.Errorf("decode fork proof step: %w", err)
		}
	case 2:
		if err := s.unmarshalLeafStep(constructor); err != nil {
			return fmt.Errorf("decode leaf proof step: %w", err)
		}
	default:
		return fmt.Errorf("unknown proof step constructor: %d", constructor.Tag())
	}
	return nil
}

func (s *ProofStep) MarshalCBOR() ([]byte, error) {
	switch s.stepType {
	case ProofStepTypeBranch:
		tmpNeighbors := make([]byte, 0, len(s.neighbors)*HashSize)
		for _, neighbor := range s.neighbors {
			tmpNeighbors = append(tmpNeighbors, neighbor.Bytes()...)
		}
		tmpData := cbor.NewConstructorEncoder(
			0,
			cbor.IndefLengthList{
				s.prefixLength,
				cbor.IndefLengthByteString{
					tmpNeighbors[0:64],
					tmpNeighbors[64:],
				},
			},
		)
		return cbor.Encode(tmpData)

	case ProofStepTypeFork:
		prefixBytes := nibblesToIndividualBytes(s.neighbor.prefix)
		tmpData := cbor.NewConstructorEncoder(
			1,
			cbor.IndefLengthList{
				s.prefixLength,
				cbor.NewConstructorEncoder(
					0,
					cbor.IndefLengthList{
						int(s.neighbor.nibble),
						prefixBytes,
						s.neighbor.root,
					},
				),
			},
		)
		return cbor.Encode(tmpData)

	case ProofStepTypeLeaf:
		tmpData := cbor.NewConstructorEncoder(
			2,
			cbor.IndefLengthList{
				s.prefixLength,
				nibblesToBytes(s.neighbor.key),
				s.neighbor.value,
			},
		)
		return cbor.Encode(tmpData)

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

const branchProofNeighborCount = 4

func (s *ProofStep) unmarshalBranchStep(
	constructor cbor.ConstructorDecoder,
) error {
	var fields []cbor.RawMessage
	if err := constructor.DecodeFields(&fields); err != nil {
		return err
	}
	if len(fields) != 2 {
		return errors.New("missing fields")
	}
	prefixLen, err := decodeNonNegativeInt(fields[0])
	if err != nil {
		return fmt.Errorf("invalid prefix length: %w", err)
	}
	neighborsBytes, err := decodeBytes(fields[1])
	if err != nil {
		return fmt.Errorf("invalid neighbors: %w", err)
	}
	expectedNeighborBytes := branchProofNeighborCount * HashSize
	if len(neighborsBytes) != expectedNeighborBytes {
		return fmt.Errorf(
			"incorrect branch neighbor data length: got %d, want %d",
			len(neighborsBytes),
			expectedNeighborBytes,
		)
	}
	neighbors := make([]Hash, 0, branchProofNeighborCount)
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
	return nil
}

func (s *ProofStep) unmarshalForkStep(constructor cbor.ConstructorDecoder) error {
	var fields []cbor.RawMessage
	if err := constructor.DecodeFields(&fields); err != nil {
		return err
	}
	if len(fields) != 2 {
		return errors.New("missing fields")
	}
	prefixLen, err := decodeNonNegativeInt(fields[0])
	if err != nil {
		return fmt.Errorf("invalid prefix length: %w", err)
	}
	var neighborConstructor cbor.ConstructorDecoder
	if err := decodeExact(fields[1], &neighborConstructor); err != nil {
		return fmt.Errorf("invalid neighbor constructor: %w", err)
	}
	if neighborConstructor.Tag() != 0 {
		return fmt.Errorf(
			"unexpected fork neighbor constructor: %d",
			neighborConstructor.Tag(),
		)
	}
	var neighborFields []cbor.RawMessage
	if err := neighborConstructor.DecodeFields(&neighborFields); err != nil {
		return err
	}
	if len(neighborFields) != 3 {
		return errors.New("fork neighbor missing fields")
	}
	neighborIdx, err := decodeNonNegativeInt(neighborFields[0])
	if err != nil {
		return fmt.Errorf("invalid fork neighbor index: %w", err)
	}
	if neighborIdx > 0xf {
		return fmt.Errorf("fork neighbor index out of range: %d", neighborIdx)
	}
	prefixBytes, err := decodeBytes(neighborFields[1])
	if err != nil {
		return fmt.Errorf("invalid fork neighbor prefix: %w", err)
	}
	rootBytes, err := decodeBytes(neighborFields[2])
	if err != nil {
		return fmt.Errorf("invalid fork neighbor root: %w", err)
	}
	neighborRoot, err := hashFromBytes(rootBytes)
	if err != nil {
		return fmt.Errorf("invalid fork neighbor root: %w", err)
	}
	neighborNibble, err := nibbleFromInt(neighborIdx)
	if err != nil {
		return fmt.Errorf("invalid fork neighbor index: %w", err)
	}
	s.stepType = ProofStepTypeFork
	s.prefixLength = prefixLen
	neighborPrefix, err := individualBytesToNibbles(prefixBytes)
	if err != nil {
		return fmt.Errorf("invalid fork neighbor prefix: %w", err)
	}
	s.neighbor = ProofStepNeighbor{
		prefix: neighborPrefix,
		nibble: neighborNibble,
		root:   neighborRoot,
	}
	return nil
}

func (s *ProofStep) unmarshalLeafStep(constructor cbor.ConstructorDecoder) error {
	var fields []cbor.RawMessage
	if err := constructor.DecodeFields(&fields); err != nil {
		return err
	}
	if len(fields) != 3 {
		return errors.New("missing fields")
	}
	prefixLen, err := decodeNonNegativeInt(fields[0])
	if err != nil {
		return fmt.Errorf("invalid prefix length: %w", err)
	}
	keyBytes, err := decodeBytes(fields[1])
	if err != nil {
		return fmt.Errorf("invalid key: %w", err)
	}
	valueBytes, err := decodeBytes(fields[2])
	if err != nil {
		return fmt.Errorf("invalid value: %w", err)
	}
	leafValue, err := hashFromBytes(valueBytes)
	if err != nil {
		return fmt.Errorf("invalid value: %w", err)
	}
	s.stepType = ProofStepTypeLeaf
	s.prefixLength = prefixLen
	s.neighbor = ProofStepNeighbor{
		key:   bytesToNibbles(keyBytes),
		value: leafValue,
	}
	return nil
}

func decodeExact(data []byte, dest any) error {
	bytesRead, err := cbor.Decode(data, dest)
	if err != nil {
		return err
	}
	if bytesRead != len(data) {
		return fmt.Errorf("trailing field data: %d bytes", len(data)-bytesRead)
	}
	return nil
}

func decodeNonNegativeInt(data []byte) (int, error) {
	var value uint64
	if err := decodeExact(data, &value); err == nil {
		if value > uint64(^uint(0)>>1) {
			return 0, fmt.Errorf("value out of range: %d", value)
		}
		return int(value), nil
	}
	var signedValue int64
	if err := decodeExact(data, &signedValue); err == nil {
		if signedValue < 0 {
			return 0, fmt.Errorf("negative value: %d", signedValue)
		}
		if uint64(signedValue) > uint64(^uint(0)>>1) {
			return 0, fmt.Errorf("value out of range: %d", signedValue)
		}
		return int(signedValue), nil
	}
	return 0, errors.New("expected integer")
}

func decodeBytes(data []byte) ([]byte, error) {
	var value []byte
	if err := decodeExact(data, &value); err != nil {
		return nil, err
	}
	return value, nil
}

func nibbleFromInt(value int) (Nibble, error) {
	if value < 0 || value > 0xf {
		return 0, fmt.Errorf("nibble out of range: %d", value)
	}
	return Nibble(uint8(value)), nil
}

func hashFromBytes(data []byte) (Hash, error) {
	if len(data) != HashSize {
		return Hash{}, fmt.Errorf("expected %d bytes for hash, got %d", HashSize, len(data))
	}
	var ret Hash
	copy(ret[:], data)
	return ret, nil
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
