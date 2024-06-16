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
	"testing"
)

func TestTrieHashChanges(t *testing.T) {
	trie := NewTrieEmpty()
	hash0 := trie.Hash().String()
	trie.Set([]byte("abcd"), []byte("1"))
	hash1 := trie.Hash().String()
	if hash0 == hash1 {
		t.Errorf("hash did not change after insert: old %s, new %s", hash0, hash1)
		return
	}
	trie.Set([]byte("bcde"), []byte("2"))
	hash2 := trie.Hash().String()
	if hash1 == hash2 {
		t.Errorf("hash did not change after insert: old %s, new %s", hash1, hash2)
	}
}

func TestTrieDelete(t *testing.T) {
	trie := NewTrieEmpty()
	trie.Set([]byte("abcd"), []byte("1"))
	hash1 := trie.Hash().String()
	trie.Set([]byte("bcde"), []byte("2"))
	hash2 := trie.Hash().String()
	if err := trie.Delete([]byte("bcde")); err != nil {
		t.Errorf("unexpected error: %s", err)
		return
	}
	hash3 := trie.Hash().String()
	if hash2 == hash3 {
		t.Errorf("hash did not change after delete")
		return
	}
	if hash1 != hash3 {
		t.Errorf("hash is different before and after set/delete of key: got %s, expected %s", hash3, hash1)
	}
}

func TestTrieGet(t *testing.T) {
	testKey := []byte{0xaa, 0xff}
	testVal := []byte("1")
	trie := NewTrieEmpty()
	trie.Set(testKey, testVal)
	tmpVal, err := trie.Get(testKey)
	if err != nil {
		t.Errorf("unexpected error getting key: %s", err)
		return
	}
	if string(tmpVal) != string(testVal) {
		t.Errorf("did not get expected value for key: got %x, expected %x", tmpVal, testVal)
	}
}

func TestTrieHas(t *testing.T) {
	testKey := []byte{0xaa, 0xff}
	trie := NewTrieEmpty()
	if trie.Has(testKey) {
		t.Errorf("has key when shouldn't")
		return
	}
	trie.Set(testKey, []byte("1"))
	if !trie.Has(testKey) {
		t.Errorf("does not have key when should")
	}
}
