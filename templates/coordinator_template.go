package mr

import (
	"log"
	"net"
	"net/http"
	"net/rpc"
	"os"
	"sync"
	"time"
)

//
// Coordinator implementation template for Lab1
//

type TaskState int

const (
	TaskIdle TaskState = iota
	TaskInProgress
	TaskCompleted
)

type Task struct {
	ID        int
	State     TaskState
	Filename  string    // For map tasks
	StartTime time.Time // For timeout detection
}

type Coordinator struct {
	mu          sync.Mutex
	files       []string  // Input files
	nReduce     int       // Number of reduce tasks
	mapTasks    []Task    // Map task states
	reduceTasks []Task    // Reduce task states
	phase       string    // "map", "reduce", "done"
}

//
// RPC handler example - you'll need to implement these
//
func (c *Coordinator) RequestTask(args *TaskRequest, reply *TaskReply) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// TODO: Implement task assignment logic
	// 1. Check current phase (map or reduce)
	// 2. Find an idle task to assign
	// 3. Update task state to in-progress
	// 4. Fill in reply with task details

	log.Printf("Worker requested task - need to implement assignment logic")
	reply.TaskType = "wait" // Placeholder
	return nil
}

func (c *Coordinator) CompleteTask(args *TaskCompleteRequest, reply *TaskCompleteReply) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// TODO: Implement task completion logic
	// 1. Mark the specified task as completed
	// 2. Check if all tasks in current phase are done
	// 3. If so, advance to next phase (map->reduce->done)

	log.Printf("Worker completed %s task %d - need to implement completion logic", 
		args.TaskType, args.TaskID)
	return nil
}

//
// Helper methods you might find useful
//

func (c *Coordinator) allMapTasksCompleted() bool {
	for _, task := range c.mapTasks {
		if task.State != TaskCompleted {
			return false
		}
	}
	return true
}

func (c *Coordinator) allReduceTasksCompleted() bool {
	for _, task := range c.reduceTasks {
		if task.State != TaskCompleted {
			return false
		}
	}
	return true
}

func (c *Coordinator) initializeTasks() {
	// Initialize map tasks
	c.mapTasks = make([]Task, len(c.files))
	for i, file := range c.files {
		c.mapTasks[i] = Task{
			ID:       i,
			State:    TaskIdle,
			Filename: file,
		}
	}

	// Initialize reduce tasks  
	c.reduceTasks = make([]Task, c.nReduce)
	for i := 0; i < c.nReduce; i++ {
		c.reduceTasks[i] = Task{
			ID:    i,
			State: TaskIdle,
		}
	}

	c.phase = "map"
}

//
// Start a thread that listens for RPCs from worker.go
//
func (c *Coordinator) server() {
	rpc.Register(c)
	rpc.HandleHTTP()
	sockname := coordinatorSock()
	os.Remove(sockname)
	l, e := net.Listen("unix", sockname)
	if e != nil {
		log.Fatal("listen error:", e)
	}
	go http.Serve(l, nil)
}

//
// main/mrcoordinator.go calls Done() periodically to find out
// if the entire job has finished.
//
func (c *Coordinator) Done() bool {
	c.mu.Lock()
	defer c.mu.Unlock()

	// TODO: Return true when all work is finished
	return c.phase == "done"
}

//
// Create a Coordinator.
// main/mrcoordinator.go calls this function.
// nReduce is the number of reduce tasks to use.
//
func MakeCoordinator(files []string, nReduce int) *Coordinator {
	c := Coordinator{
		files:   files,
		nReduce: nReduce,
	}

	// TODO: Initialize your coordinator state
	c.initializeTasks()

	c.server()
	return &c
}