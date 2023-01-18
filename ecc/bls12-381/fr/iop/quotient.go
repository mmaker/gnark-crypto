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
	"math/big"
	"math/bits"

	"github.com/consensys/gnark-crypto/ecc/bls12-381/fr"
	"github.com/consensys/gnark-crypto/ecc/bls12-381/fr/fft"
)

// ComputeQuotient computes h(entries)/X^n-1
// * entries polynomials to on which h is evaluated, they must be in LagrangeCoset basis
// with the same layout
// * h multivariate polynomial
// * domains domains used for performing ffts
// The result is in Canonical Regular.
func ComputeQuotient(entries []WrappedPolynomial, h MultivariatePolynomial, domains [2]*fft.Domain) (Polynomial, error) {

	var quotientLagrangeCoset Polynomial

	// check that the sizes are consistent
	nbPolynomials := len(entries)
	n := len(entries[0].P.Coefficients)
	for i := 0; i < nbPolynomials; i++ {
		if len(entries[i].P.Coefficients) != n {
			return quotientLagrangeCoset, ErrInconsistentSize
		}
	}

	// check that the format are consistent
	for i := 0; i < nbPolynomials; i++ {
		if entries[i].P.Basis != LagrangeCoset {
			return quotientLagrangeCoset, ErrInconsistentFormat
		}
	}

	// prepare the evaluations of x^n-1 on the big domain's coset
	xnMinusOneInverseLagrangeCoset := evaluateXnMinusOneDomainBigCoset(domains)

	// compute \rho for all polynomials
	rho := make([]int, nbPolynomials)
	for i := 0; i < nbPolynomials; i++ {
		rho[i] = len(entries[i].P.Coefficients) / entries[i].Size
	}
	ratio := int(domains[1].Cardinality / domains[0].Cardinality)

	// compute the division. We take care of the indices of the
	// polnyomials which are bit reversed.
	// The result is temporarily stored in bit reversed Lagrange form,
	// before it is actually transformed into the expected format.
	nbEntries := len(entries)
	x := make([]fr.Element, nbEntries)

	nbElmtsExtended := int(domains[1].Cardinality)
	quotientLagrangeCoset.Coefficients = make([]fr.Element, nbElmtsExtended)

	nn := uint64(64 - bits.TrailingZeros(uint(nbElmtsExtended)))

	for i := 0; i < int(nbElmtsExtended); i++ {

		for j := 0; j < nbEntries; j++ {

			if entries[j].P.Layout == Regular {

				// take in account the fact that the polynomial might be shifted...
				x[j].Set(&entries[j].P.Coefficients[uint64((i+entries[j].Shift*rho[j]))%domains[1].Cardinality])

			} else {

				// take in account the fact that the polynomial might be shifted...
				iRev := bits.Reverse64(uint64((i+entries[j].Shift*rho[j]))%domains[1].Cardinality) >> nn
				x[j].Set(&entries[j].P.Coefficients[iRev])
			}

		}

		// evaluate h on x
		iRev := bits.Reverse64(uint64(i)) >> nn
		quotientLagrangeCoset.Coefficients[iRev] = h.EvaluateSinglePoint(x)

		// divide by x^n-1 evaluated on the correct point.
		quotientLagrangeCoset.Coefficients[iRev].
			Mul(&quotientLagrangeCoset.Coefficients[iRev], &xnMinusOneInverseLagrangeCoset[i%ratio])
	}

	// at this point, the result is in LagrangeCoset, bit reversed.
	// We put it in canonical form.
	quotientLagrangeCoset.Form = Form{LagrangeCoset, BitReverse}
	quotientLagrangeCoset.ToCanonical(&quotientLagrangeCoset, domains[1])

	return quotientLagrangeCoset, nil
}

// evaluateXnMinusOneDomainBigCoset evalutes Xᵐ-1 on DomainBig coset
func evaluateXnMinusOneDomainBigCoset(domains [2]*fft.Domain) []fr.Element {

	ratio := domains[1].Cardinality / domains[0].Cardinality

	res := make([]fr.Element, ratio)

	expo := big.NewInt(int64(domains[0].Cardinality))
	res[0].Exp(domains[1].FrMultiplicativeGen, expo)

	var t fr.Element
	t.Exp(domains[1].Generator, big.NewInt(int64(domains[0].Cardinality)))

	for i := 1; i < int(ratio); i++ {
		res[i].Mul(&res[i-1], &t)
	}

	var one fr.Element
	one.SetOne()
	for i := 0; i < int(ratio); i++ {
		res[i].Sub(&res[i], &one)
	}

	res = fr.BatchInvert(res)

	return res
}
