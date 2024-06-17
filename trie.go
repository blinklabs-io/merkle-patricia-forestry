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

import (
	"fmt"
	"strings"
)

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
		nodeStr := strings.SplitN(
			t.rootNode.String(),
			"\n",
			2,
		)[1]
		ret += fmt.Sprintf(
			"\n%s",
			nodeStr,
		)
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
	keyHash := HashValue(key)
	keyHashNibbles := bytesToNibbles(keyHash.Bytes())
	if t.rootNode == nil {
		l := newLeaf(
			keyHashNibbles,
			key,
			val,
		)
		t.rootNode = l
		return
	}
	switch n := t.rootNode.(type) {
	case *Leaf:
		// Update value for matching existing leaf node
		if string(keyHashNibbles) == string(n.suffix) {
			n.Set(val)
			return
		}
		// Create new branch
		tmpBranch := newBranch(nil)
		// Insert original value
		tmpBranch.insert(n.suffix, n.key, n.value)
		// Insert new value
		tmpBranch.insert(keyHashNibbles, key, val)
		// Replace root node
		t.rootNode = tmpBranch
	case *Branch:
		n.insert(keyHashNibbles, key, val)
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
	keyHash := HashValue(key)
	keyHashNibbles := bytesToNibbles(keyHash.Bytes())
	switch n := t.rootNode.(type) {
	case *Leaf:
		if string(keyHashNibbles) == string(n.suffix) {
			t.rootNode = nil
			return nil
		}
		return ErrKeyNotExist
	case *Branch:
		if err := n.delete(keyHashNibbles); err != nil {
			return err
		}
		// Move single remaining child node to root node
		if n.size == 1 {
			tmpChildren := n.getChildren()
			tmpChild := tmpChildren[0].(*Leaf)
			childKeyHash := HashValue(tmpChild.key)
			childKeyHashNibbles := bytesToNibbles(childKeyHash.Bytes())
			newNode := newLeaf(
				childKeyHashNibbles,
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
	keyHash := HashValue(key)
	keyHashNibbles := bytesToNibbles(keyHash.Bytes())
	switch n := t.rootNode.(type) {
	case *Leaf:
		if string(n.suffix) == string(keyHashNibbles) {
			return n.value, nil
		}
		return nil, ErrKeyNotExist
	case *Branch:
		return n.get(keyHashNibbles)
	default:
		panic("unknown node type...this should never happen")
	}
}

// Has returns whether the specified key exists in the trie
func (t *Trie) Has(key []byte) bool {
	_, err := t.Get(key)
	return err == nil
}
