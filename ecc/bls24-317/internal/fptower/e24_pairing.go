package fptower

func (z *E24) nSquare(n int) {
	for i := 0; i < n; i++ {
		z.CyclotomicSquare(z)
	}
}

func (z *E24) nSquareCompressed(n int) {
	for i := 0; i < n; i++ {
		z.CyclotomicSquareCompressed(z)
	}
}

// Expt set z to x^t in E24 and return z (t is the seed of the curve)
// t = 3640754176
func (z *E24) Expt(x *E24) *E24 {
	// Expt computation is derived from the addition chain:
	//
	//	_10       = 2*1
	//	_11       = 1 + _10
	//	_11000    = _11 << 3
	//	_11000000 = _11000 << 3
	//	_11011000 = _11000 + _11000000
	//	_11011001 = 1 + _11011000
	//	return      (_11011001 << 9 + _11) << 15
	//
	// Operations: 31 squares 4 multiplies
	//
	// Generated by github.com/mmcloughlin/addchain v0.4.0.

	// Allocate Temporaries.
	var t0, t1, result E24

	// Step 1: result = x^0x2
	result.CyclotomicSquare(x)

	// Step 2: result = x^0x3
	result.Mul(x, &result)

	// Step 5: t0 = x^0x18
	t0.CyclotomicSquare(&result)
	t0.nSquare(2)

	// Step 8: t1 = x^0xc0
	t1.CyclotomicSquare(&t0)
	t1.nSquare(2)

	// Step 9: t0 = x^0xd8
	t0.Mul(&t0, &t1)

	// Step 10: t0 = x^0xd9
	t0.Mul(x, &t0)

	// Step 19: t0 = x^0x1b200
	t0.nSquareCompressed(9)
	t0.Decompress(&t0)

	// Step 20: result = x^0x1b203
	result.Mul(&result, &t0)

	// Step 35: result = x^0xd9018000
	result.nSquareCompressed(15)
	result.Decompress(&result)

	z.Set(&result)

	return z
}

// MulBy014 multiplication by sparse element (c0, c1, 0, 0, c4, 0)
func (z *E24) MulBy014(c0, c1, c4 *E4) *E24 {

	var a, b E12
	var d E4

	a.Set(&z.D0)
	a.MulBy01(c0, c1)

	b.Set(&z.D1)
	b.MulBy1(c4)
	d.Add(c1, c4)

	z.D1.Add(&z.D1, &z.D0)
	z.D1.MulBy01(c0, &d)
	z.D1.Sub(&z.D1, &a)
	z.D1.Sub(&z.D1, &b)
	z.D0.MulByNonResidue(&b)
	z.D0.Add(&z.D0, &a)

	return z
}