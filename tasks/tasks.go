package tasks

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
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
				fmt.Fprintf(tw, "%d\t%s\t%s\t%v\n", task.Id, task.Name, timediff.TimeDiff(task.CreatedAt), task.IsComplete)
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

func (t Task) AddTask(name string) error {
	lastId, err := getLastId()
	if err != nil {
		return err
	}
	newTaskId := lastId + 1
	t.Id = newTaskId
	t.Name = name

	//covert time to the required string formant and Parse it
	nowStr := time.Now().Format(time.RFC3339)
	parsedTime, err := time.Parse(time.RFC3339, nowStr)
	if err != nil {
		return err
	}
	t.CreatedAt = parsedTime
	t.IsComplete = false
	t.CompletedAt = time.Time{}

	newTask, err := json.Marshal(t)
	if err != nil {
		return err
	}

	file, err := os.OpenFile(jsonFileName, os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	//if the last id is zero this will imply there is no data currently in the file
	//write the open and closing brakcet as this is required for Json Unmarshlling
	if lastId == 0 {
		file.Write([]byte("["))
		defer file.Write([]byte("]"))
	} else {
		//if the file already has tasks last bracket has to removed so next task can be added
		fileInfo, err := file.Stat()
		if err != nil {
			return err
		}

		//Truncate the last byte from the file which is the closing bracket
		file.Truncate(fileInfo.Size() - 1)
		//write the commma and a new line charater
		file.Write([]byte(","))
		file.Write([]byte("\n"))
		defer file.Write([]byte("]"))
	}
	if _, err := file.Write(newTask); err != nil {
		return err
	}

	return nil
}
