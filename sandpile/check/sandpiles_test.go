//Madison Stulir
//Sandpiles HW

package main

import(
  "testing"
  "fmt"
  "math"
  "runtime"
)


//test InitializeBoard --
//test central , even and odd sizes, pile not divisible by 100 .
//for random do a count and determine the number of coins is correct
func TestInitializeBoard(t *testing.T) {
  fmt.Println("Testing InitializeBoard!")
  type test struct {
    size int
    pile int
    placement string
    answer Board
  }
  tests:=make([]test,4)
  //test 0 - a central board with even columns and pile that is divisible by 100
  tests[0].placement="central"
  tests[0].size=2000
  tests[0].pile=1000
  board := make(Board, 2000)
  for z := range board { board[z] = make([]int, 2000)}
  board[1000][1000]+=1000
  tests[0].answer=board

  //test 1 - a central board with odd columns and pile that is not divisible by 100 - less than 100 try
  tests[1].placement="central"
  tests[1].size=2001
  tests[1].pile=99
  board1 := make(Board, 2001)
  for z1 := range board1 { board1[z1] = make([]int, 2001)}
  board1[1000][1000]+=99
  tests[1].answer=board1
  //test 2 - a random boards with even columns and pile divisible by 100
  tests[2].placement="random"
  tests[2].size=2000
  tests[2].pile=1000
  board2 := make(Board, 2000)
  for z2 := range board2 { board2[z2] = make([]int, 2000)}
  board2[1000][1000]+=1000
  tests[2].answer=board2
  //test 3 - a random board with odd columns and pile not divisible by 100
  tests[3].placement="random"
  tests[3].size=2001
  tests[3].pile=1001
  board3 := make(Board, 2001)
  for z3 := range board3 { board3[z3] = make([]int, 2001)}
  board3[1000][1000]+=1001
  tests[3].answer=board3
  for i := range tests {
    outcome:=InitializeBoard(tests[i].size,tests[i].pile,tests[i].placement)
    //test the total number of coins is correct
    totalCoins:=CountCoins(outcome)
    if totalCoins!=tests[i].pile{
      t.Errorf("Error! For input test dataset %d, your code gives %v and the number of coins is %v", i, totalCoins, tests[i].pile)
    } else {
      fmt.Println("Correct! The number of coins is", totalCoins)
    }
    //check the dimensions of the board are correct
    //len(outcome) and len(outcome[0]) == tests[i].size
    if len(outcome)!=tests[i].size || len(outcome[0])!=tests[i].size {
      t.Errorf("Error! For input test dataset %d, your code gives %d x %d and the size of the board is  is %d x %d", i, len(outcome),len(outcome[0]),tests[i].size,tests[i].size )
    } else {
      fmt.Println("Correct! The width of the board is ", tests[i].size,"x",tests[i].size)
    }
    //if central --  check that the center position has all the coins
    if tests[i].placement=="central" {
      center:=int(math.Floor(float64(tests[i].size)/2))
      if outcome[center][center]!=tests[i].pile{
        t.Errorf("Error! For input test dataset %d, your code gives %v and the number of coins in the center is %v", i, outcome[center][center], tests[i].pile)
      } else {
        fmt.Println("Correct! For the central call, the coins in the center place is", tests[i].pile)
      }
    }
  }
}


//test CountCoins
func TestCountCoins(t *testing.T) {
  fmt.Println("Testing CountCoins!")
  type test struct {
    board Board
    answer int
  }
  tests:=make([]test,4)
  //test case 0: coin on edge of board in width by width
  board := make(Board, 2000)
  for z := range board { board[z] = make([]int, 2000)}
  board[0][0]+=1
  board[1999][1999]+=1
  board[1999][5]+=1
  board[40][1999]+=1
  tests[0].board=board
  tests[0].answer=4

  //test case 1: coins spread throughout board randomly, test more than 1 coin in a position
  board1 := make(Board, 2000)
  for z := range board1 { board1[z] = make([]int, 2000)}
  board1[200][200]+=5
  board1[199][178]+=1
  board1[10][5]+=1
  board1[40][19]+=1
  tests[0].board=board1
  tests[0].answer=8
  for i := range tests {
    outcome:=CountCoins(tests[i].board)
    if outcome != tests[i].answer{
      t.Errorf("Error! For input test dataset %d, your code gives %v and the number of coins is %v", i, outcome, tests[i].answer)
    } else {
      fmt.Println("Correct! The number of coins in the board is", outcome)
    }
  }
}


// test CheckMoreThanFour
func TestCheckMoreThanFour(t *testing.T) {
  fmt.Println("Testing CheckMoreThanFour!")
  type test struct {
    board Board
    answer bool
  }
  tests:=make([]test,3)
  //test 0: make a board that is square  -- board with no values more than 3
  board := make(Board, 2000)
  for z := range board { board[z] = make([]int, 2000)}
  board[0][0]+=1
  board[100][100]+=2
  board[800][100]+=3
  board[70][90]+=3

  tests[0].board=board
  tests[0].answer=false

  //test 1: board with a single  value that is more than 3 in the last position the board checks board[size-1][size-1]
  board1 := make(Board, 2000)
  for z1 := range board1 { board1[z1] = make([]int, 2000)}
  board1[0][0]+=1
  board1[100][100]+=2
  board1[800][100]+=3
  board1[1999][1999]+=4

  tests[1].board=board1
  tests[1].answer=true

  //test 2: a board with no coins at all
  board2 := make(Board, 2000)
  for z2 := range board2 { board2[z2] = make([]int, 2000)}
  tests[2].board=board2
  tests[2].answer=false

  for i := range tests {
    outcome:=CheckMoreThanFour(tests[i].board)
    if outcome != tests[i].answer{
      t.Errorf("Error! For input test dataset %d, your code gives %v and the answer is %v", i, outcome, tests[i].answer)
    } else {
      fmt.Println("Correct! The board has more than 3 coins is", outcome)
    }
  }
}

//test MakeTempBoard -- make sure numRows and numCols are correct
func TestMakeTempBoard(t *testing.T) {
  fmt.Println("Testing MakeTempBoard!")
  type test struct {
    numRows int
    numCols int
    answer Board
  }
  tests:=make([]test,3)
  //make a square board
  tests[0].numRows=5
  tests[0].numCols=5
  board := make(Board, 5)
  for z := range board { board[z] = make([]int, 5)}
  tests[0].answer=board


  //make a 4x5 board
  tests[1].numRows=4
  tests[1].numCols=5
  board1 := make(Board, 4)
  for z1 := range board1 { board1[z1] = make([]int, 5)}
  tests[1].answer=board1

  //make a 5x4 board
  tests[2].numRows=5
  tests[2].numCols=4
  board2 := make(Board, 5)
  for z2 := range board2 { board2[z2] = make([]int, 4)}
  tests[2].answer=board2

  //test that numRows and numCols are correct for the boards
  for i := range tests {
    outcome:=MakeTempBoard(tests[i].numRows,tests[i].numCols)
    rows:=len(outcome)
    cols:=len(outcome[0])
    if rows != len(tests[i].answer) || cols !=len(tests[i].answer[0]){
      t.Errorf("Error! For input test dataset %d, your code gives %d x %d and the board size is %d x %d", i, rows, cols, len(tests[i].answer),len(tests[i].answer[0]))
    } else {
      fmt.Println("Correct! The board size is", rows, "x", cols)
    }
  }
}

//test SandpileSerial -- make sure to keeps all the coins and the all have less than 4
func TestSandpileSerial(t *testing.T) {
  fmt.Println("Testing SandpileSerial!")
  type test struct {
    input Board
    answer Board
  }
  tests:=make([]test,2)
  //test case 0: make a 3x3 board with 10 coins in center
  board := make(Board, 3)
  for z := range board { board[z] = make([]int, 3)}
  board[1][1]+=10
  tests[0].input=board
  board2 := make(Board, 3)
  for z2 := range board2 { board2[z2] = make([]int, 3)}
  board2[0][1]+=2
  board2[1][1]+=2
  board2[1][0]+=2
  board2[1][2]+=2
  board2[2][1]+=2
  tests[0].answer=board2

  board3 := make(Board, 3)
  for z3 := range board3 { board3[z3] = make([]int, 3)}
  board3[1][1]+=4
  board3[0][2]+=2
  board3[2][0]+=4
  tests[1].input=board3
  board4 := make(Board, 3)
  for z4 := range board4 { board4[z4] = make([]int, 3)}
  board4[0][1]+=1
  board4[0][2]+=2
  board4[1][0]+=2
  board4[1][1]+=0
  board4[1][2]+=1
  board4[2][1]+=2
  tests[1].answer=board4

  for i := range tests {
    outcome:=SandpileSerial(tests[i].input)
    answer:=tests[i].answer
    numRows:=len(tests[0].input)
    numCols:=len(tests[0].input[0])
    equal:=true
    for i:=0;i<numRows;i++{
      for j:=0;j<numCols;j++{
        if outcome[i][j]!=answer[i][j]{
          equal=false
        }
      }
    }
    if equal==false {
      t.Errorf("Error!For input test dataset %d, the outcome is not the same as the answer!", i)
    } else {
      fmt.Println("Correct! The outcome is as expected.")
    }
  }
}

//test SandpileMultiProcs --make sure it does not lose coins
func TestSandpileMultiProcs(t *testing.T) {
  fmt.Println("Testing SandpileMultiProcs!")
  type test struct {
    input Board
    answer Board
  }
  tests:=make([]test,2)
  //test case 0: make a 3x3 board with 10 coins in center
  board := make(Board, 200)
  for z := range board { board[z] = make([]int, 200)}
  board[1][1]+=10
  tests[0].input=board
  board2 := make(Board, 200)
  for z2 := range board2 { board2[z2] = make([]int, 200)}
  board2[0][1]+=2
  board2[1][1]+=2
  board2[1][0]+=2
  board2[1][2]+=2
  board2[2][1]+=2
  tests[0].answer=board2

  board3 := make(Board, 200)
  for z3 := range board3 { board3[z3] = make([]int, 200)}
  board3[1][1]+=4
  board3[0][2]+=2
  board3[2][0]+=4
  tests[1].input=board3
  board4 := make(Board, 200)
  for z4 := range board4 { board4[z4] = make([]int, 200)}
  board4[0][1]+=1
  board4[0][2]+=2
  board4[1][0]+=2
  board4[1][1]+=0
  board4[1][2]+=1
  board4[2][1]+=2
  board4[3][0]+=1
  tests[1].answer=board4

  for i := range tests {
    numProcs:=runtime.NumCPU()
    outcome:=SandpileMultiProcs(tests[i].input,numProcs)
    answer:=tests[i].answer
    numRows:=len(tests[0].input)
    numCols:=len(tests[0].input[0])
    equal:=true
    for i:=0;i<numRows;i++{
      for j:=0;j<numCols;j++{
        if outcome[i][j]!=answer[i][j]{
          equal=false
        }
      }
    }
    if equal==false {
      t.Errorf("Error! For input test dataset %d, the outcome is not the same as the answer!", i)
    } else {
      fmt.Println("Correct! The outcome is as expected.")
    }
  }
}
