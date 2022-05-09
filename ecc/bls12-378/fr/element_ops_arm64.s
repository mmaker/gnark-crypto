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

#include "textflag.h"
#include "funcdata.h"

// modulus q
DATA q<>+0(SB)/8, $3643768340310130689
DATA q<>+8(SB)/8, $16926637627159085057
DATA q<>+16(SB)/8, $9761692607219216639
DATA q<>+24(SB)/8, $2371068001496280753
GLOBL q<>(SB), (RODATA+NOPTR), $32
// qInv0 q'[0]
DATA qInv0<>(SB)/8, $3643768340310130687
GLOBL qInv0<>(SB), (RODATA+NOPTR), $8
#define storeVector(ePtr, e0, e1, e2, e3) \
	STP (e0, e1), 0(ePtr)  \
	STP (e2, e3), 16(ePtr) \

// add(res, x, y *Element)
TEXT ·add(SB), NOSPLIT, $0-24
	LDP x+8(FP), (R4, R5)

	// load operands and add mod 2^r
	LDP  0(R4), (R0, R6)
	LDP  0(R5), (R1, R7)
	ADDS R0, R1, R0
	ADCS R6, R7, R1
	LDP  16(R4), (R2, R6)
	LDP  16(R5), (R3, R7)
	ADCS R2, R3, R2
	ADCS R6, R7, R3

	// load modulus and subtract
	LDP  q<>+0(SB), (R4, R5)
	SUBS R4, R0, R4
	SBCS R5, R1, R5
	LDP  q<>+16(SB), (R6, R7)
	SBCS R6, R2, R6
	SBCS R7, R3, R7

	// reduce if necessary
	CSEL CS, R4, R0, R0
	CSEL CS, R5, R1, R1
	CSEL CS, R6, R2, R2
	CSEL CS, R7, R3, R3

	// store
	MOVD res+0(FP), R4
	storeVector(R4, R0, R1, R2, R3)
	RET

// sub(res, x, y *Element)
TEXT ·sub(SB), NOSPLIT, $0-24
	LDP x+8(FP), (R4, R5)

	// load operands and subtract mod 2^r
	LDP  0(R4), (R0, R6)
	LDP  0(R5), (R1, R7)
	SUBS R1, R0, R0
	SBCS R7, R6, R1
	LDP  16(R4), (R2, R6)
	LDP  16(R5), (R3, R7)
	SBCS R3, R2, R2
	SBCS R7, R6, R3

	// load modulus and select
	MOVD $0, R8
	LDP  q<>+0(SB), (R4, R5)
	CSEL CS, R8, R4, R4
	CSEL CS, R8, R5, R5
	LDP  q<>+16(SB), (R6, R7)
	CSEL CS, R8, R6, R6
	CSEL CS, R8, R7, R7

	// augment (or not)
	ADDS R0, R4, R0
	ADCS R1, R5, R1
	ADCS R2, R6, R2
	ADCS R3, R7, R3

	// store
	MOVD res+0(FP), R4
	storeVector(R4, R0, R1, R2, R3)
	RET

// double(res, x *Element)
TEXT ·double(SB), NOSPLIT, $0-16
	LDP res+0(FP), (R5, R4)

	// load operands and add mod 2^r
	LDP  0(R4), (R0, R1)
	ADDS R0, R0, R0
	ADCS R1, R1, R1
	LDP  16(R4), (R2, R3)
	ADCS R2, R2, R2
	ADCS R3, R3, R3

	// load modulus and subtract
	LDP  q<>+0(SB), (R4, R6)
	SUBS R4, R0, R4
	SBCS R6, R1, R6
	LDP  q<>+16(SB), (R7, R8)
	SBCS R7, R2, R7
	SBCS R8, R3, R8

	// reduce if necessary
	CSEL CS, R4, R0, R0
	CSEL CS, R6, R1, R1
	CSEL CS, R7, R2, R2
	CSEL CS, R8, R3, R3

	// store
	storeVector(R5, R0, R1, R2, R3)
	RET

// neg(res, x *Element)
TEXT ·neg(SB), NOSPLIT, $0-16
	LDP res+0(FP), (R5, R4)

	// load operands and subtract
	MOVD $0, R8
	LDP  0(R4), (R0, R1)
	LDP  q<>+0(SB), (R6, R7)
	ORR  R0, R8, R8              // has x been 0 so far?
	ORR  R1, R8, R8
	SUBS R0, R6, R0
	SBCS R1, R7, R1
	LDP  16(R4), (R2, R3)
	LDP  q<>+16(SB), (R6, R7)
	ORR  R2, R8, R8              // has x been 0 so far?
	ORR  R3, R8, R8
	SBCS R2, R6, R2
	SBCS R3, R7, R3
	TST  $0xffffffffffffffff, R8
	CSEL EQ, R8, R0, R0
	CSEL EQ, R8, R1, R1
	CSEL EQ, R8, R2, R2
	CSEL EQ, R8, R3, R3

	// store
	storeVector(R5, R0, R1, R2, R3)
	RET