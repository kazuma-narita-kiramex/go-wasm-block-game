package main

import (
	"fmt"
	"math"
	"syscall/js"
)

type Canvas struct {
	ctx        js.Value
	width      int
	height     int
	offsetLeft int
}

type Ball struct {
	x      int
	y      int
	dx     int
	dy     int
	radius int
}

type Paddle struct {
	x      int
	y      int
	dx     int
	width  int
	height int
}

type Brick struct {
	x      int
	y      int
	status int
}

type Bricks struct {
	bricks      [][]Brick
	rowCount    int
	columnCount int
	width       int
	height      int
	padding     int
	offsetTop   int
	offsetLeft  int
}

type Score struct {
	score int
	x     int
	y     int
}

type Live struct {
	live int
	x    int
	y    int
}

type GameManager struct {
	canvas       Canvas
	ball         Ball
	paddle       Paddle
	bricks       Bricks
	score        Score
	live         Live
	rightPressed bool
	leftPressed  bool
	lives        int
}

func main() {
	c := make(chan struct{}, 0)
	game()
	<-c
}

func game() {
	// get canvas
	canvasElem := js.Global().Get("document").Call("getElementById", "canvas")

	// init canvas
	canvas := Canvas{
		ctx:        canvasElem.Call("getContext", "2d"),
		width:      canvasElem.Get("width").Int(),
		height:     canvasElem.Get("height").Int(),
		offsetLeft: canvasElem.Get("offsetLeft").Int(),
	}

	// init gameManager obj
	gameManager := GameManager{
		canvas: canvas,
		ball: Ball{
			x:      canvas.width / 2,
			y:      canvas.height - 30,
			dx:     2,
			dy:     -2,
			radius: 10,
		},
		paddle: Paddle{
			width:  75,
			height: 10,
			dx:     5,
		},
		bricks: Bricks{
			rowCount:    3,
			columnCount: 5,
			width:       75,
			height:      20,
			padding:     10,
			offsetTop:   30,
			offsetLeft:  30,
		},
		score: Score{
			score: 0,
			x:     8,
			y:     20,
		},
		live: Live{
			live: 1,
			x:    canvas.width - 65,
			y:    20,
		},
		rightPressed: false,
		leftPressed:  false,
	}

	// init paddle
	gameManager.paddle.x = (canvas.width - gameManager.paddle.width) / 2
	gameManager.paddle.y = (canvas.height - gameManager.paddle.height)

	// init bricks
	bricks := make([][]Brick, gameManager.bricks.columnCount)
	for i := 0; i < gameManager.bricks.columnCount; i++ {
		bricks[i] = make([]Brick, gameManager.bricks.rowCount)
		for j := 0; j < gameManager.bricks.rowCount; j++ {
			bricks[i][j] = Brick{
				x:      0,
				y:      0,
				status: 1,
			}
		}
	}
	gameManager.bricks.bricks = bricks

	// add event listener
	js.Global().Get("document").Call("addEventListener", "keydown", js.FuncOf(gameManager.keyDownHandler), false)
	js.Global().Get("document").Call("addEventListener", "keyup", js.FuncOf(gameManager.keyUpHandler), false)
	js.Global().Get("document").Call("addEventListener", "mousemove", js.FuncOf(gameManager.mouseMoveHandler), false)

	// drawing on canvas
	// for {
	// 	gameManager.draw()
	// 	time.Sleep(time.Millisecond * 10)
	// }
	gameManager.draw()
}

func (g *GameManager) keyDownHandler(this js.Value, args []js.Value) interface{} {
	e := args[0]
	key := e.Get("key").String()
	if key == "Right" || key == "ArrowRight" {
		g.rightPressed = true
	} else if key == "Left" || key == "ArrowLeft" {
		g.leftPressed = true
	}
	return nil
}

func (g *GameManager) keyUpHandler(this js.Value, args []js.Value) interface{} {
	e := args[0]
	key := e.Get("key").String()
	if key == "Right" || key == "ArrowRight" {
		g.rightPressed = false
	} else if key == "Left" || key == "ArrowLeft" {
		g.leftPressed = false
	}
	return nil
}

func (g *GameManager) mouseMoveHandler(this js.Value, args []js.Value) interface{} {
	e := args[0]
	clientX := e.Get("clientX").Int()
	relativeX := clientX - g.canvas.offsetLeft
	if relativeX > 0 && relativeX < g.canvas.width {
		g.paddle.x = relativeX - g.paddle.width/2
	}
	return nil
}

func (g *GameManager) collisionDetection() {
	for i := 0; i < g.bricks.columnCount; i++ {
		for j := 0; j < g.bricks.rowCount; j++ {
			b := g.bricks.bricks[i][j]
			if b.status == 1 {
				if g.ball.x > b.x && g.ball.x < b.x+g.bricks.width && g.ball.y > b.y && g.ball.y < b.y+g.bricks.height {
					g.ball.dy = -g.ball.dy
					g.bricks.bricks[i][j].status = 0
					g.score.score++
					if g.score.score == g.bricks.columnCount*g.bricks.rowCount {
						js.Global().Call("alert", "YOU WIN, CONGRATULATIONS!")
						js.Global().Get("document").Get("location").Call("reload")
					}
				}
			}
		}
	}
}

func (g *GameManager) draw() {
	fmt.Println(g.ball.dx)
	fmt.Println(g.ball.dy)
	g.canvas.ctx.Call("clearRect", 0, 0, g.canvas.width, g.canvas.height)
	g.ball.draw(g.canvas.ctx)
	g.paddle.draw(g.canvas.ctx)
	g.bricks.draw(g.canvas.ctx)
	g.score.draw(g.canvas.ctx)
	g.live.draw(g.canvas.ctx)
	g.collisionDetection()

	g.ball.x += g.ball.dx
	g.ball.y += g.ball.dy

	if g.ball.x+g.ball.dx > g.canvas.width-g.ball.radius || g.ball.x+g.ball.dx < g.ball.radius {
		g.ball.dx = -g.ball.dx
	}

	if g.ball.y+g.ball.dy < g.ball.radius {
		g.ball.dy = -g.ball.dy
	} else if g.ball.y+g.ball.dy > g.canvas.height-g.ball.radius {
		if g.ball.x > g.paddle.x && g.ball.x < g.paddle.x+g.paddle.width {
			g.ball.dy = -g.ball.dy
		} else {
			g.live.live--
			if g.live.live < 0 {
				js.Global().Call("alert", "GAME OVER")
				js.Global().Get("document").Get("location").Call("reload")
			} else {
				g.ball.x = g.canvas.width / 2
				g.ball.y = g.canvas.height - 30
				g.ball.dx = 2
				g.ball.dy = -2
				g.paddle.x = (g.canvas.width - g.paddle.width) / 2
			}
		}
	}

	if g.rightPressed && g.paddle.x < g.canvas.width-g.paddle.width {
		g.paddle.x += g.paddle.dx
	} else if g.leftPressed && g.paddle.x > 0 {
		g.paddle.x -= g.paddle.dx
	}

	js.Global().Call("requestAnimationFrame", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		g.draw()
		return nil
	}))
}

func (b *Ball) draw(ctx js.Value) {
	ctx.Call("beginPath")
	ctx.Call("arc", b.x, b.y, b.radius, 0, math.Pi*2)
	ctx.Set("fillStyle", "#0095DD")
	ctx.Call("fill")
	ctx.Call("closePath")
}

func (p *Paddle) draw(ctx js.Value) {
	ctx.Call("beginPath")
	ctx.Call("rect", p.x, p.y, p.width, p.height)
	ctx.Set("fillStyle", "#0095DD")
	ctx.Call("fill")
	ctx.Call("closePath")
}

func (b *Bricks) draw(ctx js.Value) {
	for i := 0; i < b.columnCount; i++ {
		for j := 0; j < b.rowCount; j++ {
			if b.bricks[i][j].status == 1 {
				brickX := (i * (b.width + b.padding)) + b.offsetLeft
				brickY := (j * (b.height + b.padding)) + b.offsetTop
				b.bricks[i][j].x = brickX
				b.bricks[i][j].y = brickY
				ctx.Call("beginPath")
				ctx.Call("rect", brickX, brickY, b.width, b.height)
				ctx.Set("fillStyle", "#0095DD")
				ctx.Call("fill")
				ctx.Call("closePath")
			}
		}
	}
}

func (s *Score) draw(ctx js.Value) {
	ctx.Set("font", "16px Arial")
	ctx.Set("fillStyle", "#0095DD")
	ctx.Call("fillText", fmt.Sprintf("Score: %d", s.score), s.x, s.y)
}

func (l *Live) draw(ctx js.Value) {
	ctx.Set("font", "16px Arial")
	ctx.Set("fillStyle", "#0095DD")
	ctx.Call("fillText", fmt.Sprintf("Lives: %d", l.live), l.x, l.y)
}
