// if any key is pressed, turn the whole screen into black
// otherwise, turn the whole screen into white

// white = 0
// black = -1 
// screen_end = SCREEN + 8192
// for(true) {
//   if (KBD==0) {
//     color=white
//   }
//   else { 
//     color=black
//   }
//   for (i = SCREEN; i < screen_end; ++i) {
//      RAM[i] = color
//   }

// white = 0
@white
M=0

// black = -1
@black
M=-1

// color=white
@white
D=M
@color
M=D

// screen_end = SCREEN + 8192
@8192
D=A
@SCREEN
D=D+A
@screen_end
M=D

(CHECK_KEY)
@KBD
D=M
@DISPLAY_WHITE
D;JEQ // if KBD == 0
@DISPLAY_BLACK
0;JMP

(DISPLAY_WHITE)
@white
D=M
@color
M=D // set color = white
@SCREEN
D=A
@i
M=D
@UPDATE_SCREEN
0;JMP

(DISPLAY_BLACK)
@black
D=M
@color
M=D // set color = white
@SCREEN
D=A
@i
M=D
@UPDATE_SCREEN
0;JMP

(UPDATE_SCREEN)
@i
D=M
@screen_end
D=D-M
@CHECK_KEY
D;JEQ // i - screen_end == 0 -> i == screen_end 


// RAM[i] = color
@color
D=M
@i
A=M
M=D

@i
M=M+1
@UPDATE_SCREEN
0;JMP





