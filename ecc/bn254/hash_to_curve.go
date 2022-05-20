// Copyright 2020 ConsenSys AG
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

package bn254

import (
	"math/big"

	"github.com/consensys/gnark-crypto/ecc"
	"github.com/consensys/gnark-crypto/ecc/bn254/fp"
	"github.com/consensys/gnark-crypto/ecc/bn254/internal/fptower"
)

// hashToFp hashes msg to count prime field elements.
// https://tools.ietf.org/html/draft-irtf-cfrg-hash-to-curve-06#section-5.2
func hashToFp(msg, dst []byte, count int) ([]fp.Element, error) {

	// 128 bits of security
	// L = ceil((ceil(log2(p)) + k) / 8), where k is the security parameter = 128
	L := 48

	lenInBytes := count * L
	pseudoRandomBytes, err := ecc.ExpandMsgXmd(msg, dst, lenInBytes)
	if err != nil {
		return nil, err
	}

	res := make([]fp.Element, count)
	for i := 0; i < count; i++ {
		res[i].SetBytes(pseudoRandomBytes[i*L : (i+1)*L])
	}
	return res, nil
}

// returns false if u>-u when seen as a bigInt
func sign0(u fp.Element) bool {
	var a, b big.Int
	u.ToBigIntRegular(&a)
	u.Neg(&u)
	u.ToBigIntRegular(&b)
	return a.Cmp(&b) <= 0
}

// ----------------------------------------------------------------------------------------
// G1Affine

// Fouque-Tibouchi method based on Shallue and van de Woestijne method,
// works for any elliptic curve in Weierstrass curve Y^2=X^3+B
// https://www.di.ens.fr/~fouque/pub/latincrypt12.pdf (section 3, defintion 2)
func svdwMapG1(u fp.Element) G1Affine {

	var res G1Affine

	// constants
	var x, y, c1, c2, one fp.Element
	c1.SetString("4407920970296243842837207485651524041948558517760411303933")
	c2.SetString("2203960485148121921418603742825762020974279258880205651966")
	one.SetOne()

	if u.IsZero() {
		x.Set(&c2)
		y.Add(&one, &bCurveCoeff).
			Sqrt(&y)

		res.X.Set(&x)
		res.Y.Set(&y)

		return res
	}

	var k, x1, x2, r, x3, fx1, fx2 fp.Element
	k.Square(&u).
		Add(&k, &bCurveCoeff).
		Add(&k, &one).
		Inverse(&k).
		Mul(&k, &u).
		Mul(&k, &c1)
	x1.Mul(&k, &u).
		Sub(&x1, &c2).
		Neg(&x1)
	x2.Add(&x1, &one).
		Neg(&x2)
	r.Square(&k).
		Inverse(&r)
	x3.Add(&r, &one)
	fx1.Square(&x1).
		Mul(&fx1, &x1).
		Add(&fx1, &bCurveCoeff)
	fx2.Square(&x2).
		Mul(&fx2, &x2).
		Add(&fx2, &bCurveCoeff)
	s1 := fx1.Legendre()
	s2 := fx2.Legendre() - s2
	x.Set(&x3)
	if s2 == 1 {
		x.Set(&x2)
	}
	if s1 == 2 {
		x.Set(&x1)
	}
	y.Square(&x).
		Mul(&y, &x).
		Add(&y, &bCurveCoeff)
	y.Sqrt(&y)
	u2 := sign0(u) && sign0(y)
	if !u2 {
		y.Neg(&y)
	}
	res.X.Set(&x)
	res.Y.Set(&y)

	return res
}

// MapToG1 maps an fp.Element to a point on the curve using the Shallue and van de Woestijne map
// https://tools.ietf.org/html/draft-irtf-cfrg-hash-to-curve-06#section-2.2.1
func MapToG1(t fp.Element) G1Affine {
	res := svdwMapG1(t)
	return res
}

// EncodeToG1 maps an fp.Element to a point on the curve using the Shallue and van de Woestijne map
// https://tools.ietf.org/html/draft-irtf-cfrg-hash-to-curve-06#section-2.2.2
func EncodeToG1(msg, dst []byte) (G1Affine, error) {
	var res G1Affine
	t, err := hashToFp(msg, dst, 1)
	if err != nil {
		return res, err
	}
	res = MapToG1(t[0])
	return res, nil
}

// HashToG1 maps an fp.Element to a point on the curve using the Shallue and van de Woestijne map
// https://tools.ietf.org/html/draft-irtf-cfrg-hash-to-curve-06#section-3
func HashToG1(msg, dst []byte) (G1Affine, error) {
	var res G1Affine
	u, err := hashToFp(msg, dst, 2)
	if err != nil {
		return res, err
	}
	Q0 := MapToG1(u[0])
	Q1 := MapToG1(u[1])
	var _Q0, _Q1, _res G1Jac
	_Q0.FromAffine(&Q0)
	_Q1.FromAffine(&Q1)
	_res.Set(&_Q1).AddAssign(&_Q0)
	res.FromJacobian(&_res)
	return res, nil
}

// ----------------------------------------------------------------------------------------
// G2Affine

// https://tools.ietf.org/html/draft-irtf-cfrg-hash-to-curve-06#section-4.1
// Shallue and van de Woestijne method, works for any elliptic curve in Weierstrass curve
func svdwMapG2(u fptower.E2) G2Affine {

	var res G2Affine

	// constants
	// sage script to find z: https://tools.ietf.org/html/draft-irtf-cfrg-hash-to-curve-06#appendix-E.1
	var z, c1, c2, c3, c4 fptower.E2
	z.A1.SetString("1")
	c1.A0.SetString("19485874751759354771024239261021720505790618469301721065564631296452457478373")
	c1.A1.SetString("266929791119991161246907387137283842545076965332900288569378510910307636689")
	c2.A1.SetString("10944121435919637611123202872628637544348155578648911831344518947322613104291")
	c3.A0.SetString("13617985070220897759416741581922326973608167195618746963957740978229330444385")
	c3.A1.SetString("6485072654231349560354894037339044590945718224403932749563131108378844487223")
	c4.A0.SetString("18685085378399381287283517099609868978155387573303020199856495763721534568303")
	c4.A1.SetString("355906388159988214995876516183045123393435953777200384759171347880410182252")

	var tv1, tv2, tv3, tv4, one, x1, gx1, x2, gx2, x3, x, gx, y fptower.E2
	one.SetOne()
	tv1.Square(&u).Mul(&tv1, &c1)
	tv2.Add(&one, &tv1)
	tv1.Sub(&one, &tv1)
	tv3.Mul(&tv2, &tv1).Inverse(&tv3)
	tv4.Mul(&u, &tv1)
	tv4.Mul(&tv4, &tv3)
	tv4.Mul(&tv4, &c3)
	x1.Sub(&c2, &tv4)
	gx1.Square(&x1)
	// 12. gx1 = gx1 + A
	gx1.Mul(&gx1, &x1)
	gx1.Add(&gx1, &bTwistCurveCoeff)
	e1 := gx1.Legendre()
	x2.Add(&c2, &tv4)
	gx2.Square(&x2)
	// 18. gx2 = gx2 + A
	gx2.Mul(&gx2, &x2)
	gx2.Add(&gx2, &bTwistCurveCoeff)
	e2 := gx2.Legendre() - e1 // 2 if is_square(gx2) AND NOT e1
	x3.Square(&tv2)
	x3.Mul(&x3, &tv3)
	x3.Square(&x3)
	x3.Mul(&x3, &c4)
	x3.Add(&x3, &z)
	if e1 == 1 {
		x.Set(&x1)
	} else {
		x.Set(&x3)
	}
	if e2 == 2 {
		x.Set(&x2)
	}
	gx.Square(&x)
	// gx = gx + A
	gx.Mul(&gx, &x)
	gx.Add(&gx, &bTwistCurveCoeff)
	y.Sqrt(&gx)
	e3 := sign0(u.A0) && sign0(y.A0)
	if !e3 {
		y.Neg(&y)
	}
	x.Set(&x)
	y.Set(&y)

	return res
}

// MapToG2 maps an fp.Element to a point on the curve using the Shallue and van de Woestijne map
// https://tools.ietf.org/html/draft-irtf-cfrg-hash-to-curve-06#section-2.2.1
func MapToG2(t fptower.E2) G2Affine {
	res := svdwMapG2(t)
	res.ClearCofactor(&res)
	return res
}

// EncodeToG2 maps an fp.Element to a point on the curve using the Shallue and van de Woestijne map
// https://tools.ietf.org/html/draft-irtf-cfrg-hash-to-curve-06#section-2.2.2
func EncodeToG2(msg, dst []byte) (G2Affine, error) {
	var res G2Affine
	_t, err := hashToFp(msg, dst, 2)
	if err != nil {
		return res, err
	}
	var t fptower.E2
	t.A0.Set(&_t[0])
	t.A1.Set(&_t[1])
	res = MapToG2(t)
	return res, nil
}

// HashToG2 maps an fp.Element to a point on the curve using the Shallue and van de Woestijne map
// https://tools.ietf.org/html/draft-irtf-cfrg-hash-to-curve-06#section-3
func HashToG2(msg, dst []byte) (G2Affine, error) {
	var res G2Affine
	u, err := hashToFp(msg, dst, 4)
	if err != nil {
		return res, err
	}
	var u0, u1 fptower.E2
	u0.A0.Set(&u[0])
	u0.A1.Set(&u[1])
	u1.A0.Set(&u[2])
	u1.A1.Set(&u[3])
	Q0 := MapToG2(u0)
	Q1 := MapToG2(u1)
	var _Q0, _Q1, _res G2Jac
	_Q0.FromAffine(&Q0)
	_Q1.FromAffine(&Q1)
	_res.Set(&_Q1).AddAssign(&_Q0)
	res.FromJacobian(&_res)
	return res, nil
}
