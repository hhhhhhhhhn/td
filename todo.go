package main

import (
	"encoding/json"
	"io/ioutil"
	"os"
)

type Todo struct {
	Done     int
	Of       int
	Title    string
	Parent   *Todo
	Children []*Todo
}

// The sum of done and ofs of the children
func (t *Todo) CalculateChildrenDoneAndOf() (done, of int) {
	for _, child := range t.Children {
		done += child.Done
		of += child.Of
	}
	return done, of
}

func (t *Todo) UpdateDoneRecursive() {
	for {
		childrenDone, childrenOf := t.CalculateChildrenDoneAndOf()

		if childrenOf == 0 {
			t.Done = 0
		} else {
			t.Done = t.Of * childrenDone / childrenOf
		}

		t = t.Parent

		if t == nil {
			return
		}
	}
}

func (parent *Todo) AddChild(index int) *Todo {
	todo := &Todo{Done: 0, Of: 1, Title: "", Parent: parent, Children: []*Todo{}}
	parent.Children = append(parent.Children, nil)
	for i:=len(parent.Children)-1; i>index; i--{
		parent.Children[i] = parent.Children[i-1]
	}
	parent.Children[index] = todo
	parent.UpdateDoneRecursive()
	return todo
}

// Needed to fix recursion issues
type jsonTodo struct {
	Done     int         `json:"done"`
	Of       int         `json:"of"`
	Title    string      `json:"title"`
	Children []*jsonTodo `json:"children"`
}

func todoToJsonTodo(todo *Todo) *jsonTodo {
	var children []*jsonTodo
	for _, child := range todo.Children {
		children = append(children, todoToJsonTodo(child))
	}

	return &jsonTodo {
		Done: todo.Done,
		Of:   todo.Of,
		Title: todo.Title,
		Children: children,
	}
}

func jsonTodoToTodo(jsonTodo *jsonTodo) *Todo {
	todo := &Todo{
		Done: jsonTodo.Done,
		Of: jsonTodo.Of,
		Title: jsonTodo.Title,
	}

	for _, jsonChild := range jsonTodo.Children {
		child := jsonTodoToTodo(jsonChild)
		child.Parent = todo
		todo.Children = append(todo.Children, child)
	}

	return todo
}

func Save(todo *Todo) error {
	for todo.Parent != nil {
		todo = todo.Parent
	}
	data, err := json.Marshal(todoToJsonTodo(todo))

	if err == nil {
		ioutil.WriteFile(os.Getenv("HOME") + "/.td", data, 0644)
		return nil
	}
	return err
}

func Load() (*Todo, error) {
	var todo *jsonTodo
	data, err := ioutil.ReadFile(os.Getenv("HOME") + "/.td")
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(data, &todo)
	if err != nil {
		return nil, err
	}
	return jsonTodoToTodo(todo), nil
}
