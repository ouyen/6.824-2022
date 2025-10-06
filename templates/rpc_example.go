package mr

//
// RPC definitions for Lab1 MapReduce
//
// Remember to capitalize all names for RPC visibility.
//

import "os"
import "strconv"

//
// Example RPC (already provided)
//
type ExampleArgs struct {
	X int
}

type ExampleReply struct {
	Y int
}

//
// TODO: Add your RPC definitions here for Lab1
//
// Here are some suggestions to get you started:

// Worker requests a task from coordinator
type TaskRequest struct {
	WorkerID string // Optional: worker identifier
}

type TaskReply struct {
	TaskType   string // "map", "reduce", "wait", "done"
	TaskID     int    // Task number
	Filename   string // Input file for map tasks
	NReduce    int    // Number of reduce tasks
	NMap       int    // Number of map tasks
}

// Worker reports task completion
type TaskCompleteRequest struct {
	TaskType string // "map" or "reduce"
	TaskID   int    // Which task was completed
}

type TaskCompleteReply struct {
	// Can be empty
}

// Cook up a unique-ish UNIX-domain socket name
// in /var/tmp, for the coordinator.
// Can't use the current directory since
// Athena AFS doesn't support UNIX-domain sockets.
func coordinatorSock() string {
	s := "/var/tmp/824-mr-"
	s += strconv.Itoa(os.Getuid())
	return s
}