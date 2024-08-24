package main

import "github.com/tiwanakd/ToDo-CLI-JSON/tasks"

func main() {
	// if err := cmd.Execute(); err != nil {
	// 	fmt.Fprintln(os.Stderr, err)
	// }
	var t tasks.Task
	t.CompleteTask(1)
}
