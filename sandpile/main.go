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
