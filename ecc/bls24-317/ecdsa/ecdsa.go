// Copyright 2020 ConsenSys Software Inc.
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

// Code generated by consensys/gnark-crypto DO NOT EDIT

package ecdsa

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/sha512"
	"io"
	"math/big"

	"github.com/consensys/gnark-crypto/ecc/bls24-317"
	"github.com/consensys/gnark-crypto/ecc/bls24-317/fr"
)

// PublicKey represents an ECDSA public key
type PublicKey struct {
	Q bls24317.G1Affine
}

// PrivateKey represents an ECDSA private key
type PrivateKey struct {
	PublicKey
	Secret big.Int
}

// Signature represents an ECDSA signature
type Signature struct {
	r, s big.Int
}

// params are the ECDSA public parameters
type params struct {
	Base  bls24317.G1Affine
	Order *big.Int
}

var one = new(big.Int).SetInt64(1)

// randFieldElement returns a random element of the order of the given
// curve using the procedure given in FIPS 186-4, Appendix B.5.1.
func (pp params) randFieldElement(rand io.Reader) (k big.Int, err error) {
	b := make([]byte, fr.Bits/8+8)
	_, err = io.ReadFull(rand, b)
	if err != nil {
		return
	}

	k = *new(big.Int).SetBytes(b)
	n := new(big.Int).Sub(pp.Order, one)
	k.Mod(&k, n)
	k.Add(&k, one)
	return
}

// GenerateKey generates a public and private key pair.
func (pp params) GenerateKey(rand io.Reader) (*PrivateKey, error) {

	k, err := pp.randFieldElement(rand)
	if err != nil {
		return nil, err

	}

	privateKey := new(PrivateKey)
	privateKey.Secret = k
	privateKey.PublicKey.Q.ScalarMultiplication(&pp.Base, &k)
	return privateKey, nil
}

// hashToInt converts a hash value to an integer. Per FIPS 186-4, Section 6.4,
// we use the left-most bits of the hash to match the bit-length of the order of
// the curve. This also performs Step 5 of SEC 1, Version 2.0, Section 4.1.3.
func hashToInt(hash []byte) big.Int {
	if len(hash) > fr.Bytes {
		hash = hash[:fr.Bytes]
	}

	ret := new(big.Int).SetBytes(hash)
	excess := len(hash)*8 - fr.Bits
	if excess > 0 {
		ret.Rsh(ret, uint(excess))
	}
	return *ret
}

type zr struct{}

// Read replaces the contents of dst with zeros. It is safe for concurrent use.
func (zr) Read(dst []byte) (n int, err error) {
	for i := range dst {
		dst[i] = 0
	}
	return len(dst), nil
}

var zeroReader = zr{}

const (
	aesIV = "gnark-crypto IV." // must be 16 chars (equal block size)
)

func nonce(rand io.Reader, privateKey *PrivateKey, hash []byte) (csprng *cipher.StreamReader, err error) {
	// This implementation derives the nonce from an AES-CTR CSPRNG keyed by:
	//
	//    SHA2-512(privateKey.Secret ∥ entropy ∥ hash)[:32]
	//
	// The CSPRNG key is indifferentiable from a random oracle as shown in
	// [Coron], the AES-CTR stream is indifferentiable from a random oracle
	// under standard cryptographic assumptions (see [Larsson] for examples).
	//
	// [Coron]: https://cs.nyu.edu/~dodis/ps/merkle.pdf
	// [Larsson]: https://web.archive.org/web/20040719170906/https://www.nada.kth.se/kurser/kth/2D1441/semteo03/lecturenotes/assump.pdf

	// Get 256 bits of entropy from rand.
	entropy := make([]byte, 32)
	_, err = io.ReadFull(rand, entropy)
	if err != nil {
		return

	}

	// Initialize an SHA-512 hash context; digest...
	md := sha512.New()
	md.Write(privateKey.Secret.Bytes()) // the private key,
	md.Write(entropy)                   // the entropy,
	md.Write(hash)                      // and the input hash;
	key := md.Sum(nil)[:32]             // and compute ChopMD-256(SHA-512),
	// which is an indifferentiable MAC.

	// Create an AES-CTR instance to use as a CSPRNG.
	block, _ := aes.NewCipher(key)

	// Create a CSPRNG that xors a stream of zeros with
	// the output of the AES-CTR instance.
	csprng = &cipher.StreamReader{
		R: zeroReader,
		S: cipher.NewCTR(block, []byte(aesIV)),
	}

	return csprng, err
}

// Sign performs the ECDSA signature
//
// k ← 𝔽r (random)
// R = k ⋅ Base
// r = x_R (mod Order)
// s = k⁻¹ . (m + sk ⋅ r)
// signature = {s, r}
//
// SEC 1, Version 2.0, Section 4.1.3
func (pp params) Sign(hash []byte, privateKey PrivateKey, rand io.Reader) (signature Signature, err error) {
	var kInv big.Int
	for {
		for {
			csprng, err := nonce(rand, &privateKey, hash)
			if err != nil {
				return Signature{}, err
			}
			k, err := pp.randFieldElement(csprng)
			if err != nil {
				return Signature{}, err
			}

			var R bls24317.G1Affine
			R.ScalarMultiplication(&pp.Base, &k)
			kInv.ModInverse(&k, pp.Order)

			R.X.BigInt(&signature.r)
			signature.r.Mod(&signature.r, pp.Order)
			if signature.r.Sign() != 0 {
				break
			}
		}
		signature.s.Mul(&signature.r, &privateKey.Secret)
		m := hashToInt(hash)
		signature.s.Add(&m, &signature.s).
			Mul(&kInv, &signature.s).
			Mod(&signature.s, pp.Order) // pp.Order != 0
		if signature.s.Sign() != 0 {
			break
		}
	}

	return signature, err
}

// Verify validates the ECDSA signature
//
// R ?= s⁻¹ ⋅ m ⋅ Base + s⁻¹ ⋅ r ⋅ publiKey
//
// SEC 1, Version 2.0, Section 4.1.4
func (pp params) Verify(hash []byte, signature Signature, publicKey bls24317.G1Affine) bool {

	if signature.r.Sign() <= 0 || signature.s.Sign() <= 0 {
		return false
	}
	if signature.r.Cmp(pp.Order) >= 0 || signature.s.Cmp(pp.Order) >= 0 {
		return false
	}

	sInv := new(big.Int).ModInverse(&signature.s, pp.Order)
	e := hashToInt(hash)
	u1 := new(big.Int).Mul(&e, sInv)
	u1.Mod(u1, pp.Order)
	u2 := new(big.Int).Mul(&signature.r, sInv)
	u2.Mod(u2, pp.Order)

	var U1, U2 bls24317.G1Jac
	U1.ScalarMultiplicationAffine(&pp.Base, u1)
	U2.ScalarMultiplicationAffine(&publicKey, u2).
		AddAssign(&U1)

	var z big.Int
	U2.Z.Square(&U2.Z).
		Inverse(&U2.Z).
		Mul(&U2.Z, &U2.X).
		BigInt(&z)

	z.Mod(&z, pp.Order)

	return z.Cmp(&signature.r) == 0

}
