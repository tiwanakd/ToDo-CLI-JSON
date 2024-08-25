package main

import (
	"fmt"

	"github.com/tiwanakd/ToDo-CLI-JSON/tasks"
)

func main() {
	// if err := cmd.Execute(); err != nil {
	// 	fmt.Fprintln(os.Stderr, err)
	// }
	var t tasks.Task
	if err := t.CompleteTask(2, 3, 4); err != nil {
		fmt.Println(err)
	}

}
