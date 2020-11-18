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

// Code generated by gurvy DO NOT EDIT

package fptower

import (
	"encoding/binary"
	"errors"
	"github.com/consensys/gurvy/bls381/fp"
	"math/big"
)

// E12 is a degree two finite field extension of fp6
type E12 struct {
	C0, C1 E6
}

// Equal returns true if z equals x, fasle otherwise
func (z *E12) Equal(x *E12) bool {
	return z.C0.Equal(&x.C0) && z.C1.Equal(&x.C1)
}

// String puts E12 in string form
func (z *E12) String() string {
	return (z.C0.String() + "+(" + z.C1.String() + ")*w")
}

// SetString sets a E12 from string
func (z *E12) SetString(s0, s1, s2, s3, s4, s5, s6, s7, s8, s9, s10, s11 string) *E12 {
	z.C0.SetString(s0, s1, s2, s3, s4, s5)
	z.C1.SetString(s6, s7, s8, s9, s10, s11)
	return z
}

// Set copies x into z and returns z
func (z *E12) Set(x *E12) *E12 {
	z.C0 = x.C0
	z.C1 = x.C1
	return z
}

// SetOne sets z to 1 in Montgomery form and returns z
func (z *E12) SetOne() *E12 {
	*z = E12{}
	z.C0.B0.A0.SetOne()
	return z
}

// ToMont converts to Mont form
func (z *E12) ToMont() *E12 {
	z.C0.ToMont()
	z.C1.ToMont()
	return z
}

// FromMont converts from Mont form
func (z *E12) FromMont() *E12 {
	z.C0.FromMont()
	z.C1.FromMont()
	return z
}

// Add set z=x+y in E12 and return z
func (z *E12) Add(x, y *E12) *E12 {
	z.C0.Add(&x.C0, &y.C0)
	z.C1.Add(&x.C1, &y.C1)
	return z
}

// Sub sets z to x sub y and return z
func (z *E12) Sub(x, y *E12) *E12 {
	z.C0.Sub(&x.C0, &y.C0)
	z.C1.Sub(&x.C1, &y.C1)
	return z
}

// Double sets z=2*x and returns z
func (z *E12) Double(x *E12) *E12 {
	z.C0.Double(&x.C0)
	z.C1.Double(&x.C1)
	return z
}

// SetRandom used only in tests
func (z *E12) SetRandom() (*E12, error) {
	if _, err := z.C0.SetRandom(); err != nil {
		return nil, err
	}
	if _, err := z.C1.SetRandom(); err != nil {
		return nil, err
	}
	return z, nil
}

// Mul set z=x*y in E12 and return z
func (z *E12) Mul(x, y *E12) *E12 {
	var a, b, c E6
	a.Add(&x.C0, &x.C1)
	b.Add(&y.C0, &y.C1)
	a.Mul(&a, &b)
	b.Mul(&x.C0, &y.C0)
	c.Mul(&x.C1, &y.C1)
	z.C1.Sub(&a, &b).Sub(&z.C1, &c)
	z.C0.MulByNonResidue(&c).Add(&z.C0, &b)
	return z
}

// Square set z=x*x in E12 and return z
func (z *E12) Square(x *E12) *E12 {

	//Algorithm 22 from https://eprint.iacr.org/2010/354.pdf
	var c0, c2, c3 E6
	c0.Sub(&x.C0, &x.C1)
	c3.MulByNonResidue(&x.C1).Neg(&c3).Add(&x.C0, &c3)
	c2.Mul(&x.C0, &x.C1)
	c0.Mul(&c0, &c3).Add(&c0, &c2)
	z.C1.Double(&c2)
	c2.MulByNonResidue(&c2)
	z.C0.Add(&c0, &c2)

	return z
}

// squares an element a+by interpreted as an Fp4 elmt, where y**2= non_residue_e2
func fp4Square(a, b, c, d *E2) {
	var tmp E2
	c.Square(a)
	tmp.Square(b).MulByNonResidue(&tmp)
	c.Add(c, &tmp)
	d.Mul(a, b).Double(d)
}

// CyclotomicSquare https://eprint.iacr.org/2009/565.pdf, 3.2
func (z *E12) CyclotomicSquare(x *E12) *E12 {

	var res, b, a E12
	var tmp E2

	// A
	fp4Square(&x.C0.B0, &x.C1.B1, &b.C0.B0, &b.C1.B1)
	a.C0.B0.Set(&x.C0.B0)
	a.C1.B1.Neg(&x.C1.B1)

	// B
	tmp.MulByNonResidueInv(&x.C1.B0)
	fp4Square(&x.C0.B2, &tmp, &b.C0.B1, &b.C1.B2)
	b.C0.B1.MulByNonResidue(&b.C0.B1)
	b.C1.B2.MulByNonResidue(&b.C1.B2)
	a.C0.B1.Set(&x.C0.B1)
	a.C1.B2.Neg(&x.C1.B2)

	// C
	fp4Square(&x.C0.B1, &x.C1.B2, &b.C0.B2, &b.C1.B0)
	b.C1.B0.MulByNonResidue(&b.C1.B0)
	a.C0.B2.Set(&x.C0.B2)
	a.C1.B0.Neg(&x.C1.B0)

	res.Set(&b)
	b.Sub(&b, &a).Double(&b)
	z.Add(&res, &b)

	return z
}

// Inverse set z to the inverse of x in E12 and return z
func (z *E12) Inverse(x *E12) *E12 {
	// Algorithm 23 from https://eprint.iacr.org/2010/354.pdf

	var t0, t1, tmp E6
	t0.Square(&x.C0)
	t1.Square(&x.C1)
	tmp.MulByNonResidue(&t1)
	t0.Sub(&t0, &tmp)
	t1.Inverse(&t0)
	z.C0.Mul(&x.C0, &t1)
	z.C1.Mul(&x.C1, &t1).Neg(&z.C1)

	return z
}

// Exp sets z=x**e and returns it
func (z *E12) Exp(x *E12, e big.Int) *E12 {
	var res E12
	res.SetOne()
	b := e.Bytes()
	for i := range b {
		w := b[i]
		mask := byte(0x80)
		for j := 7; j >= 0; j-- {
			res.Square(&res)
			if (w&mask)>>j != 0 {
				res.Mul(&res, x)
			}
			mask = mask >> 1
		}
	}
	z.Set(&res)
	return z
}

// InverseUnitary inverse a unitary element
func (z *E12) InverseUnitary(x *E12) *E12 {
	return z.Conjugate(x)
}

// Conjugate set z to x conjugated and return z
func (z *E12) Conjugate(x *E12) *E12 {
	*z = *x
	z.C1.Neg(&z.C1)
	return z
}

// SizeOfGT represents the size in bytes that a GT element need in binary form
const SizeOfGT = 48 * 12

// Bytes returns the regular (non montgomery) value
// of z as a big-endian byte array.
// z.C1.B2.A1 | z.C1.B2.A0 | z.C1.B1.A1 | ...
func (z *E12) Bytes() (r [SizeOfGT]byte) {
	_z := *z
	_z.FromMont()
	binary.BigEndian.PutUint64(r[568:576], _z.C0.B0.A0[0])
	binary.BigEndian.PutUint64(r[560:568], _z.C0.B0.A0[1])
	binary.BigEndian.PutUint64(r[552:560], _z.C0.B0.A0[2])
	binary.BigEndian.PutUint64(r[544:552], _z.C0.B0.A0[3])
	binary.BigEndian.PutUint64(r[536:544], _z.C0.B0.A0[4])
	binary.BigEndian.PutUint64(r[528:536], _z.C0.B0.A0[5])

	binary.BigEndian.PutUint64(r[520:528], _z.C0.B0.A1[0])
	binary.BigEndian.PutUint64(r[512:520], _z.C0.B0.A1[1])
	binary.BigEndian.PutUint64(r[504:512], _z.C0.B0.A1[2])
	binary.BigEndian.PutUint64(r[496:504], _z.C0.B0.A1[3])
	binary.BigEndian.PutUint64(r[488:496], _z.C0.B0.A1[4])
	binary.BigEndian.PutUint64(r[480:488], _z.C0.B0.A1[5])

	binary.BigEndian.PutUint64(r[472:480], _z.C0.B1.A0[0])
	binary.BigEndian.PutUint64(r[464:472], _z.C0.B1.A0[1])
	binary.BigEndian.PutUint64(r[456:464], _z.C0.B1.A0[2])
	binary.BigEndian.PutUint64(r[448:456], _z.C0.B1.A0[3])
	binary.BigEndian.PutUint64(r[440:448], _z.C0.B1.A0[4])
	binary.BigEndian.PutUint64(r[432:440], _z.C0.B1.A0[5])

	binary.BigEndian.PutUint64(r[424:432], _z.C0.B1.A1[0])
	binary.BigEndian.PutUint64(r[416:424], _z.C0.B1.A1[1])
	binary.BigEndian.PutUint64(r[408:416], _z.C0.B1.A1[2])
	binary.BigEndian.PutUint64(r[400:408], _z.C0.B1.A1[3])
	binary.BigEndian.PutUint64(r[392:400], _z.C0.B1.A1[4])
	binary.BigEndian.PutUint64(r[384:392], _z.C0.B1.A1[5])

	binary.BigEndian.PutUint64(r[376:384], _z.C0.B2.A0[0])
	binary.BigEndian.PutUint64(r[368:376], _z.C0.B2.A0[1])
	binary.BigEndian.PutUint64(r[360:368], _z.C0.B2.A0[2])
	binary.BigEndian.PutUint64(r[352:360], _z.C0.B2.A0[3])
	binary.BigEndian.PutUint64(r[344:352], _z.C0.B2.A0[4])
	binary.BigEndian.PutUint64(r[336:344], _z.C0.B2.A0[5])

	binary.BigEndian.PutUint64(r[328:336], _z.C0.B2.A1[0])
	binary.BigEndian.PutUint64(r[320:328], _z.C0.B2.A1[1])
	binary.BigEndian.PutUint64(r[312:320], _z.C0.B2.A1[2])
	binary.BigEndian.PutUint64(r[304:312], _z.C0.B2.A1[3])
	binary.BigEndian.PutUint64(r[296:304], _z.C0.B2.A1[4])
	binary.BigEndian.PutUint64(r[288:296], _z.C0.B2.A1[5])

	binary.BigEndian.PutUint64(r[280:288], _z.C1.B0.A0[0])
	binary.BigEndian.PutUint64(r[272:280], _z.C1.B0.A0[1])
	binary.BigEndian.PutUint64(r[264:272], _z.C1.B0.A0[2])
	binary.BigEndian.PutUint64(r[256:264], _z.C1.B0.A0[3])
	binary.BigEndian.PutUint64(r[248:256], _z.C1.B0.A0[4])
	binary.BigEndian.PutUint64(r[240:248], _z.C1.B0.A0[5])

	binary.BigEndian.PutUint64(r[232:240], _z.C1.B0.A1[0])
	binary.BigEndian.PutUint64(r[224:232], _z.C1.B0.A1[1])
	binary.BigEndian.PutUint64(r[216:224], _z.C1.B0.A1[2])
	binary.BigEndian.PutUint64(r[208:216], _z.C1.B0.A1[3])
	binary.BigEndian.PutUint64(r[200:208], _z.C1.B0.A1[4])
	binary.BigEndian.PutUint64(r[192:200], _z.C1.B0.A1[5])

	binary.BigEndian.PutUint64(r[184:192], _z.C1.B1.A0[0])
	binary.BigEndian.PutUint64(r[176:184], _z.C1.B1.A0[1])
	binary.BigEndian.PutUint64(r[168:176], _z.C1.B1.A0[2])
	binary.BigEndian.PutUint64(r[160:168], _z.C1.B1.A0[3])
	binary.BigEndian.PutUint64(r[152:160], _z.C1.B1.A0[4])
	binary.BigEndian.PutUint64(r[144:152], _z.C1.B1.A0[5])

	binary.BigEndian.PutUint64(r[136:144], _z.C1.B1.A1[0])
	binary.BigEndian.PutUint64(r[128:136], _z.C1.B1.A1[1])
	binary.BigEndian.PutUint64(r[120:128], _z.C1.B1.A1[2])
	binary.BigEndian.PutUint64(r[112:120], _z.C1.B1.A1[3])
	binary.BigEndian.PutUint64(r[104:112], _z.C1.B1.A1[4])
	binary.BigEndian.PutUint64(r[96:104], _z.C1.B1.A1[5])

	binary.BigEndian.PutUint64(r[88:96], _z.C1.B2.A0[0])
	binary.BigEndian.PutUint64(r[80:88], _z.C1.B2.A0[1])
	binary.BigEndian.PutUint64(r[72:80], _z.C1.B2.A0[2])
	binary.BigEndian.PutUint64(r[64:72], _z.C1.B2.A0[3])
	binary.BigEndian.PutUint64(r[56:64], _z.C1.B2.A0[4])
	binary.BigEndian.PutUint64(r[48:56], _z.C1.B2.A0[5])

	binary.BigEndian.PutUint64(r[40:48], _z.C1.B2.A1[0])
	binary.BigEndian.PutUint64(r[32:40], _z.C1.B2.A1[1])
	binary.BigEndian.PutUint64(r[24:32], _z.C1.B2.A1[2])
	binary.BigEndian.PutUint64(r[16:24], _z.C1.B2.A1[3])
	binary.BigEndian.PutUint64(r[8:16], _z.C1.B2.A1[4])
	binary.BigEndian.PutUint64(r[0:8], _z.C1.B2.A1[5])

	return
}

// SetBytes interprets e as the bytes of a big-endian GT
// sets z to that value (in Montgomery form), and returns z.
// size(e) == 48 * 12
// z.C1.B2.A1 | z.C1.B2.A0 | z.C1.B1.A1 | ...
func (z *E12) SetBytes(e []byte) error {
	if len(e) != SizeOfGT {
		return errors.New("invalid buffer size")
	}
	z.C0.B0.A0.SetBytes(e[528 : 528+fp.Bytes])

	z.C0.B0.A1.SetBytes(e[480 : 480+fp.Bytes])

	z.C0.B1.A0.SetBytes(e[432 : 432+fp.Bytes])

	z.C0.B1.A1.SetBytes(e[384 : 384+fp.Bytes])

	z.C0.B2.A0.SetBytes(e[336 : 336+fp.Bytes])

	z.C0.B2.A1.SetBytes(e[288 : 288+fp.Bytes])

	z.C1.B0.A0.SetBytes(e[240 : 240+fp.Bytes])

	z.C1.B0.A1.SetBytes(e[192 : 192+fp.Bytes])

	z.C1.B1.A0.SetBytes(e[144 : 144+fp.Bytes])

	z.C1.B1.A1.SetBytes(e[96 : 96+fp.Bytes])

	z.C1.B2.A0.SetBytes(e[48 : 48+fp.Bytes])

	z.C1.B2.A1.SetBytes(e[0 : 0+fp.Bytes])

	return nil
}