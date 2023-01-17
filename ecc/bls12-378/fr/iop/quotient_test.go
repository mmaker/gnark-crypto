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

	"github.com/consensys/gnark-crypto/ecc"
	"github.com/consensys/gnark-crypto/ecc/bls12-378/fr"
	"github.com/consensys/gnark-crypto/ecc/bls12-378/fr/fft"
)

// func printLayout(f Form) {

// 	if f.Basis == Canonical {
// 		fmt.Printf("CANONICAL")
// 	} else if f.Basis == LagrangeCoset {
// 		fmt.Printf("LAGRANGE_COSET")
// 	} else {
// 		fmt.Printf("LAGRANGE")
// 	}
// 	fmt.Println("")

// 	if f.Layout == Regular {
// 		fmt.Printf("REGULAR")
// 	} else {
// 		fmt.Printf("BIT REVERSED")
// 	}
// 	fmt.Println("")

// }

// computes x₃ in h(x₁,x₂,x₃) = x₁^{2}*x₂ + x₃ - x₁^{3}
// from x₁ and x₂.
func computex3(x []fr.Element) fr.Element {

	var a, b fr.Element
	a.Square(&x[0]).Mul(&a, &x[1])
	b.Square(&x[0]).Mul(&b, &x[0])
	a.Sub(&b, &a)
	return a

}

func buildPoly(size int, form Form) WrappedPolynomial {
	var f Polynomial
	f.Coefficients = make([]fr.Element, size)
	f.Basis = form.Basis
	f.Layout = form.Layout
	return WrappedPolynomial{P: &f, Shift: 0, Size: size}
}

func evalCanonical(p Polynomial, x fr.Element) fr.Element {

	var res fr.Element
	for i := len(p.Coefficients) - 1; i >= 0; i-- {
		res.Mul(&res, &x)
		res.Add(&res, &p.Coefficients[i])
	}
	return res
}

func TestQuotient(t *testing.T) {

	// create the multivariate polynomial h
	// h(x₁,x₂,x₃) = x₁^{2}*x₂ + x₃ - x₁^{3}
	nbEntries := 3

	var h MultivariatePolynomial
	var one, minusOne fr.Element
	one.SetOne()
	minusOne.SetOne().Neg(&minusOne)
	h.AddMonomial(one, []int{2, 1, 0})
	h.AddMonomial(one, []int{0, 0, 1})
	h.AddMonomial(minusOne, []int{3, 0, 0})

	// create an instance (f_i) where h holds
	sizeSystem := 8

	form := Form{Basis: Lagrange, Layout: Regular}

	entries := make([]WrappedPolynomial, nbEntries)
	entries[0] = buildPoly(sizeSystem, form)
	entries[1] = buildPoly(sizeSystem, form)
	entries[2] = buildPoly(sizeSystem, form)

	for i := 0; i < sizeSystem; i++ {

		entries[0].P.Coefficients[i].SetRandom()
		entries[1].P.Coefficients[i].SetRandom()
		tmp := computex3(
			[]fr.Element{entries[0].P.Coefficients[i],
				entries[1].P.Coefficients[i]})
		entries[2].P.Coefficients[i].Set(&tmp)

		x := []fr.Element{
			entries[0].P.Coefficients[i],
			entries[1].P.Coefficients[i],
			entries[2].P.Coefficients[i],
		}
		tmp = h.EvaluateSinglePoint(x)
		if !tmp.IsZero() {
			t.Fatal("system does not vanish on x^n-1")
		}
	}

	// compute the quotient where the entries are in Regular layout
	var domains [2]*fft.Domain
	domains[0] = fft.NewDomain(uint64(sizeSystem))
	domains[1] = fft.NewDomain(ecc.NextPowerOfTwo(h.Degree() * domains[0].Cardinality))

	entries[0].P.ToCanonical(entries[0].P, domains[0]).
		ToRegular(entries[0].P).
		ToLagrangeCoset(entries[0].P, domains[1]).
		ToRegular(entries[0].P)

	entries[1].P.ToCanonical(entries[1].P, domains[0]).
		ToRegular(entries[1].P).
		ToLagrangeCoset(entries[1].P, domains[1]).
		ToRegular(entries[1].P)

	entries[2].P.ToCanonical(entries[2].P, domains[0]).
		ToRegular(entries[2].P).
		ToLagrangeCoset(entries[2].P, domains[1]).
		ToRegular(entries[2].P)

	quotientA, err := ComputeQuotient(entries, h, domains)
	if err != nil {
		t.Fatal(err)
	}
	quotientA.ToRegular(&quotientA)

	// compute the quotient where some entries are in BitReverse layout
	// (only the 2 first entries)
	entries[0].P.ToBitreverse(entries[0].P)
	entries[1].P.ToBitreverse(entries[1].P)

	quotientB, err := ComputeQuotient(entries, h, domains)
	if err != nil {
		t.Fatal(err)
	}

	// check that the two results are the same
	quotientB.ToRegular(&quotientB)
	if quotientB.Form != quotientA.Form {
		t.Fatal("quotient is inconsistent when entries are in bitRerverse and Regular")
	}
	for i := 0; i < int(domains[1].Cardinality); i++ {
		if !quotientA.Coefficients[i].Equal(&quotientB.Coefficients[i]) {
			t.Fatal("quotient is inconsistent when entries are in bitRerverse and Regular")
		}
	}

	// checks that h(f_i) = (x^n-1)*q by evaluating the relation
	// at a random point
	var c fr.Element
	c.SetRandom()

	entries[0].P.ToCanonical(entries[0].P, domains[1])
	entries[0].P.ToRegular(entries[0].P)

	entries[1].P.ToCanonical(entries[1].P, domains[1])
	entries[1].P.ToRegular(entries[1].P)

	entries[2].P.ToCanonical(entries[2].P, domains[1])
	entries[2].P.ToRegular(entries[2].P)

	x := []fr.Element{
		evalCanonical(*entries[0].P, c),
		evalCanonical(*entries[1].P, c),
		evalCanonical(*entries[2].P, c),
	}
	l := h.EvaluateSinglePoint(x)
	var xnminusone fr.Element
	xnminusone.Set(&c).
		Square(&xnminusone).
		Square(&xnminusone).
		Square(&xnminusone).
		Sub(&xnminusone, &one)
	r := evalCanonical(quotientA, c)
	r.Mul(&r, &xnminusone)

	if !r.Equal(&l) {
		t.Fatal("error quotient")
	}

}
