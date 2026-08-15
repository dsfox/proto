// Copyright 2024 Teamgram Authors
//  All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//   http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
// Author: teamgramio (teamgram.io@gmail.com)
//

package crypto

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"math/big"
)

type RSACryptor struct {
	key *rsa.PrivateKey
}

func NewRSACryptor(keyFile string) (*RSACryptor, error) {
	pkcs1PemPrivateKey, err := ioutil.ReadFile(keyFile)
	if err != nil {
		return nil, err
	}
	return NewRSACryptorByKeyData(pkcs1PemPrivateKey)
}

func NewRSACryptorByKeyData(pkcs1PemPrivateKey []byte) (*RSACryptor, error) {
	block, _ := pem.Decode(pkcs1PemPrivateKey)
	if block == nil {
		return nil, fmt.Errorf("invalid pemsKeyData")
	}

	key, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return nil, err
	}

	return &RSACryptor{key}, nil
}

// size is the modulus in bytes - the length an RSA operation is defined to
// produce, whatever the value happens to be worth.
func (m *RSACryptor) size() int {
	return (m.key.N.BitLen() + 7) / 8
}

// Encrypt and Decrypt both hand back exactly size() bytes, big-endian, padded
// on the left.
//
// big.Int has no notion of a leading zero, so Bytes() returns 255 bytes for the
// one value in 256 whose top byte is zero - and both halves of this file used
// to get that wrong. Decrypt returned the short slice, and the handshake, which
// requires 256, refused it inside a pool job whose error nobody logs: the
// client got no answer and no refusal, and sat on "Connecting" until it gave up
// and started again. Encrypt padded on the right, which for a big-endian number
// is multiplication by 256 - a different ciphertext, silently.
//
// FillBytes pads on the left and panics if the value does not fit; it cannot
// here, because both results are reduced modulo N and the buffer is N's size.
func (m *RSACryptor) Encrypt(b []byte) []byte {
	c := new(big.Int)
	c.Exp(new(big.Int).SetBytes(b), big.NewInt(int64(m.key.E)), m.key.N)
	return c.FillBytes(make([]byte, m.size()))
}

func (m *RSACryptor) Decrypt(b []byte) []byte {
	c := new(big.Int)
	c.Exp(new(big.Int).SetBytes(b), m.key.D, m.key.N)
	return c.FillBytes(make([]byte, m.size()))
}
