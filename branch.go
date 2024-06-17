// Copyright 2024 Blink Labs Software
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

import "fmt"

type ChildNode interface {
	isChildNode()
	Hash() Hash
}

type Branch struct {
	hash     Hash
	prefix   []Nibble
	children [16]ChildNode
	size     int
}

func newBranch(prefix []Nibble) *Branch {
	b := &Branch{}
	if prefix != nil {
		b.prefix = append(b.prefix, prefix...)
	}
	return b
}

func (b Branch) isChildNode() {}

func (b *Branch) Hash() Hash {
	return b.hash
}

func (b *Branch) updateHash() {
	var tmpVal []byte
	// Append prefix
	for _, nibble := range b.prefix {
		tmpVal = append(tmpVal, byte(nibble))
	}
	// Calculate merkle root for children and append
	childrenHash := b.merkleRootChildren()
	tmpVal = append(tmpVal, childrenHash.Bytes()...)
	// Calculate hash
	b.hash = HashValue(tmpVal)
}

func (b *Branch) get(key []Nibble) ([]byte, error) {
	childIdx := int(key[0])
	if b.children[childIdx] == nil {
		return nil, ErrKeyNotExist
	}
	existingChild := b.children[childIdx]
	switch v := existingChild.(type) {
	case *Leaf:
		if string(key[1:]) == string(v.suffix) {
			return v.value, nil
		}
		return nil, ErrKeyNotExist
	case *Branch:
		return v.get(key[1:])
	default:
		panic(
			fmt.Sprintf(
				"unknown ChildNode type %T...this should never happen",
				existingChild,
			),
		)
	}
}

func (b *Branch) insert(key []Nibble, val []byte) {
	childIdx := int(key[0])
	if b.children[childIdx] == nil {
		b.children[childIdx] = newLeaf(
			key[1:],
			val,
		)
		b.size++
		b.updateHash()
	} else {
		existingChild := b.children[childIdx]
		switch v := existingChild.(type) {
		case *Leaf:
			tmpPrefix := commonPrefix(key[1:], v.suffix)
			tmpBranch := newBranch(tmpPrefix)
			// Add original leaf node to new branch
			tmpBranch.children[childIdx] = v
			tmpBranch.size++
			// Insert new value to new branch
			tmpBranch.insert(
				key[1:],
				val,
			)
			// Replace existing leaf node with new branch node
			b.children[childIdx] = tmpBranch
		case *Branch:
			v.insert(
				key[1:],
				val,
			)
		default:
			panic(
				fmt.Sprintf(
					"unknown ChildNode type %T...this should never happen",
					existingChild,
				),
			)
		}
		b.size++
		b.updateHash()
	}
}

func (b *Branch) delete(key []Nibble) error {
	childIdx := int(key[0])
	if b.children[childIdx] == nil {
		return ErrKeyNotExist
	}
	existingChild := b.children[childIdx]
	switch v := existingChild.(type) {
	case *Leaf:
		if string(v.suffix) != string(key[1:]) {
			return ErrKeyNotExist
		}
		b.children[childIdx] = nil
		b.updateHash()
	case *Branch:
		err := v.delete(
			key[1:],
		)
		if err != nil {
			return err
		}
		// Merge branch with only one child
		if v.size == 1 {
			// Find non-nil child entry
			for _, tmpChild := range v.children {
				if tmpChild != nil {
					// Prepend branch prefix to child prefix
					switch v2 := tmpChild.(type) {
					case *Leaf:
						v2.suffix = append(v.prefix, v2.suffix...)
					case *Branch:
						v2.prefix = append(v.prefix, v2.prefix...)
					}
					b.children[childIdx] = tmpChild
					break
				}
			}
		}
		b.updateHash()
	default:
		panic(
			fmt.Sprintf(
				"unknown ChildNode type %T...this should never happen",
				existingChild,
			),
		)
	}
	b.size--
	return nil
}

func (b *Branch) merkleRootChildren() Hash {
	// Gather child node hashes
	tmpHashes := make([]Hash, 0, len(b.children))
	for _, child := range b.children {
		tmpHash := NullHash
		if child != nil {
			tmpHash = child.Hash()
		}
		tmpHashes = append(tmpHashes, tmpHash)
	}
	// Concat and hash child hashes in pairs, repeating until only a single hash remains
	for len(tmpHashes) > 1 {
		newTmpHashes := make([]Hash, 0, len(tmpHashes)/2)
		for i := 0; i < len(tmpHashes); i = i + 2 {
			tmpVal := append(
				tmpHashes[i].Bytes(),
				tmpHashes[i+1].Bytes()...,
			)
			tmpHash := HashValue(tmpVal)
			newTmpHashes = append(newTmpHashes, tmpHash)
		}
		tmpHashes = newTmpHashes
	}
	return tmpHashes[0]
}

func commonPrefix(prefixA []Nibble, prefixB []Nibble) []Nibble {
	var ret []Nibble
	tmpLen := len(prefixA)
	if len(prefixB) < tmpLen {
		tmpLen = len(prefixB)
	}
	for i := 0; i < tmpLen; i++ {
		if prefixA[i] != prefixB[i] {
			break
		}
		ret = append(ret, prefixA[i])
	}
	return ret
}
