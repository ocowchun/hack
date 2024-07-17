// Draw a filled rectangle at the upper left corner of the screen,
// 16 pixels wide and RAM[0] pixels long

// Pseudo code
// long = RAM[0]
// for (i = 0; i < long; ++i) {
//   draw 16 black pixels at the begining of row i
//

// addr = SCREEN
// n = RAM[0]
// i = 0

// LOOP:
// if i > n, go to END
// RAM[addr] = -1 // 1111111111111111
// addr = addr + 32
// i = i + 1
// go to LOOP

// END:
// go to END;

// set addr = SCREEN
@SCREEN
D=A
@addr
M=D

// set n = RAM[0]
@0
D=M
@n
M=D

// set i = 0
@i
M=0


(LOOP)
// if i == n, go to end
@i
D=M
@n
D=D-M
@END
D;JEQ // i - n == 0


// RAM[addr] = -1 // 1111111111111111
@addr
A=M
M=-1

// addr = addr + 32
@32
D=A
@addr
M=D+M

// i = i + 1
@i
M=M+1
// go to LOOP
@LOOP
0;JMP


(END)
@END
0;JMP



