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
	"testing"

	"github.com/blinklabs-io/gouroboros/cbor"
)

var proofTestDefs = []struct {
	key             []byte
	expectedCborHex string
}{
	{
		key:             []byte(fruitsTestEntries[0].key),
		expectedCborHex: "9fd8799f005f5840c7bfa4472f3a98ebe0421e8f3f03adf0f7c4340dec65b4b92b1c9f0bed209eb47238ba5d16031b6bace4aee22156f5028b0ca56dc24f7247d6435292e82c039c58403490a825d2e8deddf8679ce2f95f7e3a59d9c3e1af4a49b410266d21c9344d6d79519b8cdfbd053e5a86cf28a781debae71638cd77f85aad4b88869373d9dcfdffffd87b9f0058205cddcd30a0a388cf6feb3fd6e112c96e9daf23e3a9c8a334e7044650471aaa9e5820f429821ddf89c9df3c7fbb5aa6fadb6c246d75ceede53173ce59d70dde375d14ffd87b9f0058205e7ccfedd44c90423b191ecca1eb21dfbac865d561bace8c0f3e94ae7edf444058207c3715aba2db74d565a6ce6cc72f20d9cb4652ddb29efe6268be15b105e40911ffff",
	},
	{
		key:             []byte(fruitsTestEntries[1].key),
		expectedCborHex: "9fd8799f005f58404be28f4839135e1f8f5372a90b54bb7bfaf997a5d13711bb4d7d93f9d4e04fbe280ada5ef30d55433934bbc73c89d550ee916f62822c34645e04bb66540c120f5840965c07fa815b86794e8703cee7e8f626c88d7da639258d2466aae67d5d041c5a117abf0e19fb78e0535891d82e5ece1310a1cf11674587dbba304c395769a988ffffff",
	},
	{
		key:             []byte(fruitsTestEntries[2].key),
		expectedCborHex: "9fd8799f005f5840c7bfa4472f3a98ebe0421e8f3f03adf0f7c4340dec65b4b92b1c9f0bed209eb45fdf82687b1ab133324cebaf46d99d49f92720c5ded08d5b02f57530f2cc5a5f5840cf22cbaac4ab605dd13dbde57080661b53d8a7e23534c733acf50125cf0e5bcac9431d708d20021f1fa3f4f03468b8de194398072a402e7877376d06f747575affffd87b9f0158203ed002d6885ab5d92e1307fccd1d021c32ec429192aea10cb2fd688b92aef3ac58207c3715aba2db74d565a6ce6cc72f20d9cb4652ddb29efe6268be15b105e40911ffff",
	},
	{
		key:             []byte(fruitsTestEntries[3].key),
		expectedCborHex: "9fd8799f005f58404be28f4839135e1f8f5372a90b54bb7bfaf997a5d13711bb4d7d93f9d4e04fbefa63eb4576001d8658219f928172eccb5448b4d7d62cd6d95228e13ebcbd53505840be527bcfc7febe3c560057d97f4190bd24b537a322315f84daafab3ada562b50c2f2115774c117f184b58dba7a23d2c93968aa40387ceb0c9a9f53e4f594e881ffffd87b9f005820b67e71b092e6a54576fa23b0eb48c5e5794a3fb5480983e48b40e453596cc48b58207c3715aba2db74d565a6ce6cc72f20d9cb4652ddb29efe6268be15b105e40911ffff",
	},
	{
		key:             []byte(fruitsTestEntries[4].key),
		expectedCborHex: "9fd8799f005f5840c7bfa4472f3a98ebe0421e8f3f03adf0f7c4340dec65b4b92b1c9f0bed209eb45fdf82687b1ab133324cebaf46d99d49f92720c5ded08d5b02f57530f2cc5a5f58401508f13471a031a21277db8817615e62a50a7427d5f8be572746aa5f0d498417520a7f805c5f674e2deca5230b6942bbc71586dc94a783eebe1ed58c9a864e53ffffd8799f035f58402549707d84ecc2fa100fd85bf15f2ec99da70d4b3a39588c1138331eb0e00d3e85c09af929492a871e4fae32d9d5c36e352471cd659bcdb61de08f1722acc3b158400eb923b0cbd24df54401d998531feead35a47a99f4deed205de4af81120f97610000000000000000000000000000000000000000000000000000000000000000ffffff",
	},
	{
		key:             []byte(fruitsTestEntries[17].key),
		expectedCborHex: "9fd8799f005f58404be28f4839135e1f8f5372a90b54bb7bfaf997a5d13711bb4d7d93f9d4e04fbe280ada5ef30d55433934bbc73c89d550ee916f62822c34645e04bb66540c120f5840965c07fa815b86794e8703cee7e8f626c88d7da639258d2466aae67d5d041c5ada1771d107c86c8e68da458063a47f9cdb63ddb9e922ab6ccb18d9e6d4b7aaf9ffffd87b9f005820fb69c0d60ec9bfb6cafa5cf54675edfbb0017b873ee92a5dbb6bdabcfb3521455820b5898c51c32083e91b8c18c735d0ba74e08f964a20b1639c189d1e8704b78a09ffff",
	},
}

func TestProofMarshalCbor(t *testing.T) {
	trie := NewTrie()
	for _, entry := range fruitsTestEntries {
		trie.Set([]byte(entry.key), []byte(entry.value))
	}
	for _, testDef := range proofTestDefs {
		proof, err := trie.Prove(testDef.key)
		if err != nil {
			t.Fatalf("got unexpected error when generating proof: %s", err)
		}
		proofCbor, err := cbor.Encode(proof)
		if err != nil {
			t.Fatalf("got unexpected error when encoding proof as CBOR: %s", err)
		}
		cborHex := hex.EncodeToString(proofCbor)
		if cborHex != testDef.expectedCborHex {
			t.Fatalf("did not get expected proof CBOR\n  got:    %s\n  wanted: %s", cborHex, testDef.expectedCborHex)
		}
	}
}
