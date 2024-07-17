// R2 = R0 * R1

// a = R0
// b = R1
// i = 0
// prod = 0
// for (i = 0; i < R1; ++i) {
//   prod = prod + R0
// }

//i = 0
@i
M=0

// prod = 0
@prod
M=0

(LOOP)
@i
D=M
@R1
D=D-M // i - R1
@SET_R2
D;JEQ // i - R1 = 0 -> i = R1

@R0
D=M
@prod
M=D+M // prod = prod + R0

// i++
@i
M=M+1
@LOOP
0;JMP

(SET_R2)
@prod
D=M
@R2
M=D
@END
0;JMP


(END)
@END
0;JMP

