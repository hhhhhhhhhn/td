package main

import (
	"os"
	"strings"

	gc "github.com/rthornton128/goncurses"
)

var width, height int // Terminal width and height

func initColors() {
	gc.InitPair(1, gc.C_WHITE, gc.C_BLACK)
	gc.InitPair(2, gc.C_BLACK, gc.C_WHITE)
	gc.InitPair(3, gc.C_WHITE, gc.C_BLUE)
}

func initWindow() (*gc.Window, *gc.Screen, error) {
	screen, err := gc.NewTerm("", os.Stdout, os.Stdin)
	if err != nil {
		return nil, nil, err
	}
	screen.Set()
	
	gc.CBreak(true)
	gc.Echo(false)
	gc.SetEscDelay(10)
	gc.Cursor(0)

	window := gc.StdScr()
	window.Keypad(true)
	height, width = window.MaxYX()

	gc.StartColor()
	//gc.UseDefaultColors()
	initColors()

	return window, screen, nil
}

var offset = 1

func renderTodoChildren(todo *Todo, scroll, selection int) {
	for y:=offset; y<height; y++ {
		todoIndex := scroll + y - offset

		if todoIndex >= len(todo.Children) {
			window.Move(y, 0)
			window.ClearToBottom()
			break
		}

		var color int16

		if todoIndex == selection {
			color = 2
		} else {
			color = 1
		}

		window.ColorOn(color)
		window.Move(y, 0)
		window.ClearToEOL()

		if todoIndex == selection {
			window.Print(strings.Repeat(" ", width))
		}

		if !hasChildren(todo.Children[todoIndex]) {
			window.MovePrint(y, 0, todo.Children[todoIndex].Title)
			window.AttrOn(gc.A_UNDERLINE)
		} else {
			window.MovePrint(y, 0, "> " + todo.Children[todoIndex].Title)
		}

		window.MovePrintf(y, width - 8, "%2d", todo.Children[todoIndex].Done)

		window.AttrOff(gc.A_UNDERLINE)
		window.MovePrint(y, width - 5, "/")

		if !hasChildren(todo.Children[todoIndex]) {
			window.AttrOn(gc.A_UNDERLINE)
		}
		window.MovePrintf(y, width - 3, "%2d", todo.Children[todoIndex].Of)
		window.AttrOff(gc.A_UNDERLINE)
	}
}

func renderLocation(todo *Todo) {
	location := todo.Title
	for {
		todo = todo.Parent
		if todo == nil {
			break
		}
		location = todo.Title + " > " + location
	}
	window.ColorOn(3)
	window.MovePrint(0, 0, center(location, width))
}

func center(str string, width int) string {
	paddingLeft := (width - len(str)) / 2
	paddingRight := width - paddingLeft - len(str)
	return strings.Repeat(" ", paddingLeft) + str + strings.Repeat(" ", paddingRight)
}

func enterEditMode(str string, y int, x int) string {
	gc.Cursor(1)
	originalCopy := str
	cursor := len(str)
	window.ColorOn(2)
	window.MovePrint(y, x, str)
	for {
		chr := window.GetChar()
		switch chr {
		case gc.KEY_LEFT:
			if cursor > 0 {
				cursor--
			}
			break
		case gc.KEY_RIGHT:
			if cursor < len(str) {
				cursor++
			}
		case gc.KEY_BACKSPACE:
			if cursor > 0 {
				str = str[:cursor - 1] + str[cursor:]
				cursor--
				window.MovePrint(y, x + len(str), " ")
			}
			break
		case '\n':
			gc.Cursor(0)
			return str
		case gc.KEY_ESC:
			gc.Cursor(0)
			return originalCopy
		default:
			str = str[:cursor] + string(chr) + str[cursor:]
			cursor++
			break
		}
		window.MovePrint(y, x, str)
		window.Move(y, x + cursor)
	}
}
