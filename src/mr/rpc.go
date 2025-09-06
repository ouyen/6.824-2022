package mr

//
// RPC definitions.
//
// remember to capitalize all names.
//

import (
	"container/list"
	"os"
	"time"
)
import "strconv"

//
// example to show how to declare the arguments
// and reply for an RPC.
//

type ExampleArgs struct {
	X int
}

type ExampleReply struct {
	Y int
}

// Add your RPC definitions here.

// Cook up a unique-ish UNIX-domain socket name
// in /var/tmp, for the coordinator.
// Can't use the current directory since
// Athena AFS doesn't support UNIX-domain sockets.
func coordinatorSock() string {
	s := "/var/tmp/824-mr-"
	s += strconv.Itoa(os.Getuid())
	return s
}

type TaskType int

const (
	Map TaskType = iota
	Reduce
	Wait
	Exit
)

type Task struct {
	Type      TaskType // "Map", "Reduce", "Wait", "Exit"
	ID        int
	FileName  string // Map任务输入文件
	NReduce   int    // Reduce任务数量
	StartTime time.Time
	// 其他你需要的字段
}

type TaskStatus int

type TaskList struct {
	Tasks list.List
	IdMap map[int]*list.Element
}

func MakeTaskList() *TaskList {
	return &TaskList{
		Tasks: list.List{},
		IdMap: make(map[int]*list.Element),
	}
}

type RequestTaskArgs struct {
}

type RequestTaskReply struct {
	Task Task
}

type ReportTaskArgs struct {
	TaskID int
	Type   TaskType
}

type ReportTaskReply struct{}
