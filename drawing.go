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
