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
	"encoding/hex"
	"fmt"

	"golang.org/x/crypto/blake2b"
)

const (
	HashSize = 32
)

type Hash [HashSize]byte

var NullHash = Hash{}

func (h Hash) String() string {
	return hex.EncodeToString(h[:])
}

func (h Hash) Bytes() []byte {
	return h[:]
}

func HashValue(val []byte) Hash {
	tmpHash, err := blake2b.New(HashSize, nil)
	if err != nil {
		panic(
			fmt.Sprintf(
				"failed hashing value...this should never happen: %s",
				err,
			),
		)
	}
	tmpHash.Write(val)
	return Hash(tmpHash.Sum(nil))
}
