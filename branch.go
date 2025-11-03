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
	"fmt"
	"slices"
	"strconv"
	"strings"
)

type Branch struct {
	hash     Hash
	prefix   []Nibble
	children [16]Node
	size     int
}

func newBranch(prefix []Nibble) *Branch {
	b := &Branch{}
	if prefix != nil {
		b.prefix = append(b.prefix, prefix...)
	}
	return b
}

func (b *Branch) isNode() {}

func (b *Branch) String() string {
	var sb strings.Builder
	sb.WriteString(nibblesToHexString(b.prefix))
	sb.WriteByte(' ')
	sb.WriteByte('#')
	sb.WriteString(b.hash.String()[:10])
	for idx, child := range b.children {
		if child == nil {
			continue
		}
		childStr := child.String()
		childStr = strings.ReplaceAll(childStr, "\n", "\n  ")
		sb.WriteString("\n -")
		sb.WriteString(strconv.FormatInt(int64(idx), 16))
		sb.WriteString(childStr)
	}
	return sb.String()
}

func (b *Branch) Hash() Hash {
	return b.hash
}

func (b *Branch) updateHash() {
	tmpVal := []byte{}
	// Append prefix
	for _, nibble := range b.prefix {
		tmpVal = append(tmpVal, byte(nibble))
	}
	// Calculate merkle root for children and append
	childrenHash := merkleRoot(b.children[:])
	tmpVal = append(tmpVal, childrenHash.Bytes()...)
	// Calculate hash
	b.hash = HashValue(tmpVal)
}

func (b *Branch) get(path []Nibble) ([]byte, error) {
	cmnPrefix := commonPrefix(path, b.prefix)
	if string(cmnPrefix) == string(b.prefix) {
		// Determine path minus the current node prefix
		pathMinusPrefix := path[len(b.prefix):]
		// Determine which child slot the next nibble in the path fits in
		childIdx := int(pathMinusPrefix[0])
		// Determine sub-path for key. We strip off the first nibble, since it's implied by
		// the child slot that it's in
		subPath := pathMinusPrefix[1:]
		if b.children[childIdx] == nil {
			return nil, ErrKeyNotExist
		}
		existingChild := b.children[childIdx]
		switch v := existingChild.(type) {
		case *Leaf:
			if string(subPath) == string(v.suffix) {
				return v.value, nil
			}
			return nil, ErrKeyNotExist
		case *Branch:
			return v.get(subPath)
		default:
			panic(
				fmt.Sprintf(
					"unknown Node type %T...this should never happen",
					existingChild,
				),
			)
		}
	}
	return nil, ErrKeyNotExist
}

func (b *Branch) insert(path []Nibble, key []byte, val []byte) {
	// Determine path minus the current node prefix
	pathMinusPrefix := path[len(b.prefix):]
	// Determine which child slot the next nibble in the path fits in
	childIdx := int(pathMinusPrefix[0])
	// Determine sub-path for key. We strip off the first nibble, since it's implied by
	// the child slot that it's in
	subPath := pathMinusPrefix[1:]
	// Create leaf node and add to appropriate slot if there's not already a node there
	if b.children[childIdx] == nil {
		b.addChild(
			childIdx,
			newLeaf(
				subPath,
				key,
				val,
			),
		)
		return
	}
	existingChild := b.children[childIdx]
	switch v := existingChild.(type) {
	// Existing child node is a leaf. We'll need to replace it with a branch with both
	// the original leaf node and the new leaf node
	case *Leaf:
		// Determine the common prefix nibbles between existing leaf and new leaf node
		tmpPrefix := commonPrefix(subPath, v.suffix)
		// Update value for existing key
		if string(tmpPrefix) == string(v.suffix) {
			v.Set(val)
			b.updateHash()
			return
		}
		// Create a new branch node with the common prefix
		tmpBranch := newBranch(tmpPrefix)
		// Add original leaf node values to new branch
		tmpBranch.insert(
			v.suffix,
			v.key,
			v.value,
		)
		// Insert new value to new branch
		tmpBranch.insert(
			subPath,
			key,
			val,
		)
		// Replace existing leaf node with new branch node
		b.children[childIdx] = tmpBranch
		b.updateHash()

	case *Branch:
		// Determine the common prefix nibbles between existing branch and new leaf node
		tmpPrefix := commonPrefix(subPath, v.prefix)
		// Check for common prefix matching branch prefix
		if string(tmpPrefix) == string(v.prefix) {
			// Insert new value in existing branch
			v.insert(
				subPath,
				key,
				val,
			)
			b.updateHash()
			return
		}
		// Create a new branch node with the common prefix
		tmpBranch := newBranch(tmpPrefix)
		// Adjust existing branch prefix and add to new branch
		newOrigBranchPrefix := v.prefix[len(tmpPrefix):]
		v.prefix = newOrigBranchPrefix[1:]
		v.updateHash()
		tmpBranch.addChild(int(newOrigBranchPrefix[0]), v)
		// Insert new value in new branch
		tmpBranch.insert(
			subPath,
			key,
			val,
		)
		// Replace existing branch node with new branch node
		b.children[childIdx] = tmpBranch
		b.updateHash()

	default:
		panic(
			fmt.Sprintf(
				"unknown Node type %T...this should never happen",
				existingChild,
			),
		)
	}
}

func (b *Branch) delete(path []Nibble) error {
	// Determine path minus the current node prefix
	pathMinusPrefix := path[len(b.prefix):]
	// Determine which child slot the next nibble in the path fits in
	childIdx := int(pathMinusPrefix[0])
	// Determine sub-path for key. We strip off the first nibble, since it's implied by
	// the child slot that it's in
	subPath := pathMinusPrefix[1:]
	if b.children[childIdx] == nil {
		return ErrKeyNotExist
	}
	existingChild := b.children[childIdx]
	switch v := existingChild.(type) {
	case *Leaf:
		if string(v.suffix) != string(subPath) {
			return ErrKeyNotExist
		}
		b.children[childIdx] = nil
		b.size--
		b.updateHash()
	case *Branch:
		err := v.delete(
			subPath,
		)
		if err != nil {
			return err
		}
		// Merge branch with only one child
		if v.size == 1 {
			// Find non-nil child entry
			for tmpChildIdx, tmpChild := range v.children {
				if tmpChild != nil {
					// Update child node suffix to include branch prefix and implied nibble from child slot
					switch v2 := tmpChild.(type) {
					case *Leaf:
						newSuffix := slices.Clone(v.prefix)
						newSuffix = append(newSuffix, Nibble(tmpChildIdx))
						newSuffix = append(newSuffix, v2.suffix...)
						v2.suffix = newSuffix
						v2.updateHash()
					case *Branch:
						newPrefix := slices.Clone(v.prefix)
						newPrefix = append(newPrefix, Nibble(tmpChildIdx))
						newPrefix = append(newPrefix, v2.prefix...)
						v2.prefix = newPrefix
						v2.updateHash()
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
				"unknown Node type %T...this should never happen",
				existingChild,
			),
		)
	}
	return nil
}

func (b *Branch) generateProof(path []Nibble) (*Proof, error) {
	// Determine path minus the current node prefix
	pathMinusPrefix := path[len(b.prefix):]
	// Determine which child slot the next nibble in the path fits in
	childIdx := int(pathMinusPrefix[0])
	// Determine sub-path for key. We strip off the first nibble, since it's implied by
	// the child slot that it's in
	subPath := pathMinusPrefix[1:]
	if b.children[childIdx] == nil {
		return nil, ErrKeyNotExist
	}
	existingChild := b.children[childIdx]
	proof, err := existingChild.generateProof(subPath)
	if err != nil {
		return nil, err
	}
	proof.Rewind(childIdx, len(b.prefix), b.children[:])
	return proof, nil
}

func (b *Branch) addChild(slot int, child Node) {
	empty := b.children[slot] == nil

	b.children[slot] = child
	// Increment the child node count
	if empty {
		b.size++
	}
	b.updateHash()
}

func commonPrefix(prefixA []Nibble, prefixB []Nibble) []Nibble {
	ret := []Nibble{}
	tmpLen := min(len(prefixB), len(prefixA))
	for i := range tmpLen {
		if prefixA[i] != prefixB[i] {
			break
		}
		ret = append(ret, prefixA[i])
	}
	return ret
}
