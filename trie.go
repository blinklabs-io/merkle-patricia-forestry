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
	"strings"
)

//nolint:unused
type Trie struct {
	rootNode Node
	size     int
}

func NewTrie() *Trie {
	return &Trie{}
}

// String returns a string representation of the entire trie
func (t *Trie) String() string {
	ret := fmt.Sprintf(
		"** #%s **",
		t.Hash().String(),
	)
	if t.rootNode != nil {
		// Strip off first line of output for root node, since we're showing the root hash above
		nodeArr := strings.SplitN(
			t.rootNode.String(),
			"\n",
			2,
		)
		if nodeArr == nil {
			// SplitN gave us nil, return original node from above
			return ret
		}
		nodeStr := nodeArr[1]
		ret += "\n" + nodeStr
	}
	return ret
}

// IsEmpty returns whether the trie is empty
func (t *Trie) IsEmpty() bool {
	return t.rootNode == nil
}

// Hash returns the root hash for the trie
func (t *Trie) Hash() Hash {
	if t.rootNode == nil {
		return NullHash
	}
	return t.rootNode.Hash()
}

// Set adds the specified key and value to the trie. If the key already exists, the value will be updated
func (t *Trie) Set(key []byte, val []byte) {
	path := keyToPath(key)
	if t.rootNode == nil {
		l := newLeaf(
			path,
			key,
			val,
		)
		t.rootNode = l
		return
	}
	switch n := t.rootNode.(type) {
	case *Leaf:
		// Update value for matching existing leaf node
		if string(path) == string(n.suffix) {
			n.Set(val)
			return
		}
		tmpPrefix := commonPrefix(path, n.suffix)
		// Create new branch
		tmpBranch := newBranch(tmpPrefix)
		// Insert original value
		tmpBranch.insert(n.suffix, n.key, n.value)
		// Insert new value
		tmpBranch.insert(path, key, val)
		// Replace root node
		t.rootNode = tmpBranch
	case *Branch:
		// Determine the common prefix nibbles between existing branch and new leaf node
		tmpPrefix := commonPrefix(path, n.prefix)
		// Check for common prefix matching branch prefix
		if string(tmpPrefix) == string(n.prefix) {
			// Insert new value in existing branch
			n.insert(
				path,
				key,
				val,
			)
			return
		}
		// Create a new branch node with the common prefix
		tmpBranch := newBranch(tmpPrefix)
		// Adjust existing branch prefix and add to new branch
		newOrigBranchPrefix := n.prefix[len(tmpPrefix):]
		n.prefix = newOrigBranchPrefix[1:]
		n.updateHash()
		tmpBranch.addChild(int(newOrigBranchPrefix[0]), n)
		// Insert new value in new branch
		tmpBranch.insert(
			path,
			key,
			val,
		)
		t.rootNode = tmpBranch
	default:
		panic("unknown node type...this should never happen")
	}
}

// Delete removes the specified key and associated value from the trie. Returns ErrKeyNotExist
// if the specified key doesn't exist
func (t *Trie) Delete(key []byte) error {
	if t.rootNode == nil {
		return ErrKeyNotExist
	}
	path := keyToPath(key)
	switch n := t.rootNode.(type) {
	case *Leaf:
		if string(path) == string(n.suffix) {
			t.rootNode = nil
			return nil
		}
		return ErrKeyNotExist
	case *Branch:
		if err := n.delete(path); err != nil {
			return err
		}
		// Move single remaining child node to root node
		if n.size == 1 {
			tmpChildren := n.getChildren()
			tmpChild := tmpChildren[0].(*Leaf)
			childPath := keyToPath(tmpChild.key)
			newNode := newLeaf(
				childPath,
				tmpChild.key,
				tmpChild.value,
			)
			t.rootNode = newNode
		}
	default:
		panic("unknown node type...this should never happen")
	}
	return nil
}

// Get returns the value for the specified key or ErrKeyNotExist if the key
// doesn't exist in the trie
func (t *Trie) Get(key []byte) ([]byte, error) {
	if t.rootNode == nil {
		return nil, ErrKeyNotExist
	}
	path := keyToPath(key)
	switch n := t.rootNode.(type) {
	case *Leaf:
		if string(n.suffix) == string(path) {
			return n.value, nil
		}
		return nil, ErrKeyNotExist
	case *Branch:
		return n.get(path)
	default:
		panic("unknown node type...this should never happen")
	}
}

// Has returns whether the specified key exists in the trie
func (t *Trie) Has(key []byte) bool {
	_, err := t.Get(key)
	return err == nil
}

// Prove returns a proof that the given key exists in the trie or ErrKeyNotExist if
// the key doesn't exist in the trie
func (t *Trie) Prove(key []byte) (*Proof, error) {
	if t.rootNode == nil {
		return nil, ErrKeyNotExist
	}
	path := keyToPath(key)
	return t.rootNode.generateProof(path)
}
