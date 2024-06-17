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

import "testing"

func TestLeafPrefixEven(t *testing.T) {
	testPrefix := []Nibble{0xa, 0xb}
	testValue := []byte{0x0, 0x1, 0xe, 0xf}
	// TODO: verify this is actually the expected value. this was pulled from the code output
	expectedHashHex := "349a800c2b1cf864e03adf4e2c6004a8c7d40e0572700424ac9543695d407f18"
	l := newLeaf(
		testPrefix,
		testValue,
	)
	leafHashHex := l.Hash().String()
	if leafHashHex != expectedHashHex {
		t.Errorf("did not get expected hash: got %s, expected %s", leafHashHex, expectedHashHex)
	}
}

func TestLeafPrefixOdd(t *testing.T) {
	testPrefix := []Nibble{0xa, 0xb, 0xc}
	testValue := []byte{0x0, 0x1, 0xe, 0xf}
	// TODO: verify this is actually the expected value. this was pulled from the code output
	expectedHashHex := "5cef0f2da340f3856c3a54d12d8ffdedaf7c97dc49634edec4f6c8638ae0d0f8"
	l := newLeaf(
		testPrefix,
		testValue,
	)
	leafHashHex := l.Hash().String()
	if leafHashHex != expectedHashHex {
		t.Errorf("did not get expected hash: got %s, expected %s", leafHashHex, expectedHashHex)
	}
}
