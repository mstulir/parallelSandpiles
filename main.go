package main

import(
  //"fmt"
  "math/rand"
  "math"
  "runtime"
  "log"
  "time"
  //"os"
)

type Board [][]int

func main() {
  //size is the width of the checkerboard in each direction - a positive int
  //it will be os.args[1]
  size:=1000 //need to panic if this is an odd input

  //pile is the number of coins to distribute on the board
  //it will be os.args[2]
  pile:=50000

  //placement is the distribution of the pile - central or random
  placement:="random"

  TimingSandpiles(size,pile,placement)

}

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


func SandpileMultiProcs(board Board, numProcs int) Board {
  size:=len(board)

  for { //this will be a while loop where we need to check that the number in each position of the board is less than 4.
    if CheckMoreThanFour(board) { //loop will continue while the check less than 4 returns true, once it is false the loop will stop
      //make the channel for the board slices to return to
      c:=make(chan map[int]Board, numProcs)
      //make the channel for the coins to return to (slice of ints where the first int is the starting row of the board that it came out of and the second whether it went up or down out of the board)
      coins:=make(chan []int, size*2*numProcs)
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
      //make a slice of boards of the length numProcs
      //input their keys into a list -- can we use those indices to properly sort out the boards?
      //make a map for all of the individual maps to be compiled to
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
      //need to check for coins with negative values in position 1 and do not include them. otherwise increment the stitch board at the positions given as row,col

      for i:=0;i<size*2*numProcs;i++{
        coinLocSlice:=<- coins
        //check the 1st value in the string --  is it negative -- if so, we can ignore
        if coinLocSlice[0]>=0 && coinLocSlice[0]<size{
          //otherwise, we can add 1 to the value of stitch at the position indicated by the 1st and 2nd numbers
          stitch[coinLocSlice[0]][coinLocSlice[1]]+=1
        }
      }

      board=stitch
    } else {
      break
    }
  }
return board

}


func SandpileSingleProc(board Board, startIndex int, c chan map[int]Board, coins chan []int) {
  //get the numRows and numCols of the input board (it will not be square)
  numRows:=len(board)
  //loop through and distribute coins
  numCols:=len(board[0])
  //make a currentBoard of numRows by numCols
  currentBoard:=MakeTempBoard(numRows, numCols)


  //have  tO ADD IN COLLECTING COINS FALLING OFF THE TOP OR BOTTOM

  //collect coins that fall off edge into the coins channel
  for i:=0;i<numRows;i++ {
    for j:=0;j<numCols;j++ {
      //if board[i][j]>3{
        //need to account for cases in which the coin will go off the edge!!! (in each direction)
        if i==0 && j==numCols-1 { //top left corner 0,size
          if board[i][j]>3{
            currentBoard[i+1][j]+=1
            currentBoard[i][j-1]+=1
            currentBoard[i][j]+=(board[i][j]-4)
            coins <- []int{startIndex-1,j}
            } else {
              currentBoard[i][j]+=board[i][j]
              coins <- []int{-1,-1}
            }
          } else if i==numRows-1 && j==numCols-1 { //top right corner size,size
            if board[i][j]>3{
              currentBoard[i-1][j]+=1
              currentBoard[i][j-1]+=1
              currentBoard[i][j]+=(board[i][j]-4)
              coins <- []int{startIndex+i+1,j}
            } else {
              currentBoard[i][j]+=board[i][j]
              coins <- []int{-1,-1}
            }
          } else if i==0 && j==0 { //bottom left corner 0,0
            if board[i][j]>3{
              currentBoard[i+1][j]+=1
              currentBoard[i][j+1]+=1
              currentBoard[i][j]+=(board[i][j]-4)
              coins <- []int{startIndex-1,j}
            } else {
              currentBoard[i][j]+=board[i][j]
              coins <- []int{-1,-1}
            }
          } else if i==numRows-1 && j==0 { //bottom right corner size,0
            if board[i][j]>3{
              currentBoard[i-1][j]+=1
              currentBoard[i][j+1]+=1
              currentBoard[i][j]+=(board[i][j]-4)
              coins <- []int{startIndex+i+1,j}
            } else {
              currentBoard[i][j]+=board[i][j]
              coins <- []int{-1,-1}
            }
          } else if i==numRows-1 { //top side i=size
            if board[i][j]>3{
              currentBoard[i-1][j]+=1
              currentBoard[i][j-1]+=1
              currentBoard[i][j+1]+=1
              currentBoard[i][j]+=(board[i][j]-4)
              coins <- []int{i+startIndex+1,j}
            } else {
              currentBoard[i][j]+=board[i][j]
              coins <- []int{-1,-1}
            }
          } else if i==0 { //bottom side i=0
            if board[i][j]>3{
              currentBoard[i+1][j]+=1
              currentBoard[i][j-1]+=1
              currentBoard[i][j+1]+=1
              currentBoard[i][j]+=(board[i][j]-4)
              coins <- []int{startIndex-1,j}
            } else {
              currentBoard[i][j]+=board[i][j]
              coins <- []int{-1,-1}
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
          } else {
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

}


func MakeTempBoard(numRows, numCols int) Board {
  var board Board
  board = make(Board, numRows)
  for z := range board { board[z] = make([]int, numCols)}
  return board

}

//serial function to be called for updating sandpiles
//Input: a board
//output: a board with coins fully distributed
func SandpileSerial(board Board) Board {
  size:=len(board)

  for { //this will be a while loop where we need to check that the number in each position of the board is less than 4.
    if CheckMoreThanFour(board) { //loop will continue while the check less than 4 returns true, once it is false the loop will stop
      //make a board to work with in this generation
      currentBoard:=InitializeBoard(len(board),0,"central")
      //assign the values of currentBoard according to the values in board
      //loop through the boards values and if greater than 4, add 1 to each of its neghbors in currentBoard and subtract 4 from its value and assign that to itself in currentBoard
      for i:=0;i<size;i++ {
        for j:=0;j<size;j++ {
          //if board[i][j]>3{
            //need to account for cases in which the coin will go off the edge!!! (in each direction)
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
              } else {
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
          //} else { //assign the currentBoard to the value currently in board

          //  currentBoard[i][j]+=board[i][j]
        //  }
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
