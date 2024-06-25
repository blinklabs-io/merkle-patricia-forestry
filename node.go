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

type Node interface {
	isNode()
	Hash() Hash
	String() string
	generateProof([]Nibble) (*Proof, error)
}

func merkleRoot(nodes []Node) Hash {
	// Gather child node hashes
	tmpHashes := make([]Hash, 0, len(nodes))
	for _, child := range nodes {
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
