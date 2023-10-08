//Madison Stulir
//Sandpiles HW

package main

import(
  //"fmt"
  "math/rand"
  "math"
  "runtime"
  "log"
  "time"
  "os"
  "strconv"
)

type Board [][]int

func main() {
  //size is the width of the checkerboard in each direction - a positive int
  //it will be os.args[1]
  size,err1:=strconv.Atoi(os.Args[1])
  if err1!=nil{
		panic(err1)
	}
  //pile is the number of coins to distribute on the board
  //it will be os.args[2]
  pile,err2:=strconv.Atoi(os.Args[2])
  if err2!=nil{
		panic(err2)
	}

  //placement is the distribution of the pile - central or random
  placement:=os.Args[3]
  if placement!="central" && placement!="random" {
    panic("Placement input was not central or random.")
  }

  TimingSandpiles(size,pile,placement)
}

//TimingSandpiles completes a sandpile toppling in serial and parallel and compares the time they take to do so
//Input: size, an integer representing the width of the board, pile, an int representing the number of coins to put on the board, and placement, a string of either central or random
//Output: none. 2 png images representing the result from each method are generated. The time each method took is printed to the console
func TimingSandpiles(size, pile int, placement string) {
  board:=InitializeBoard(size,pile,placement)
  numProcs:=runtime.NumCPU()

  start:=time.Now()
  parallelBoard:=SandpileMultiProcs(board, numProcs)
  elapsed:=time.Since(start)
  log.Printf("Creating sandpiles in parallel took %s",elapsed)

  start2:=time.Now()
  serialBoard:=SandpileSerial(board)
  elapsed2:=time.Since(start2)
  log.Printf("Creating in serial took %s",elapsed2)

  c:=DrawBoard(serialBoard,1)

  c.SaveToPNG("serial.png")

  d:=DrawBoard(parallelBoard,1)

  d.SaveToPNG("parallel.png")
}


//initializeBoard takes in a size and returns a Board of that size, filled with pile number of coins distributed either centrally or randomly according to placement
//Input: size of the desired board and a number of coins, pile, as integers and a string representing the method by which to distribute the pile
//Output: a board of the size width and height, with pile num coins distributed according to placement
func InitializeBoard(size,pile int, placement string) Board {
  var board Board
  board = make(Board, size)
  for z := range board { board[z] = make([]int, size)}
  center:=int(math.Floor(float64(size)/2))
  if placement=="central" {
     board[center][center]+=pile
  } else if placement=="random" {
    //need to account for pile that is not divisible by 100
    //- use modulo and the remainder is the number of coins that need to be added as an extra to some piles (add 1 to each of the first remainder number of piles so it is as evenly distributed as possible)
    remainder:=pile%100
    //loop through 100 times, selecting a random spot within the board and placing pile/100 coins in that location
    for i:=0;i<100;i++{
      row:=rand.Intn(size)
      col:=rand.Intn(size)
      board[row][col]+=pile/100
      //add the remainder to the first remainder number of spots on the board by 1s to distribute evenly
      if i<remainder{
        board[row][col]+=1
      }
    }
  }
  return board
}

//SandpileMultiProcs splits a board to distribute coins in parallel
//Input: an initialBoard of type board, and numProcs, an int representing the number of processors to split the board into
//Output: an updated board that has been fully toppled such that no position has more than 3 coins
func SandpileMultiProcs(board Board, numProcs int) Board {
  size:=len(board)
  for { //this will be a while loop where we need to check that the number in each position of the board is less than 4.
    if CheckMoreThanFour(board) { //loop will continue while the check less than 4 returns true, once it is false the loop will stop
      //make the channel for the board slices to return to
      c:=make(chan map[int]Board, numProcs)
      //make the channel for the coins to return to (slice of ints where the first int is the starting row of the board that it came out of and the second whether it went up or down out of the board)
      coins:=make(chan []int, numProcs)
      startIndexes:=make([]int,0)
      //divide the board into numProcs pieces and call go routines for each to the Sandpile single proc function
      for i := 0; i < numProcs; i++ {
    		//split the array of genomes into numProcs pieces
    		startIndex := i * (size / numProcs)
        startIndexes=append(startIndexes,startIndex)
    		endIndex := (i + 1) * (size / numProcs)
    		if i < numProcs-1 {
    			//call the go routine for the number of genomes within this section
    			go SandpileSingleProc(board[startIndex:endIndex], startIndex, c, coins)
    		} else {
    			//end of slice
    			go SandpileSingleProc(board[startIndex:], startIndex, c, coins)
    		}
      }
      //pull the boards from the channel (they are maps with their corresponding starting index)
      m:=make(map[int]Board)
      for i:=0;i<numProcs;i++{
        mapOut:= <- c
        for key,value :=range mapOut{
          m[key]=value
        }
      }
      //stitch the boards together (they need to be sorted)
      //make an empty board of length 0 to use
      stitch:=make(Board,0)
      //loop through the indices in startIndexes and append their value to create the board stitch
      for i:=0;i<numProcs;i++{
        index:=startIndexes[i]
        stitch=append(stitch,m[index]...)
      }
      //add the coins into the board from the other channel
      for i:=0;i<numProcs;i++{
        coinLocSlice:=<- coins
        n:=len(coinLocSlice)
        //list of all coins that fell off edge from that proc. even is the row, odd is the col
        //jump through by 2
        for j:=0;j<n;j+=2{
          if coinLocSlice[j]>=0 && coinLocSlice[j]<size{
            stitch[coinLocSlice[j]][coinLocSlice[j+1]]+=1
          }
        }
      }
      board=stitch
    } else {
      break
    }
  }
return board
}

//SandpileSingleProc takes a Board (subsection of the original board) and completes the topple operation for 1 generation. It pushes any coins that fall off the edge into a channel and pushes the completed board into another channel
//Input: a board of type Board, the startIndex of the board in the bigger board, c a channel which takes in the result of the toppled board, and coins, a channel that collects coins that fall off the edge in an array
//Output: There is no output. The updated board is pushed into a channel. The coins that fall off the edge are pushed into another channel (as a single array for all coins that came off in this processor)
func SandpileSingleProc(board Board, startIndex int, c chan map[int]Board, coins chan []int) {
  //get the numRows and numCols of the input board (it will not be square)
  numRows:=len(board)
  //loop through and distribute coins
  numCols:=len(board[0])
  //make a currentBoard of numRows by numCols
  currentBoard:=MakeTempBoard(numRows, numCols)
  //make a slice of ints to add to for each coin that falls off. push into coins channel at the end -- all odd indexes will be the starting point of a coin
  coinsFall:=make([]int,0)
  for i:=0;i<numRows;i++ {
    for j:=0;j<numCols;j++ {
        if i==0 && j==numCols-1 { //top left corner 0,size
          if board[i][j]>3{
            currentBoard[i+1][j]+=1
            currentBoard[i][j-1]+=1
            currentBoard[i][j]+=(board[i][j]-4)
            coinsFall=append(coinsFall,startIndex-1)
            coinsFall=append(coinsFall,j)
            } else {
              currentBoard[i][j]+=board[i][j]
            }
          } else if i==numRows-1 && j==numCols-1 { //top right corner size,size
            if board[i][j]>3{
              currentBoard[i-1][j]+=1
              currentBoard[i][j-1]+=1
              currentBoard[i][j]+=(board[i][j]-4)
              coinsFall=append(coinsFall,startIndex+i+1)
              coinsFall=append(coinsFall,j)
            } else {
              currentBoard[i][j]+=board[i][j]
            }
          } else if i==0 && j==0 { //bottom left corner 0,0
            if board[i][j]>3{
              currentBoard[i+1][j]+=1
              currentBoard[i][j+1]+=1
              currentBoard[i][j]+=(board[i][j]-4)
              coinsFall=append(coinsFall,startIndex-1)
              coinsFall=append(coinsFall,j)
            } else {
              currentBoard[i][j]+=board[i][j]
            }
          } else if i==numRows-1 && j==0 { //bottom right corner size,0
            if board[i][j]>3{
              currentBoard[i-1][j]+=1
              currentBoard[i][j+1]+=1
              currentBoard[i][j]+=(board[i][j]-4)
              coinsFall=append(coinsFall,startIndex+i+1)
              coinsFall=append(coinsFall,j)
            } else {
              currentBoard[i][j]+=board[i][j]
            }
          } else if i==numRows-1 { //top side i=size
            if board[i][j]>3{
              currentBoard[i-1][j]+=1
              currentBoard[i][j-1]+=1
              currentBoard[i][j+1]+=1
              currentBoard[i][j]+=(board[i][j]-4)
              coinsFall=append(coinsFall,startIndex+i+1)
              coinsFall=append(coinsFall,j)
            } else {
              currentBoard[i][j]+=board[i][j]
            }
          } else if i==0 { //bottom side i=0
            if board[i][j]>3{
              currentBoard[i+1][j]+=1
              currentBoard[i][j-1]+=1
              currentBoard[i][j+1]+=1
              currentBoard[i][j]+=(board[i][j]-4)
              coinsFall=append(coinsFall,startIndex-1)
              coinsFall=append(coinsFall,j)
            } else {
              currentBoard[i][j]+=board[i][j]
            }
          } else if j==numCols-1 { //right side j=size
            if board[i][j]>3{
              currentBoard[i-1][j]+=1
              currentBoard[i+1][j]+=1
              currentBoard[i][j-1]+=1
              currentBoard[i][j]+=(board[i][j]-4)
            } else {
              currentBoard[i][j]+=board[i][j]
            }
          } else if j==0 { //left side j=0
            if board[i][j]>3{
              currentBoard[i-1][j]+=1
              currentBoard[i+1][j]+=1
              currentBoard[i][j+1]+=1
              currentBoard[i][j]+=(board[i][j]-4)
            } else {
              currentBoard[i][j]+=board[i][j]
            }
          } else { //not on a board edge
            if board[i][j]>3{
              currentBoard[i-1][j]+=1
              currentBoard[i+1][j]+=1
              currentBoard[i][j-1]+=1
              currentBoard[i][j+1]+=1
              currentBoard[i][j]+=(board[i][j]-4)
            } else {
              currentBoard[i][j]+=board[i][j]
            }
          }
      }
    }
  //push the map of starting index to board into the c channel
  m := make(map[int]Board)
  m[startIndex]=currentBoard
  c <- m
  //push the coin array into the coin channel
  coins <- coinsFall
}

//MakeTempBoard makes a Board of numRows x numCols
//Input: numRows and numCols as integers
//Output: a Board of the size indicated
func MakeTempBoard(numRows, numCols int) Board {
  var board Board
  board = make(Board, numRows)
  for z := range board { board[z] = make([]int, numCols)}
  return board
}

//serial function to be called for updating sandpiles
//Input: a board of type Board
//output: a board with coins fully distributed
func SandpileSerial(board Board) Board {
  size:=len(board)
  for { //this will be a while loop where we need to check that the number in each position of the board is less than 4.
    if CheckMoreThanFour(board) { //loop will continue while the check less than 4 returns true, once it is false the loop will stop
      //make a board to work with in this generation
      currentBoard:=InitializeBoard(len(board),0,"central")
      //assign the values of currentBoard according to the values in board
      //loop through the boards values and if greater than or equal to 4, add 1 to each of its neghbors in currentBoard and subtract 4 from its value and assign that to itself in currentBoard
      for i:=0;i<size;i++ {
        for j:=0;j<size;j++ {
            if i==0 && j==size-1 { //top left corner 0,size
              if board[i][j]>3{
                currentBoard[i+1][j]+=1
                currentBoard[i][j-1]+=1
                currentBoard[i][j]+=(board[i][j]-4)
              } else {
                currentBoard[i][j]+=board[i][j]
              }
              } else if i==size-1 && j==size-1 { //top right corner size,size
                if board[i][j]>3{
                  currentBoard[i-1][j]+=1
                  currentBoard[i][j-1]+=1
                  currentBoard[i][j]+=(board[i][j]-4)
                } else {
                  currentBoard[i][j]+=board[i][j]
                }
              } else if i==0 && j==0 { //bottom left corner 0,0
                if board[i][j]>3{
                  currentBoard[i+1][j]+=1
                  currentBoard[i][j+1]+=1
                  currentBoard[i][j]+=(board[i][j]-4)
                } else {
                  currentBoard[i][j]+=board[i][j]
                }
              } else if i==size-1 && j==0 { //bottom right corner size,0
                if board[i][j]>3{
                  currentBoard[i-1][j]+=1
                  currentBoard[i][j+1]+=1
                  currentBoard[i][j]+=(board[i][j]-4)
                } else {
                  currentBoard[i][j]+=board[i][j]
                }
              } else if i==size-1 { //top side i=size
                if board[i][j]>3{
                  currentBoard[i-1][j]+=1
                  currentBoard[i][j-1]+=1
                  currentBoard[i][j+1]+=1
                  currentBoard[i][j]+=(board[i][j]-4)
                } else {
                  currentBoard[i][j]+=board[i][j]
                }
              } else if i==0 { //bottom side i=0
                if board[i][j]>3{
                  currentBoard[i+1][j]+=1
                  currentBoard[i][j-1]+=1
                  currentBoard[i][j+1]+=1
                  currentBoard[i][j]+=(board[i][j]-4)
                } else {
                  currentBoard[i][j]+=board[i][j]
                }
              } else if j==size-1 { //right side j=size
                if board[i][j]>3{
                  currentBoard[i-1][j]+=1
                  currentBoard[i+1][j]+=1
                  currentBoard[i][j-1]+=1
                  currentBoard[i][j]+=(board[i][j]-4)
                } else {
                  currentBoard[i][j]+=board[i][j]
                }
              } else if j==0 { //left side j=0
                if board[i][j]>3{
                  currentBoard[i-1][j]+=1
                  currentBoard[i+1][j]+=1
                  currentBoard[i][j+1]+=1
                  currentBoard[i][j]+=(board[i][j]-4)
                } else {
                  currentBoard[i][j]+=board[i][j]
                }
              } else { //not an edge
                if board[i][j]>3{
                  currentBoard[i-1][j]+=1
                  currentBoard[i+1][j]+=1
                  currentBoard[i][j-1]+=1
                  currentBoard[i][j+1]+=1
                  currentBoard[i][j]+=(board[i][j]-4)
                } else {
                  currentBoard[i][j]+=board[i][j]
                }
              }
        }
      }
      //assign our currentBoard back to board
      board=currentBoard
    } else {
      break
    }
  }
  return board
}

//CheckMoreThanFour is a function that takes in a Board and returns false if all squares have a value less than 4 and true if not
//Input: a Board named board
//Output: a boolean as to whether the board is evenly distributed
func CheckMoreThanFour(board Board) bool {
  //loop through all positions of the board and check their values
  size:=len(board) //we know it will have equal numbers of rows and columns
  for i:=0;i<size;i++ {
    for j:=0;j<size;j++ {
      // once we reach a position that is 4 or greater, we return true
      if board[i][j]>3{
        return true
      }
    }
  }
  // if we get through the whole board, we return false
  return false
}

//CountCoins determines the total number of coins in the board
//Input: a board of type Board
//Output: an integer representing how many coins are on the board
func CountCoins(board Board) int {
  size:=len(board)
  count:=0
  for i:=0;i<size;i++ {
    for j:=0;j<size;j++ {
      count+=board[i][j]
    }
  }
  return count
}
package main

import (
  "canvas"
  //"image"
)

//DrawBoard takes a Board objects as input along with a cellWidth and n parameter.
//It returns an image corresponding to drawing every nth board to a file,
//where each cell is cellWidth x cellWidth pixels.
func DrawBoard(b Board, cellWidth int) canvas.Canvas {
	// need to know how many pixels wide and tall to make our image
	height := len(b) * cellWidth
	width := len(b[0]) * cellWidth

	// think of a canvas as a PowerPoint slide that we draw on
	c := canvas.CreateNewCanvas(width, height)

	// canvas will start as black, so we should fill in colored squares
	//fmt.Println("board",b)
	for i := range b {
		for j := range b[i] {
			val := b[i][j]
      var red uint8
      var green uint8
      var blue uint8
      if val == 0{
        red=0
        green=0
        blue=0
      } else if val==1{
        red=85
        green=85
        blue=85
      } else if val==2{
        red=170
        green=170
        blue=170
      } else if val==3 {
        red=255
        green=255
        blue=255
      }
			// draw a rectangle in right place with this color
			c.SetFillColor(canvas.MakeColor(red,green,blue))

			x := i * cellWidth
			y := j * cellWidth
			c.ClearRect(x, y, x+cellWidth, y+cellWidth)
			c.Fill()
		}
	}

	// canvas has an image field that we should return
	return c
}
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
