package main

import (
	"fmt" 
	"os"
	"os/signal"
	"syscall"
	"sort"

	gc "github.com/rthornton128/goncurses"
)

var scroll, selection int
var currentTodo *Todo
var window *gc.Window
var screen *gc.Screen

func main() {
	var err error
	window, screen, err = initWindow()
	defer screen.Delete()
	defer screen.End()
	defer window.Delete()

	currentTodo, err = Load()
	if err != nil || currentTodo == nil {
		fmt.Fprintln(os.Stderr, err)
		currentTodo = &Todo{Title: "Home", Of: 10}
		currentTodo.AddChild(0)
		currentTodo.Children[0].Title = "Your first todo"
	}

	setupCloseHandler(&currentTodo)

	renderTodoChildren(currentTodo, scroll, selection)
	renderLocation(currentTodo)

	for {
		switch window.GetChar() {
		case gc.KEY_ESC, 'q':
			err := Save(currentTodo)
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
			}
			gc.End()
			return
		case 'j':
			changeSelection(selection + 1)
			break
		case 'k':
			changeSelection(selection - 1)
			break
		case 'l':
			if !hasChildren(currentTodo.Children[selection]) {
				currentTodo.Children[selection].AddChild(0)
			}
			navigateTo(currentTodo.Children[selection])
			break
		case 'h':
			if currentTodo.Parent != nil {
				navigateTo(currentTodo.Parent)
			}
			break
		case 'K':
			if !hasChildren(currentTodo.Children[selection]) {
				currentTodo.Children[selection].Done++
				currentTodo.UpdateRecursive()
			}
			break
		case 'J':
			if !hasChildren(currentTodo.Children[selection]) {
				currentTodo.Children[selection].Done--
				currentTodo.UpdateRecursive()
			}
			break
		case 'L':
			if !hasChildren(currentTodo.Children[selection]) {
				currentTodo.Children[selection].Of++
				currentTodo.UpdateRecursive()
			}
			break
		case 'H':
			if !hasChildren(currentTodo.Children[selection]) {
				currentTodo.Children[selection].Of--
				currentTodo.UpdateRecursive()
			}
			break
		case 'd':
			if window.GetChar() == 'd' { // double press
				deleteSelection()
			}
			break
		case 'e', 'A', 'a':
			editSelection(len(currentTodo.Children[selection].Title))
			break
		case 'I', 'i':
			editSelection(0)
			break
		case 'o':
			currentTodo.AddChild(selection + 1)
			changeSelection(selection + 1)
			renderTodoChildren(currentTodo, scroll, selection)
			editSelection(0)
			break
		case 'O':
			currentTodo.AddChild(selection)
			changeSelection(selection)
			renderTodoChildren(currentTodo, scroll, selection)
			editSelection(0)
			break
		case 'r':
			window.Clear()
			sortTodos()
			renderLocation(currentTodo)
			break
		}
	renderTodoChildren(currentTodo, scroll, selection)
	}
}

func handleScroll() {
	if scroll + height - offset < selection + 1 {
		scroll++
	} else if selection < scroll {
		scroll--
	}
}

func changeSelection(newSelection int) {
	if newSelection >= 0 && newSelection < len(currentTodo.Children) {
		selection = newSelection
	}
	handleScroll()
}

func navigateTo(todo *Todo) {
	currentTodo = todo
	sortTodos()
	window.Erase()
	renderLocation(currentTodo)
	selection = 0
	scroll = 0
}

// Sorts by work yet to be done
// Overriden by numbers
func sortTodos() {
	sort.SliceStable(currentTodo.Children, func(i, j int) bool {
		iStart := startingNumber(currentTodo.Children[i].Title)
		jStart := startingNumber(currentTodo.Children[j].Title)

		if iStart != jStart {
			return iStart < jStart
		}

		return (
			(currentTodo.Children[i].Of - currentTodo.Children[i].Done) >
			(currentTodo.Children[j].Of - currentTodo.Children[j].Done))
	})
}

func startingNumber(str string) int {
	value := 0
	for i, char := range str {
		digit := asDigit(char)
		if i == 0 && digit == -1 {
			return 9999999 // So that all non-numbered todos go at the end
		}
		if digit == -1 {
			break
		}
		value = value*10 + digit
	}
	return value
}

// Returns -1 if is not a digit
func asDigit(char rune) int {
	value := char - '0'
	if value >= 10 || value < 0 {
		return -1
	}
	return int(value)
}

func hasChildren(todo *Todo) bool {
	return len(todo.Children) > 0
}

func deleteSelection() {
	currentTodo.Children = append(
		currentTodo.Children[:selection], 
		currentTodo.Children[selection+1:]..., 
	)
	if selection == len(currentTodo.Children) && selection != 0{
		selection--
	}
	if !hasChildren(currentTodo) {
		if currentTodo.Parent == nil {
			currentTodo.AddChild(0)
		} else {
			navigateTo(currentTodo.Parent)
		}
	}
	currentTodo.UpdateRecursive()
}

func editSelection(cursor int) {
	var x int
	if len(currentTodo.Children[selection].Children) == 0 {
		x = 0
	} else {
		x = 2
	}
	currentTodo.Children[selection].Title = enterEditMode(
		currentTodo.Children[selection].Title,
		cursor,
		selection - scroll + offset,
		x,
	)
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
		screen.Delete()
		screen.End()
		window.Delete()
		gc.End()
		os.Exit(0)
	}()
}
