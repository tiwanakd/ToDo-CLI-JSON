package cmd

import (
	"strconv"

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
		var err error
		if listAll {
			err = tasks.ListTasks(true)
		} else {
			err = tasks.ListTasks(false)
		}
		return err
	},
}
var task tasks.Task
var cmdAddTask = cobra.Command{
	Use:   "add [new task name]",
	Short: "add new task",
	Long:  "add new task with the name passed as argument",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return task.AddTask(args...)
	},
}

var cmdCompleteTask = cobra.Command{
	Use:   "complete [taskids]",
	Short: "complete tasks",
	Long:  "compelte all the tasks with the provided ids",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ids, err := convertIdstoInt(args)
		if err != nil {
			return err
		}
		return task.CompleteTask(ids...)
	},
}

var cmdDeleteTask = cobra.Command{
	Use:   "delete [taskids]",
	Short: "delete tasks",
	Long:  "delete the tasks with given id(S)",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ids, err := convertIdstoInt(args)
		if err != nil {
			return err
		}
		return task.DeleteTask(ids...)
	},
}

func convertIdstoInt(args []string) ([]int, error) {
	intSlice := make([]int, len(args))
	for i, arg := range args {
		id, err := strconv.Atoi(arg)
		if err != nil {
			return nil, err
		}
		intSlice[i] = id
	}
	return intSlice, nil
}

func init() {
	cmdListTasks.Flags().BoolVarP(&listAll, "all", "a", false, "list all the tasks comleted and uncompleted")

	rootCmd.AddCommand(&cmdListTasks)
	rootCmd.AddCommand(&cmdAddTask)
	rootCmd.AddCommand(&cmdCompleteTask)
	rootCmd.AddCommand(&cmdDeleteTask)
}

func Execute() error {
	if err := rootCmd.Execute(); err != nil {
		return err
	}
	return nil
}
