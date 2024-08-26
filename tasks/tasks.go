package tasks

import (
	"cmp"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"slices"
	"text/tabwriter"
	"time"

	"github.com/mergestat/timediff"
)

type Task struct {
	Id          int       `json:"id"`
	Name        string    `json:"name"`
	CreatedAt   time.Time `json:"created_at"`
	IsComplete  bool      `json:"is_complete"`
	CompletedAt time.Time `json:"completed_at"`
}

const jsonFileName = "tasks.json"

var errNoTasks = errors.New("no tasks found, please add tasks")

func allTasks() ([]Task, error) {
	data, err := os.ReadFile(jsonFileName)
	if err != nil {
		return nil, err
	}

	if len(data) == 0 {
		return nil, errNoTasks
	}
	var tasks []Task
	err = json.Unmarshal(data, &tasks)
	if err != nil {
		return nil, err
	}

	return tasks, nil
}

func ListTasks(listAll bool) error {
	tasks, err := allTasks()
	if err != nil {
		return err
	}

	tw := tabwriter.NewWriter(os.Stdout, 0, 2, 3, ' ', tabwriter.TabIndent)
	if listAll {
		fmt.Fprint(tw, "ID\tName\tCreated\tCompleted\tCompletedAt\n")
		for _, task := range tasks {
			if task.IsComplete {
				fmt.Fprintf(tw, "%d\t%s\t%s\t%v\t%s\n", task.Id, task.Name, timediff.TimeDiff(task.CreatedAt), task.IsComplete, task.CompletedAt.Format(time.RFC1123))
			} else {
				fmt.Fprintf(tw, "%d\t%s\t%s\t%v\t%s\n", task.Id, task.Name, timediff.TimeDiff(task.CreatedAt), task.IsComplete, "N/A")
			}
		}
	} else {
		fmt.Fprintf(tw, "ID\tName\tCreated\n")
		for _, task := range tasks {
			if !task.IsComplete {
				fmt.Fprintf(tw, "%d\t%s\t%s\n", task.Id, task.Name, timediff.TimeDiff(task.CreatedAt))
			}
		}
	}
	tw.Flush()

	return nil
}

// get the id of the last task as the new tasks will increment on it
func getLastId() (int, error) {
	tasks, err := allTasks()
	if err == errNoTasks {
		return 0, nil
	}
	if err != nil {
		return -1, err
	}

	var lastid int
	for _, task := range tasks {
		lastid = task.Id
	}

	return lastid, nil
}

func (t Task) AddTask(names ...string) error {
	lastId, err := getLastId()
	if err != nil {
		return err
	}

	//increment the last Id based on the lastid from the file
	taskId := lastId + 1

	//using a slice wherer the new tasks will be placed
	tSlice := make([]Task, len(names))

	for i, name := range names {
		t.Id = taskId
		t.Name = name

		//covert time to the required string formant and Parse it
		nowStr := time.Now().Format(time.RFC3339)
		parsedTime, err := time.Parse(time.RFC3339, nowStr)
		if err != nil {
			return err
		}
		t.CreatedAt = parsedTime
		t.IsComplete = false

		nilParsedTime, err := time.Parse(time.RFC3339, "2006-01-02T15:04:05-07:00")
		if err != nil {
			return err
		}
		t.CompletedAt = nilParsedTime
		tSlice[i] = t

		taskId++
	}

	newTask, err := json.MarshalIndent(tSlice, "", "\t")
	if err != nil {
		return err
	}

	file, err := os.OpenFile(jsonFileName, os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	//if the lastId is 0 which implies no tasks were added yet
	//remove the required elemers for proper json formatting
	if lastId != 0 {
		newTask = newTask[1:] // remove the first element of byte as this is [
		fileInfo, err := file.Stat()
		if err != nil {
			return err
		}
		//Truncate the last byte from the file which is the closing bracket
		file.Truncate(fileInfo.Size() - 1)
		//There is space added which will also be removed
		file.Truncate(fileInfo.Size() - 2)
		file.Write([]byte(","))
	}

	if _, err := file.Write(newTask); err != nil {
		return err
	}

	return nil
}

func (t Task) CompleteTask(ids ...int) error {
	tasks, err := allTasks()
	if err != nil {
		return err
	}

	//create a map to hold all the ids provided
	//set the values to false; these will be set to true if id match is found
	//this allows to track all the invalid ids
	idMap := make(map[int]bool)
	for _, id := range ids {
		idMap[id] = false
	}

	for _, id := range ids {
	inner:
		for i := range tasks {
			taskId := tasks[i].Id
			if taskId == id {
				tasks[i].IsComplete = true
				parsedTime, err := time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
				if err != nil {
					return err
				}
				tasks[i].CompletedAt = parsedTime
				idMap[id] = true
				fmt.Fprintln(os.Stdout, "Task Completed:", tasks[i].Name)
				break inner
			}
		}
	}

	//print out the invalid ids
	for id, found := range idMap {
		if !found {
			fmt.Fprintln(os.Stderr, "invalid id:", id)
		}
	}

	completedTasks, err := json.MarshalIndent(tasks, "", "\t")
	if err != nil {
		return err
	}

	return os.WriteFile(jsonFileName, completedTasks, 0644)
}

func (t Task) DeleteTask(ids ...int) error {

	tasks, err := allTasks()
	if err != nil {
		return err
	}

	idMap := make(map[int]bool)
	for _, id := range ids {
		idMap[id] = false
	}

	//create a slice that will hold the indexes of task to be deleted
	var deleteIndexs []int

	for _, id := range ids {
		for i, task := range tasks {
			if task.Id == id {
				idMap[id] = true
				deleteIndexs = append(deleteIndexs, i)
			}
		}
	}

	for id, found := range idMap {
		if !found {
			fmt.Fprintln(os.Stderr, "invalid id:", id)
		}
	}

	//procced to delete if there are any deleteindexes
	if len(deleteIndexs) > 0 {
		//sort the deleteindex slice in decending order before proceeding to delete from the tasks slice
		//this will ensusre that there is no out of bounds panic
		slices.SortFunc(deleteIndexs, func(a, b int) int {
			return cmp.Compare(b, a)
		})

		for _, deleteIndex := range deleteIndexs {
			fmt.Fprintln(os.Stdout, "deleting task:", tasks[deleteIndex].Name)
			tasks = slices.Delete(tasks, deleteIndex, deleteIndex+1)
		}

		updatedTasks, err := json.MarshalIndent(tasks, "", "\t")
		if err != nil {
			return err
		}

		if err := os.WriteFile(jsonFileName, updatedTasks, 0644); err != nil {
			return err
		}
	}
	return nil
}
