package cmd

import (
	"github.com/spf13/cobra"
	"github.com/tiwanakd/ToDo-CLI-JSON/tasks"
)

var rootCmd = cobra.Command{
	Use: "tasks",
}

var listAll bool
var cmdListTasks = cobra.Command{
	Use:   "list",
	Short: "list the tasks",
	Long:  "list all the tasks completed and uncompleted as per the flag passed",
	RunE: func(cmd *cobra.Command, args []string) error {
		if listAll {
			if err := tasks.ListTasks(true); err != nil {
				return err
			}
		} else {
			if err := tasks.ListTasks(false); err != nil {
				return err
			}
		}
		return nil
	},
}
var cmdAddTask = cobra.Command{
	Use:   "add [new task name]",
	Short: "add new task",
	Long:  "add new task with the name passed as argument",
	RunE: func(cmd *cobra.Command, args []string) error {
		var task tasks.Task
		return task.AddTask(args...)
	},
}

func init() {
	cmdListTasks.Flags().BoolVarP(&listAll, "all", "a", false, "list all the tasks comleted and uncompleted")

	rootCmd.AddCommand(&cmdListTasks)
	rootCmd.AddCommand(&cmdAddTask)
}

func Execute() error {
	if err := rootCmd.Execute(); err != nil {
		return err
	}
	return nil
}
