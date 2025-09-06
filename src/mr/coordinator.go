package mr

import (
	"log"
	"sync"
	"time"
)
import "net"
import "os"
import "net/rpc"
import "net/http"

// TODO:
//  1. Add crash-resume (detect timeout)
//  2. Rewrite RPC
//     2.1 Use a single RPC for both requesting and reporting tasks
//     2.2 Maybe use channel instead of RPC
//  3. Optimize the locking strategy
type Coordinator struct {
	// Your definitions here.
	NReduce       int
	AllDone       bool
	mu            sync.Mutex
	mapTasks      TaskList
	reduceTasks   TaskList
	inputFiles    []string
	currentReduce int
	currentMap    int
}

// start a thread that listens for RPCs from worker.go
func (c *Coordinator) server() {
	rpc.Register(c)
	rpc.HandleHTTP()
	//l, e := net.Listen("tcp", ":1234")
	sockname := coordinatorSock()
	os.Remove(sockname)
	l, e := net.Listen("unix", sockname)
	if e != nil {
		log.Fatal("listen error:", e)
	}
	go http.Serve(l, nil)
}

// main/mrcoordinator.go calls Done() periodically to find out
// if the entire job has finished.
func (c *Coordinator) Done() bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.AllDone
}

// create a Coordinator.
// main/mrcoordinator.go calls this function.
// nReduce is the number of reduce tasks to use.
func MakeCoordinator(files []string, nReduce int) *Coordinator {
	c := Coordinator{}

	// Your code here.
	c.NReduce = nReduce
	c.inputFiles = files
	c.mapTasks = *MakeTaskList()
	c.reduceTasks = *MakeTaskList()

	c.server()
	return &c
}

func (c *Coordinator) allocMapId() (ret int) {
	ret = c.currentMap
	c.currentMap++
	return
}

func (c *Coordinator) allocReduceId() (ret int) {
	ret = c.currentReduce
	c.currentReduce++
	return
}

func (c *Coordinator) RequestTask(args *RequestTaskArgs, reply *RequestTaskReply) error {
	// 这里就是你写的任务分配逻辑！
	if c.AllDone {
		reply.Task.Type = Exit
		return nil
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	var id int
	if len(c.inputFiles) > c.currentMap {
		// 分配map任务
		id = c.allocMapId()
		task := Task{
			Type:      Map,
			FileName:  c.inputFiles[id],
			ID:        id,
			NReduce:   c.NReduce,
			StartTime: time.Now(),
		}
		reply.Task = task
		c.mapTasks.Tasks.PushBack(task)
		c.mapTasks.IdMap[id] = c.mapTasks.Tasks.Back()
		return nil
	}
	// All Map work alloc done
	if c.mapTasks.Tasks.Len() == 0 {
		if c.currentReduce < c.NReduce {
			// 分配reduce任务
			id = c.allocReduceId()
			task := Task{
				Type:      Reduce,
				ID:        id,
				NReduce:   c.NReduce,
				StartTime: time.Now(),
			}
			reply.Task = task
			c.reduceTasks.Tasks.PushBack(task)
			c.reduceTasks.IdMap[id] = c.reduceTasks.Tasks.Back()
			return nil
		} else {
			// All Reduce work alloc done
			if c.reduceTasks.Tasks.Len() == 0 {
				c.AllDone = true
				reply.Task.Type = Exit
				return nil
			}
		}
	}
	//WAIT
	reply.Task.Type = Wait
	return nil
}

//Coordinator.ReportTaskDone

func (c *Coordinator) ReportTask(args *ReportTaskArgs, reply *ReportTaskReply) error {
	// 这里就是你写的任务完成上报逻辑！
	id := args.TaskID
	c.mu.Lock()
	defer c.mu.Unlock()
	switch args.Type {
	case Map:
		taskElement := c.mapTasks.IdMap[id]
		c.mapTasks.Tasks.Remove(taskElement)
	case Reduce:
		taskElement := c.reduceTasks.IdMap[id]
		c.reduceTasks.Tasks.Remove(taskElement)
	default:

	}
	return nil
}
