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

package iop

import (
	"testing"

	"github.com/consensys/gnark-crypto/ecc/bls24-317/fr"
	"github.com/consensys/gnark-crypto/ecc/bls24-317/fr/fft"
)

// getPermutation returns a deterministic permutation
// of n elements where n is even. The result should be
// interpreted as
// a permutation σ(i)=permutation[i]
// g is a generator of ℤ/nℤ
func getPermutation(n, g int) []int {

	res := make([]int, n)
	a := g
	for i := 0; i < n; i++ {
		res[i] = a
		a += g
		a %= n
	}
	return res
}

func getPermutedPolynomials(sizePolynomials, nbPolynomials int) ([]Polynomial, []Polynomial, []int) {

	numerator := make([]Polynomial, nbPolynomials)
	for i := 0; i < nbPolynomials; i++ {
		// numerator[i] = new(Polynomial)
		numerator[i].Coefficients = randomVector(sizePolynomials)
		numerator[i].Basis = Lagrange
		numerator[i].Layout = Regular
		numerator[i].Status = Locked
	}

	// get permutation
	sigma := getPermutation(sizePolynomials*nbPolynomials, 3)

	// the denominator is the permuted version of the numerators
	// concatenated
	denominator := make([]Polynomial, nbPolynomials)
	for i := 0; i < nbPolynomials; i++ {
		// denominator[i] = new(Polynomial)
		denominator[i].Coefficients = make([]fr.Element, sizePolynomials)
		denominator[i].Basis = Lagrange
		denominator[i].Layout = Regular
		denominator[i].Status = Locked
	}
	for i := 0; i < len(sigma); i++ {
		id := int(sigma[i] / sizePolynomials)
		od := sigma[i] % sizePolynomials
		in := int(i / sizePolynomials)
		on := i % sizePolynomials
		denominator[in].Coefficients[on].Set(&numerator[id].Coefficients[od])
	}

	return numerator, denominator, sigma

}

func TestBuildRatioShuffledVectors(t *testing.T) {

	// generate random vectors, interpreted in Lagrange form,
	// regular layout. It is enough for this test if TestPutInLagrangeForm
	// passes.
	sizePolynomials := 8
	nbPolynomials := 4
	numerator, denominator, _ := getPermutedPolynomials(sizePolynomials, nbPolynomials)

	// build the ratio polynomial
	expectedForm := Form{Basis: Lagrange, Layout: Regular, Status: Unlocked}
	domain := fft.NewDomain(uint64(sizePolynomials))
	var beta fr.Element
	beta.SetRandom()
	ratio, err := BuildRatioShuffledVectors(numerator, denominator, beta, expectedForm, domain)
	if err != nil {
		t.Fatal()
	}

	// check that the whole product is equal to one
	var a, b, c, d fr.Element
	b.SetOne()
	d.SetOne()
	for i := 0; i < nbPolynomials; i++ {
		a.Sub(&beta, &numerator[i].Coefficients[sizePolynomials-1])
		b.Mul(&a, &b)
		c.Sub(&beta, &denominator[i].Coefficients[sizePolynomials-1])
		d.Mul(&c, &d)
	}
	a.Mul(&b, &ratio.Coefficients[sizePolynomials-1]).
		Div(&a, &d)
	var one fr.Element
	one.SetOne()
	if !a.Equal(&one) {
		t.Fatal("accumulating ratio is not equal to one")
	}

	// check that the ratio is correct when the inputs are
	// bit reversed
	for i := 0; i < nbPolynomials; i++ {
		fft.BitReverse(numerator[i].Coefficients)
		numerator[i].Layout = BitReverse
		fft.BitReverse(denominator[i].Coefficients)
		denominator[i].Layout = BitReverse
	}
	{
		var err error
		_ratio, err := BuildRatioShuffledVectors(numerator, denominator, beta, expectedForm, domain)
		if err != nil {
			t.Fatal(err)
		}
		checkCoeffs := cmpCoefficents(_ratio.Coefficients, ratio.Coefficients)
		if !checkCoeffs {
			t.Fatal(err)
		}
	}

	// check that the ratio is correct when the inputs are in
	// canonical form, regular
	for i := 0; i < nbPolynomials; i++ {
		domain.FFTInverse(numerator[i].Coefficients, fft.DIT)
		numerator[i].Basis = Canonical
		numerator[i].Layout = Regular
		domain.FFTInverse(denominator[i].Coefficients, fft.DIT)
		denominator[i].Basis = Canonical
		denominator[i].Layout = Regular
	}
	{
		var err error
		_ratio, err := BuildRatioShuffledVectors(numerator, denominator, beta, expectedForm, domain)
		if err != nil {
			t.Fatal(err)
		}
		checkCoeffs := cmpCoefficents(_ratio.Coefficients, ratio.Coefficients)
		if !checkCoeffs {
			t.Fatal("coefficients of ratio are not consistent")
		}
	}

	// check that the ratio is correct when the inputs are in
	// canonical form, bit reverse
	for i := 0; i < nbPolynomials; i++ {
		fft.BitReverse(numerator[i].Coefficients)
		numerator[i].Layout = BitReverse
		fft.BitReverse(denominator[i].Coefficients)
		denominator[i].Layout = BitReverse
	}

	{
		var err error
		_ratio, err := BuildRatioShuffledVectors(numerator, denominator, beta, expectedForm, domain)
		if err != nil {
			t.Fatal(err)
		}
		checkCoeffs := cmpCoefficents(_ratio.Coefficients, ratio.Coefficients)
		if !checkCoeffs {
			t.Fatal("coefficients of ratio are not consistent")
		}
	}

}

// sizePolynomial*nbPolynomial must be divisible by 2.
// The function generates a list of nbPolynomials (P_i) of size n=sizePolynomials
// such that [P₁ ∥ .. ∥ P₂ ] is invariant under the permutation
// σ defined by:
// σ = (12)(34)..(2n-1 2n)
// so σ is a product of cycles length 2.
func getInvariantEntriesUnderPermutation(sizePolynomials, nbPolynomials int) ([]Polynomial, []int64) {
	res := make([]Polynomial, nbPolynomials)
	form := Form{Layout: Regular, Basis: Lagrange, Status: Locked}
	for i := 0; i < nbPolynomials; i++ {
		// tmp := make([]fr.Element, sizePolynomials)
		// res[i] = &Polynomial{Coefficients: tmp, Form: form}
		res[i].Form = form
		res[i].Coefficients = make([]fr.Element, sizePolynomials)
		for j := 0; j < sizePolynomials/2; j++ {
			res[i].Coefficients[2*j].SetRandom()
			res[i].Coefficients[2*j+1].Set(&res[i].Coefficients[2*j])
		}
	}
	permutation := make([]int64, nbPolynomials*sizePolynomials)
	for i := int64(0); i < int64(nbPolynomials*sizePolynomials/2); i++ {
		permutation[2*i] = 2*i + 1
		permutation[2*i+1] = 2 * i
	}
	return res, permutation
}

func TestBuildRatioCopyConstraint(t *testing.T) {

	// generate random vectors, interpreted in Lagrange form,
	// regular layout. It is enough for this test if TestPutInLagrangeForm
	// passes.
	sizePolynomials := 8
	nbPolynomials := 4
	entries, sigma := getInvariantEntriesUnderPermutation(sizePolynomials, nbPolynomials)

	// build the ratio polynomial
	expectedForm := Form{Basis: Lagrange, Layout: Regular, Status: Unlocked}
	domain := fft.NewDomain(uint64(sizePolynomials))
	var beta, gamma fr.Element
	beta.SetRandom()
	gamma.SetRandom()
	ratio, err := BuildRatioCopyConstraint(entries, sigma, beta, gamma, expectedForm, domain)
	if err != nil {
		t.Fatal()
	}

	// check that the whole product is equal to one
	suppID := getSupportIdentityPermutation(nbPolynomials, domain)
	var a, b, c, d fr.Element
	b.SetOne()
	d.SetOne()
	for i := 0; i < nbPolynomials; i++ {
		a.Mul(&beta, &suppID[(i+1)*sizePolynomials-1]).
			Add(&a, &entries[i].Coefficients[sizePolynomials-1]).
			Add(&a, &gamma)
		b.Mul(&b, &a)

		c.Mul(&beta, &suppID[sigma[(i+1)*sizePolynomials-1]]).
			Add(&c, &entries[i].Coefficients[sizePolynomials-1]).
			Add(&c, &gamma)
		d.Mul(&d, &c)
	}
	a.Mul(&b, &ratio.Coefficients[sizePolynomials-1]).
		Div(&a, &d)
	var one fr.Element
	one.SetOne()
	if !a.Equal(&one) {
		t.Fatal("accumulating ratio is not equal to one")
	}

	// check that the ratio is correct when the inputs are
	// bit reversed
	for i := 0; i < nbPolynomials; i++ {
		fft.BitReverse(entries[i].Coefficients)
		entries[i].Layout = BitReverse
	}
	{
		var err error
		_ratio, err := BuildRatioCopyConstraint(entries, sigma, beta, gamma, expectedForm, domain)
		if err != nil {
			t.Fatal(err)
		}
		checkCoeffs := cmpCoefficents(_ratio.Coefficients, ratio.Coefficients)
		if !checkCoeffs {
			t.Fatal(err)
		}
	}

	// check that the ratio is correct when the inputs are in
	// canonical form, regular
	for i := 0; i < nbPolynomials; i++ {
		domain.FFTInverse(entries[i].Coefficients, fft.DIT)
		entries[i].Basis = Canonical
		entries[i].Layout = Regular
	}
	{
		var err error
		_ratio, err := BuildRatioCopyConstraint(entries, sigma, beta, gamma, expectedForm, domain)
		if err != nil {
			t.Fatal(err)
		}
		checkCoeffs := cmpCoefficents(_ratio.Coefficients, ratio.Coefficients)
		if !checkCoeffs {
			t.Fatal("coefficients of ratio are not consistent")
		}
	}

	// check that the ratio is correct when the inputs are in
	// canonical form, bit reverse
	for i := 0; i < nbPolynomials; i++ {
		fft.BitReverse(entries[i].Coefficients)
		entries[i].Layout = BitReverse
	}

	{
		var err error
		_ratio, err := BuildRatioCopyConstraint(entries, sigma, beta, gamma, expectedForm, domain)
		if err != nil {
			t.Fatal(err)
		}
		checkCoeffs := cmpCoefficents(_ratio.Coefficients, ratio.Coefficients)
		if !checkCoeffs {
			t.Fatal("coefficients of ratio are not consistent")
		}
	}
}
