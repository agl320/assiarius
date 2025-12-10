package main

import (
	"assiarius/internal/debug"
	"assiarius/internal/root"
	"fmt"
)

func main() {
	fmt.Println(debug.Debug("Assiarius is starting..."))
	root.Execute()
}
