package crypto

import (
	"bytes"
	"math/big"
	"testing"
)

// One RSA result in 256 has a zero as its top byte, and big.Int does not keep
// it. That was issue #68: a handshake that died about once every few hundred
// connections, with nothing in the log, because the length check that caught
// the short slice ran inside a pool job whose error is dropped. The client saw
// no answer and no refusal, waited out its timeout and started again.
//
// Waiting for the one-in-256 to turn up is not a test. Both cases below build
// the awkward value on purpose.

// The plaintext is chosen; the ciphertext is whatever encrypts to it. Nothing
// here goes through Encrypt, so this measures Decrypt alone.
func TestDecryptKeepsTheLeadingZero(t *testing.T) {
	m, err := NewRSACryptorByKeyData(pkcs1PemPrivateKey)
	if err != nil {
		t.Fatalf("test key: %v", err)
	}

	plain := make([]byte, 256)
	for i := 1; i < len(plain); i++ {
		plain[i] = byte(i) // plain[0] stays zero - that is the whole point
	}

	cipher := new(big.Int).Exp(
		new(big.Int).SetBytes(plain), big.NewInt(int64(m.key.E)), m.key.N)

	got := m.Decrypt(cipher.FillBytes(make([]byte, 256)))
	if len(got) != 256 {
		t.Fatalf("decrypted to %d bytes, and the handshake requires 256; "+
			"the leading zero was dropped", len(got))
	}
	if !bytes.Equal(got, plain) {
		t.Fatalf("decrypted to something other than what was encrypted")
	}
}

// Padding a big-endian number on the right multiplies it by 256, so the old
// Encrypt did not merely return an odd length - it returned a different number.
func TestEncryptPadsOnTheLeft(t *testing.T) {
	m, err := NewRSACryptorByKeyData(pkcs1PemPrivateKey)
	if err != nil {
		t.Fatalf("test key: %v", err)
	}

	// Walk until the ciphertext is one of the short ones. It arrives within a
	// few hundred tries; the bound is only there so a broken key cannot hang.
	e := big.NewInt(int64(m.key.E))
	for i := int64(2); i < 20000; i++ {
		plain := big.NewInt(i)
		want := new(big.Int).Exp(plain, e, m.key.N)
		if len(want.Bytes()) == 256 {
			continue
		}

		got := m.Encrypt(plain.Bytes())
		if len(got) != 256 {
			t.Fatalf("encrypted %d to %d bytes, want 256", i, len(got))
		}
		if new(big.Int).SetBytes(got).Cmp(want) != 0 {
			t.Fatalf("encrypted %d to the wrong number: padded on the wrong "+
				"side, so it reads 256 times too large", i)
		}
		return
	}
	t.Skip("no short ciphertext in 20000 tries - unexpected, but not a failure")
}
