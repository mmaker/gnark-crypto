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

package fptower

import (
	"math/big"
	"testing"

	"github.com/consensys/gnark-crypto/ecc/bls12-378/fp"
	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/prop"
)

// ------------------------------------------------------------
// tests

func TestE12Serialization(t *testing.T) {

	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100

	properties := gopter.NewProperties(parameters)

	genA := GenE12()

	properties.Property("[BLS12-378] SetBytes(Bytes()) should stay constant", prop.ForAll(
		func(a *E12) bool {
			var b E12
			buf := a.Bytes()
			if err := b.SetBytes(buf[:]); err != nil {
				return false
			}
			return a.Equal(&b)
		},
		genA,
	))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}

func TestE12ReceiverIsOperand(t *testing.T) {

	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100

	properties := gopter.NewProperties(parameters)

	genA := GenE12()
	genB := GenE12()

	properties.Property("[BLS12-378] Having the receiver as operand (addition) should output the same result", prop.ForAll(
		func(a, b *E12) bool {
			var c, d E12
			d.Set(a)
			c.Add(a, b)
			a.Add(a, b)
			b.Add(&d, b)
			return a.Equal(b) && a.Equal(&c) && b.Equal(&c)
		},
		genA,
		genB,
	))

	properties.Property("[BLS12-378] Having the receiver as operand (sub) should output the same result", prop.ForAll(
		func(a, b *E12) bool {
			var c, d E12
			d.Set(a)
			c.Sub(a, b)
			a.Sub(a, b)
			b.Sub(&d, b)
			return a.Equal(b) && a.Equal(&c) && b.Equal(&c)
		},
		genA,
		genB,
	))

	properties.Property("[BLS12-378] Having the receiver as operand (mul) should output the same result", prop.ForAll(
		func(a, b *E12) bool {
			var c, d E12
			d.Set(a)
			c.Mul(a, b)
			a.Mul(a, b)
			b.Mul(&d, b)
			return a.Equal(b) && a.Equal(&c) && b.Equal(&c)
		},
		genA,
		genB,
	))

	properties.Property("[BLS12-378] Having the receiver as operand (square) should output the same result", prop.ForAll(
		func(a *E12) bool {
			var b E12
			b.Square(a)
			a.Square(a)
			return a.Equal(&b)
		},
		genA,
	))

	properties.Property("[BLS12-378] Having the receiver as operand (double) should output the same result", prop.ForAll(
		func(a *E12) bool {
			var b E12
			b.Double(a)
			a.Double(a)
			return a.Equal(&b)
		},
		genA,
	))

	properties.Property("[BLS12-378] Having the receiver as operand (Inverse) should output the same result", prop.ForAll(
		func(a *E12) bool {
			var b E12
			b.Inverse(a)
			a.Inverse(a)
			return a.Equal(&b)
		},
		genA,
	))

	properties.Property("[BLS12-378] Having the receiver as operand (Cyclotomic square) should output the same result", prop.ForAll(
		func(a *E12) bool {
			var b E12
			b.CyclotomicSquare(a)
			a.CyclotomicSquare(a)
			return a.Equal(&b)
		},
		genA,
	))

	properties.Property("[BLS12-378] Having the receiver as operand (Conjugate) should output the same result", prop.ForAll(
		func(a *E12) bool {
			var b E12
			b.Conjugate(a)
			a.Conjugate(a)
			return a.Equal(&b)
		},
		genA,
	))

	properties.Property("[BLS12-378] Having the receiver as operand (Frobenius) should output the same result", prop.ForAll(
		func(a *E12) bool {
			var b E12
			b.Frobenius(a)
			a.Frobenius(a)
			return a.Equal(&b)
		},
		genA,
	))

	properties.Property("[BLS12-378] Having the receiver as operand (FrobeniusSquare) should output the same result", prop.ForAll(
		func(a *E12) bool {
			var b E12
			b.FrobeniusSquare(a)
			a.FrobeniusSquare(a)
			return a.Equal(&b)
		},
		genA,
	))

	properties.Property("[BLS12-378] Having the receiver as operand (FrobeniusCube) should output the same result", prop.ForAll(
		func(a *E12) bool {
			var b E12
			b.FrobeniusCube(a)
			a.FrobeniusCube(a)
			return a.Equal(&b)
		},
		genA,
	))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}

func TestE12Ops(t *testing.T) {

	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100

	properties := gopter.NewProperties(parameters)

	genA := GenE12()
	genB := GenE12()
	genExp := GenFp()

	properties.Property("[BLS12-378] sub & add should leave an element invariant", prop.ForAll(
		func(a, b *E12) bool {
			var c E12
			c.Set(a)
			c.Add(&c, b).Sub(&c, b)
			return c.Equal(a)
		},
		genA,
		genB,
	))

	properties.Property("[BLS12-378] mul & inverse should leave an element invariant", prop.ForAll(
		func(a, b *E12) bool {
			var c, d E12
			d.Inverse(b)
			c.Set(a)
			c.Mul(&c, b).Mul(&c, &d)
			return c.Equal(a)
		},
		genA,
		genB,
	))

	properties.Property("[BLS12-378] inverse twice should leave an element invariant", prop.ForAll(
		func(a *E12) bool {
			var b E12
			b.Inverse(a).Inverse(&b)
			return a.Equal(&b)
		},
		genA,
	))

	properties.Property("[BLS12-378] square and mul should output the same result", prop.ForAll(
		func(a *E12) bool {
			var b, c E12
			b.Mul(a, a)
			c.Square(a)
			return b.Equal(&c)
		},
		genA,
	))

	properties.Property("[BLS12-378] a + pi(a), a-pi(a) should be real", prop.ForAll(
		func(a *E12) bool {
			var b, c, d E12
			var e, f, g E6
			b.Conjugate(a)
			c.Add(a, &b)
			d.Sub(a, &b)
			e.Double(&a.C0)
			f.Double(&a.C1)
			return c.C1.Equal(&g) && d.C0.Equal(&g) && e.Equal(&c.C0) && f.Equal(&d.C1)
		},
		genA,
	))

	properties.Property("[BLS12-378] pi**12=id", prop.ForAll(
		func(a *E12) bool {
			var b E12
			b.Frobenius(a).
				Frobenius(&b).
				Frobenius(&b).
				Frobenius(&b).
				Frobenius(&b).
				Frobenius(&b).
				Frobenius(&b).
				Frobenius(&b).
				Frobenius(&b).
				Frobenius(&b).
				Frobenius(&b).
				Frobenius(&b)
			return b.Equal(a)
		},
		genA,
	))

	properties.Property("[BLS12-378] (pi**2)**6=id", prop.ForAll(
		func(a *E12) bool {
			var b E12
			b.FrobeniusSquare(a).
				FrobeniusSquare(&b).
				FrobeniusSquare(&b).
				FrobeniusSquare(&b).
				FrobeniusSquare(&b).
				FrobeniusSquare(&b)
			return b.Equal(a)
		},
		genA,
	))

	properties.Property("[BLS12-378] (pi**3)**4=id", prop.ForAll(
		func(a *E12) bool {
			var b E12
			b.FrobeniusCube(a).
				FrobeniusCube(&b).
				FrobeniusCube(&b).
				FrobeniusCube(&b)
			return b.Equal(a)
		},
		genA,
	))

	properties.Property("[BLS12-378] cyclotomic square (Granger-Scott) and square should be the same in the cyclotomic subgroup", prop.ForAll(
		func(a *E12) bool {
			var b, c, d E12
			b.Conjugate(a)
			a.Inverse(a)
			b.Mul(&b, a)
			a.FrobeniusSquare(&b).Mul(a, &b)
			c.Square(a)
			d.CyclotomicSquare(a)
			return c.Equal(&d)
		},
		genA,
	))

	properties.Property("[BLS12-378] compressed cyclotomic square (Karabina) and square should be the same in the cyclotomic subgroup", prop.ForAll(
		func(a *E12) bool {
			var b, c, d E12
			b.Conjugate(a)
			a.Inverse(a)
			b.Mul(&b, a)
			a.FrobeniusSquare(&b).Mul(a, &b)
			c.Square(a)
			d.CyclotomicSquareCompressed(a).Decompress(&d)
			return c.Equal(&d)
		},
		genA,
	))

	properties.Property("[BLS12-378] batch decompress and individual decompress (Karabina) should be the same", prop.ForAll(
		func(a *E12) bool {
			var b E12
			// put in the cyclotomic subgroup
			b.Conjugate(a)
			a.Inverse(a)
			b.Mul(&b, a)
			a.FrobeniusSquare(&b).Mul(a, &b)

			var a2, a4, a17 E12
			a2.Set(a)
			a4.Set(a)
			a17.Set(a)
			a2.nSquareCompressed(2)
			a4.nSquareCompressed(4)
			a17.nSquareCompressed(17)
			batch := BatchDecompress([]E12{a2, a4, a17})
			a2.Decompress(&a2)
			a4.Decompress(&a4)
			a17.Decompress(&a17)

			return a2.Equal(&batch[0]) && a4.Equal(&batch[1]) && a17.Equal(&batch[2])
		},
		genA,
	))

	properties.Property("[BLS12-378] Exp and CyclotomicExp results must be the same in the cyclotomic subgroup", prop.ForAll(
		func(a *E12, e fp.Element) bool {
			var b, c, d E12
			// put in the cyclo subgroup
			b.Conjugate(a)
			a.Inverse(a)
			b.Mul(&b, a)
			a.FrobeniusSquare(&b).Mul(a, &b)

			var _e big.Int
			k := new(big.Int).SetUint64(12)
			e.Exp(e, k)
			e.ToBigIntRegular(&_e)

			c.Exp(a, _e)
			d.CyclotomicExp(a, _e)

			return c.Equal(&d)
		},
		genA,
		genExp,
	))

	properties.Property("[BLS12-378] Frobenius of x in E12 should be equal to x^q", prop.ForAll(
		func(a *E12) bool {
			var b, c E12
			q := fp.Modulus()
			b.Frobenius(a)
			c.Exp(a, *q)
			return c.Equal(&b)
		},
		genA,
	))

	properties.Property("[BLS12-378] FrobeniusSquare of x in E12 should be equal to x^(q^2)", prop.ForAll(
		func(a *E12) bool {
			var b, c E12
			q := fp.Modulus()
			b.FrobeniusSquare(a)
			c.Exp(a, *q).Exp(&c, *q)
			return c.Equal(&b)
		},
		genA,
	))

	properties.Property("[BLS12-378] FrobeniusCube of x in E12 should be equal to x^(q^3)", prop.ForAll(
		func(a *E12) bool {
			var b, c E12
			q := fp.Modulus()
			b.FrobeniusCube(a)
			c.Exp(a, *q).Exp(&c, *q).Exp(&c, *q)
			return c.Equal(&b)
		},
		genA,
	))

	properties.TestingRun(t, gopter.ConsoleReporter(false))

}

// ------------------------------------------------------------
// benches

func BenchmarkE12Add(b *testing.B) {
	var a, c E12
	a.SetRandom()
	c.SetRandom()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		a.Add(&a, &c)
	}
}

func BenchmarkE12Sub(b *testing.B) {
	var a, c E12
	a.SetRandom()
	c.SetRandom()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		a.Sub(&a, &c)
	}
}

func BenchmarkE12Mul(b *testing.B) {
	var a, c E12
	a.SetRandom()
	c.SetRandom()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		a.Mul(&a, &c)
	}
}

func BenchmarkE12Cyclosquare(b *testing.B) {
	var a E12
	a.SetRandom()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		a.CyclotomicSquare(&a)
	}
}

func BenchmarkE12Square(b *testing.B) {
	var a E12
	a.SetRandom()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		a.Square(&a)
	}
}

func BenchmarkE12Inverse(b *testing.B) {
	var a E12
	a.SetRandom()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		a.Inverse(&a)
	}
}

func BenchmarkE12Conjugate(b *testing.B) {
	var a E12
	a.SetRandom()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		a.Conjugate(&a)
	}
}

func BenchmarkE12Frobenius(b *testing.B) {
	var a E12
	a.SetRandom()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		a.Frobenius(&a)
	}
}

func BenchmarkE12FrobeniusSquare(b *testing.B) {
	var a E12
	a.SetRandom()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		a.FrobeniusSquare(&a)
	}
}

func BenchmarkE12FrobeniusCube(b *testing.B) {
	var a E12
	a.SetRandom()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		a.FrobeniusCube(&a)
	}
}

func BenchmarkE12Expt(b *testing.B) {
	var a E12
	a.SetRandom()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		a.Expt(&a)
	}
}

func TestE12Div(t *testing.T) {

	parameters := gopter.DefaultTestParameters()
	properties := gopter.NewProperties(parameters)

	genA := GenE12()
	genB := GenE12()

	properties.Property("[BLS12-378] dividing then multiplying by the same element does nothing", prop.ForAll(
		func(a, b *E12) bool {
			var c E12
			c.Div(a, b)
			c.Mul(&c, b)
			return c.Equal(a)
		},
		genA,
		genB,
	))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}
