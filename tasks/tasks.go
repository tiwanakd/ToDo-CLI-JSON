package tasks

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
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
	file, err := os.OpenFile(jsonFileName, os.O_RDWR, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	dec := json.NewDecoder(file)

	//read the open bracket
	_, err = dec.Token()
	if err != nil {
		return err
	}

	idFound := false

	for _, id := range ids {
		var matchedInputOffset int64
	inner:
		for dec.More() {
			currentInputOffSet := dec.InputOffset()
			err := dec.Decode(&t)

			if err != nil {
				return err
			}

			if t.Id == id {
				t.IsComplete = true
				nowStr := time.Now().Format(time.RFC3339)
				parsedTime, err := time.Parse(time.RFC3339, nowStr)
				if err != nil {
					return err
				}
				t.CompletedAt = parsedTime
				matchedInputOffset = currentInputOffSet - matchedInputOffset

				fmt.Println(matchedInputOffset)
				if matchedInputOffset == 1 {
					file.Seek(matchedInputOffset+4, io.SeekStart)
				} else {
					file.Seek(matchedInputOffset+5, io.SeekStart)
				}
				updatedTask, err := json.MarshalIndent(t, "\t", "\t")
				if err != nil {
					return err
				}
				updatedTask = updatedTask[2 : len(updatedTask)-2]
				fmt.Println(string(updatedTask))
				file.Write(updatedTask)

				idFound = true
				break inner
			}
			if err == io.EOF {
				idFound = false
				break
			}
		}
	}

	if !idFound {
		return fmt.Errorf("no match for any of the given id(s)")
	}

	// read closing bracket
	_, err = dec.Token()
	if err != nil {
		return err
	}

	return nil
}
