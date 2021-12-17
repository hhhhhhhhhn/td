package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	gc "github.com/rthornton128/goncurses"
)

func main() {
	window, screen, _ := initWindow()
	defer screen.Delete()
	defer screen.End()
	defer window.Delete()

	currentTodo, err := Load()
	if err != nil || currentTodo == nil {
		fmt.Fprintln(os.Stderr, err)
		currentTodo = &Todo{Title: "Home", Of: 10}
		currentTodo.AddChild(0)
		currentTodo.Children[0].Title = "Your first todo"
	}

	setupCloseHandler(&currentTodo)

	scroll := 0
	selection := 0

	renderTodoChildren(window, currentTodo, scroll, selection)
	renderLocation(window, currentTodo)

	for {
		switch window.GetChar() {
		case gc.KEY_ESC:
			err := Save(currentTodo)
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
			}
			gc.End()
			return
		case 'j':
			if selection < len(currentTodo.Children) - 1 {
				selection++
				if scroll + height - offset < selection + 1 {
					scroll++
				}
			}
			break
		case 'k':
			if selection > 0 {
				selection--
				if selection < scroll {
					scroll--
				}
			}
			break
		case 'l':
			if len(currentTodo.Children[selection].Children) == 0 {
				currentTodo.Children[selection].AddChild(0)
			}
			currentTodo = currentTodo.Children[selection]
			window.Erase()
			renderLocation(window, currentTodo)
			selection = 0
			scroll = 0
		case 'h':
			if currentTodo.Parent != nil {
				currentTodo = currentTodo.Parent
				window.Erase()
				renderLocation(window, currentTodo)
				selection = 0
				scroll = 0
			}
		case 'K':
			if len(currentTodo.Children[selection].Children) == 0 {
				currentTodo.Children[selection].Done++
				currentTodo.UpdateDoneRecursive()
			}
			break
		case 'J':
			if len(currentTodo.Children[selection].Children) == 0 {
				currentTodo.Children[selection].Done--
				currentTodo.UpdateDoneRecursive()
			}
			break
		case 'L':
			currentTodo.Children[selection].Of++
			if len(currentTodo.Children[selection].Children) > 0 {
				currentTodo.Children[selection].UpdateDoneRecursive()
			}
			break
		case 'H':
			currentTodo.Children[selection].Of--
			if len(currentTodo.Children[selection].Children) > 0 {
				currentTodo.Children[selection].UpdateDoneRecursive()
			}
			break
		case 'd':
			if window.GetChar() == 'd' { // double press
				currentTodo.Children = append(
					currentTodo.Children[:selection], 
					currentTodo.Children[selection+1:]..., 
				)
				if selection == len(currentTodo.Children) && selection != 0{
					selection--
				}
			}
			if len(currentTodo.Children) == 0 {
				if currentTodo.Parent == nil {
					currentTodo.AddChild(0)
				} else {
					currentTodo = currentTodo.Parent
					selection = 0
					scroll = 0
				}
			}
			window.Erase()
			renderLocation(window, currentTodo)
			break
		case 'e':
			var x int
			if len(currentTodo.Children[selection].Children) == 0 {
				x = 0
			} else {
				x = 2
			}
			currentTodo.Children[selection].Title = enterEditMode(
				window,
				currentTodo.Children[selection].Title,
				selection - scroll + offset,
				x,
			)
			break
		case 'o':
			selection++
			if scroll + height - offset < selection + 1 {
				scroll++
			}
			currentTodo.AddChild(selection)
			renderTodoChildren(window, currentTodo, scroll, selection)
			currentTodo.Children[selection].Title = enterEditMode(
				window,
				currentTodo.Children[selection].Title,
				selection - scroll + offset,
				0,
			)
			break
		case 'r':
			window.Clear()
			renderLocation(window, currentTodo)
			Save(currentTodo)
			break
		}
	renderTodoChildren(window, currentTodo, scroll, selection)
	}
}

func setupCloseHandler(currentTodo **Todo) {
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-c
		err := Save(*currentTodo)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
		}
		gc.End()
		os.Exit(0)
	}()
}
