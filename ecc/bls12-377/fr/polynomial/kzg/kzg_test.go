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

package kzg

import (
	"bytes"
	"math/big"
	"reflect"
	"testing"

	"github.com/consensys/gnark-crypto/ecc/bls12-377"
	"github.com/consensys/gnark-crypto/ecc/bls12-377/fr"
	"github.com/consensys/gnark-crypto/ecc/bls12-377/fr/fft"
	bls12377_pol "github.com/consensys/gnark-crypto/ecc/bls12-377/fr/polynomial"
	"github.com/consensys/gnark-crypto/polynomial"
)

var _alphaSetup fr.Element

func init() {
	//_alphaSetup.SetRandom()
	_alphaSetup.SetString("1234")
}

// NewKZG returns a new Scheme instance
// this is used for test purposes
func NewKZG(size int, alpha fr.Element) Scheme {

	var s Scheme
	d := fft.NewDomain(uint64(size), 0, false)
	s.Domain = *d
	s.SRS.G1 = make([]bls12377.G1Affine, size)

	// generate the SRS
	var alphaBigInt big.Int
	alpha.ToBigIntRegular(&alphaBigInt)

	_, _, gen1Aff, gen2Aff := bls12377.Generators()
	s.SRS.G1[0] = gen1Aff
	s.SRS.G2[0] = gen2Aff
	s.SRS.G2[1].ScalarMultiplication(&gen2Aff, &alphaBigInt)

	alphas := make([]fr.Element, size-1)
	alphas[0] = alpha
	for i := 1; i < len(alphas); i++ {
		alphas[i].Mul(&alphas[i-1], &alpha)
	}
	for i := 0; i < len(alphas); i++ {
		alphas[i].FromMont()
	}
	g1s := bls12377.BatchScalarMultiplicationG1(&gen1Aff, alphas)
	copy(s.SRS.G1[1:], g1s)

	return s
}

func randomPolynomial(size int) bls12377_pol.Polynomial {
	f := make(bls12377_pol.Polynomial, size)
	for i := 0; i < size; i++ {
		f[i].SetRandom()
	}
	return f
}

func TestDividePolyByXminusA(t *testing.T) {

	sizePol := 230

	domain := fft.NewDomain(uint64(sizePol), 0, false)

	// build random polynomial
	pol := make(bls12377_pol.Polynomial, sizePol)
	for i := 0; i < sizePol; i++ {
		pol[i].SetRandom()
	}

	// evaluate the polynomial at a random point
	var point fr.Element
	point.SetRandom()
	evaluation := pol.Eval(&point).(fr.Element)

	// compute f-f(a)/x-a
	h := dividePolyByXminusA(*domain, pol, evaluation, point)

	if len(h) != 229 {
		t.Fatal("inconsistant size of quotient")
	}

	// probabilistic test (using Schwartz Zippel lemma, evaluation at one point is enough)
	var randPoint, xminusa fr.Element
	randPoint.SetRandom()

	polRandpoint := pol.Eval(&randPoint).(fr.Element)
	polRandpoint.Sub(&polRandpoint, &evaluation) // f(rand)-f(point)

	hRandPoint := h.Eval(&randPoint).(fr.Element)
	xminusa.Sub(&randPoint, &point) // rand-point

	// f(rand)-f(point)	==? h(rand)*(rand-point)
	hRandPoint.Mul(&hRandPoint, &xminusa)

	if !hRandPoint.Equal(&polRandpoint) {
		t.Fatal("Error f-f(a)/x-a")
	}
}

func TestSerialization(t *testing.T) {

	// create a KZG scheme
	s := NewKZG(64, _alphaSetup)

	// serialize it...
	var buf bytes.Buffer
	_, err := s.WriteTo(&buf)
	if err != nil {
		t.Fatal(err)
	}

	// reconstruct the scheme
	var _s Scheme
	_, err = _s.ReadFrom(&buf)
	if err != nil {
		t.Fatal(err)
	}

	// compare
	if !reflect.DeepEqual(&s, &_s) {
		t.Fatal("scheme serialization failed")
	}

}

func TestCommit(t *testing.T) {

	// create a KZG scheme
	s := NewKZG(64, _alphaSetup)

	// create a polynomial
	f := make(bls12377_pol.Polynomial, 60)
	for i := 0; i < 60; i++ {
		f[i].SetRandom()
	}

	// commit using the method from KZG
	_kzgCommit, err := s.Commit(&f)
	if err != nil {
		t.Fatal(err)
	}
	var kzgCommit bls12377.G1Affine
	kzgCommit.Set(_kzgCommit.(*bls12377.G1Affine))

	// check commitment using manual commit
	var x fr.Element
	x.SetString("1234")
	fx := f.Eval(&x).(fr.Element)
	var fxbi big.Int
	fx.ToBigIntRegular(&fxbi)
	var manualCommit bls12377.G1Affine
	manualCommit.Set(&s.SRS.G1[0])
	manualCommit.ScalarMultiplication(&manualCommit, &fxbi)

	// compare both results
	if !kzgCommit.Equal(&manualCommit) {
		t.Fatal("error KZG commitment")
	}

}

func TestVerifySinglePoint(t *testing.T) {

	// create a KZG scheme
	s := NewKZG(64, _alphaSetup)

	// create a polynomial
	f := randomPolynomial(60)

	// commit the polynomial
	digest, err := s.Commit(&f)
	if err != nil {
		t.Fatal(err)
	}

	// compute opening proof at a random point
	var point fr.Element
	point.SetString("4321")
	proof, err := s.Open(&point, &f)
	if err != nil {
		t.Fatal(err)
	}

	// verify the claimed valued
	_proof := proof.(*Proof)
	expected := f.Eval(point).(fr.Element)
	if !_proof.ClaimedValue.Equal(&expected) {
		t.Fatal("inconsistant claimed value")
	}

	// verify correct proof
	err = s.Verify(digest, proof)
	if err != nil {
		t.Fatal(err)
	}

	// verify wrong proof
	_proof = proof.(*Proof)
	_proof.ClaimedValue.Double(&_proof.ClaimedValue)
	err = s.Verify(digest, _proof)
	if err == nil {
		t.Fatal("verifying wrong proof should have failed")
	}
}

func TestBatchVerifySinglePoint(t *testing.T) {

	// create a KZG scheme
	s := NewKZG(64, _alphaSetup)

	// create polynomials
	f := make([]polynomial.Polynomial, 10)
	for i := 0; i < 10; i++ {
		_f := randomPolynomial(60)
		f[i] = &_f
	}

	// commit the polynomials
	digests := make([]polynomial.Digest, 10)
	for i := 0; i < 10; i++ {
		digests[i], _ = s.Commit(f[i])

	}

	// compute opening proof at a random point
	var point fr.Element
	point.SetString("4321")
	proof, err := s.BatchOpenSinglePoint(&point, digests, f)
	if err != nil {
		t.Fatal(err)
	}

	// verify the claimed values
	_proof := proof.(*BatchProofsSinglePoint)
	for i := 0; i < 10; i++ {
		expectedClaim := f[i].Eval(point).(fr.Element)
		if !expectedClaim.Equal(&_proof.ClaimedValues[i]) {
			t.Fatal("inconsistant claimed values")
		}
	}

	// verify correct proof
	err = s.BatchVerifySinglePoint(digests, proof)
	if err != nil {
		t.Fatal(err)
	}

	// verify wrong proof
	_proof.ClaimedValues[0].Double(&_proof.ClaimedValues[0])
	err = s.BatchVerifySinglePoint(digests, _proof)
	if err == nil {
		t.Fatal("verifying wrong proof should have failed")
	}

}

const benchSize = 1 << 16

func BenchmarkKZGCommit(b *testing.B) {
	// kzg scheme
	s := NewKZG(benchSize, _alphaSetup)

	// random polynomial
	p := randomPolynomial(benchSize / 2)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = s.Commit(&p)
	}
}

func BenchmarkKZGOpen(b *testing.B) {
	// kzg scheme
	s := NewKZG(benchSize, _alphaSetup)

	// random polynomial
	p := randomPolynomial(benchSize / 2)
	var r fr.Element
	r.SetRandom()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = s.Open(r, &p)
	}
}

func BenchmarkKZGVerify(b *testing.B) {
	// kzg scheme
	s := NewKZG(benchSize, _alphaSetup)

	// random polynomial
	p := randomPolynomial(benchSize / 2)
	var r fr.Element
	r.SetRandom()

	// commit
	comm, err := s.Commit(&p)
	if err != nil {
		b.Fatal(err)
	}

	// open
	openingProof, err := s.Open(r, &p)
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		s.Verify(comm, openingProof)
	}
}

func BenchmarkKZGBatchOpen10(b *testing.B) {
	// kzg scheme
	s := NewKZG(benchSize, _alphaSetup)

	// 10 random polynomials
	var ps [10]polynomial.Polynomial
	for i := 0; i < 10; i++ {
		_p := randomPolynomial(benchSize / 2)
		ps[i] = &_p
	}

	// commitments
	var commitments [10]polynomial.Digest
	for i := 0; i < 10; i++ {
		commitments[i], _ = s.Commit(ps[i])
	}

	var r fr.Element
	r.SetRandom()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		s.BatchOpenSinglePoint(r, commitments[:], ps[:])
	}
}

func BenchmarkKZGBatchVerify10(b *testing.B) {
	// kzg scheme
	s := NewKZG(benchSize, _alphaSetup)

	// 10 random polynomials
	var ps [10]polynomial.Polynomial
	for i := 0; i < 10; i++ {
		_p := randomPolynomial(benchSize / 2)
		ps[i] = &_p
	}

	// commitments
	var commitments [10]polynomial.Digest
	for i := 0; i < 10; i++ {
		commitments[i], _ = s.Commit(ps[i])
	}

	var r fr.Element
	r.SetRandom()

	proof, err := s.BatchOpenSinglePoint(r, commitments[:], ps[:])
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		s.BatchVerifySinglePoint(commitments[:], proof)
	}
}
