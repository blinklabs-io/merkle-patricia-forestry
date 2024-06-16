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

type Trie struct {
	*Branch
}

func NewTrie(rootHash [HashSize]byte) *Trie {
	t := &Trie{
		Branch: newBranch(nil),
	}
	t.Branch.hash = rootHash
	return t
}

func NewTrieEmpty() *Trie {
	return NewTrie(NullHash)
}

func (t *Trie) IsEmpty() bool {
	return t.Branch.size == 0
}

func (t *Trie) Set(key []byte, val []byte) {
	keyNibbles := bytesToNibbles(key)
	t.insert(keyNibbles, val)
}

func (t *Trie) Delete(key []byte) error {
	keyNibbles := bytesToNibbles(key)
	return t.delete(keyNibbles)
}

func (t *Trie) Get(key []byte) ([]byte, error) {
	keyNibbles := bytesToNibbles(key)
	return t.get(keyNibbles)
}

func (t *Trie) Has(key []byte) bool {
	keyNibbles := bytesToNibbles(key)
	_, err := t.get(keyNibbles)
	return err == nil
}
