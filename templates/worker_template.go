package mr

import (
	"encoding/json"
	"fmt"
	"hash/fnv"
	"io/ioutil"
	"log"
	"net/rpc"
	"os"
	"sort"
	"time"
)

//
// Worker implementation template for Lab1
//

//
// Map functions return a slice of KeyValue.
//
type KeyValue struct {
	Key   string
	Value string
}

//
// use ihash(key) % NReduce to choose the reduce
// task number for each KeyValue emitted by Map.
//
func ihash(key string) int {
	h := fnv.New32a()
	h.Write([]byte(key))
	return int(h.Sum32() & 0x7fffffff)
}

//
// main/mrworker.go calls this function.
//
func Worker(mapf func(string, string) []KeyValue, reducef func(string, []string) string) {
	// TODO: Implement worker logic
	// 1. Request tasks from coordinator in a loop
	// 2. Execute map or reduce tasks as assigned
	// 3. Report completion back to coordinator
	// 4. Exit when told the job is done

	for {
		// Request a task
		task := requestTask()

		if task.TaskType == "done" {
			break
		} else if task.TaskType == "wait" {
			time.Sleep(time.Second)
			continue
		} else if task.TaskType == "map" {
			doMapTask(task, mapf)
		} else if task.TaskType == "reduce" {
			doReduceTask(task, reducef)
		}

		// Report task completion
		reportTaskComplete(task)
	}
}

//
// Request a task from the coordinator
//
func requestTask() TaskReply {
	args := TaskRequest{}
	reply := TaskReply{}

	ok := call("Coordinator.RequestTask", &args, &reply)
	if !ok {
		fmt.Printf("call failed!\n")
		// In a real system, we might want to retry or exit
		reply.TaskType = "done"
	}
	return reply
}

//
// Report task completion to coordinator
//
func reportTaskComplete(task TaskReply) {
	args := TaskCompleteRequest{
		TaskType: task.TaskType,
		TaskID:   task.TaskID,
	}
	reply := TaskCompleteReply{}

	ok := call("Coordinator.CompleteTask", &args, &reply)
	if !ok {
		fmt.Printf("call failed!\n")
	}
}

//
// Execute a map task
//
func doMapTask(task TaskReply, mapf func(string, string) []KeyValue) {
	// TODO: Implement map task execution
	// 1. Read the input file
	// 2. Call the map function
	// 3. Partition the output by reduce task (using ihash)
	// 4. Write intermediate files mr-X-Y where X=map task, Y=reduce task

	log.Printf("Executing map task %d on file %s", task.TaskID, task.Filename)

	// Read input file
	content, err := ioutil.ReadFile(task.Filename)
	if err != nil {
		log.Fatalf("cannot read %v", task.Filename)
	}

	// Call map function
	kva := mapf(task.Filename, string(content))

	// Partition by reduce task
	buckets := make([][]KeyValue, task.NReduce)
	for _, kv := range kva {
		bucket := ihash(kv.Key) % task.NReduce
		buckets[bucket] = append(buckets[bucket], kv)
	}

	// Write intermediate files
	for i := 0; i < task.NReduce; i++ {
		filename := fmt.Sprintf("mr-%d-%d", task.TaskID, i)
		writeIntermediateFile(filename, buckets[i])
	}
}

//
// Execute a reduce task
//
func doReduceTask(task TaskReply, reducef func(string, []string) string) {
	// TODO: Implement reduce task execution
	// 1. Read all intermediate files mr-*-Y where Y=reduce task number
	// 2. Sort by key
	// 3. For each unique key, call reduce function with all values
	// 4. Write final output to mr-out-Y

	log.Printf("Executing reduce task %d", task.TaskID)

	// Read all intermediate files for this reduce task
	var kva []KeyValue
	for i := 0; i < task.NMap; i++ {
		filename := fmt.Sprintf("mr-%d-%d", i, task.TaskID)
		kvs := readIntermediateFile(filename)
		kva = append(kva, kvs...)
	}

	// Sort by key
	sort.Sort(ByKey(kva))

	// Create output file
	outputFile := fmt.Sprintf("mr-out-%d", task.TaskID)
	ofile, _ := os.Create(outputFile)
	defer ofile.Close()

	// Group by key and call reduce function
	i := 0
	for i < len(kva) {
		j := i + 1
		for j < len(kva) && kva[j].Key == kva[i].Key {
			j++
		}
		values := []string{}
		for k := i; k < j; k++ {
			values = append(values, kva[k].Value)
		}
		output := reducef(kva[i].Key, values)

		// Write to output file
		fmt.Fprintf(ofile, "%v %v\n", kva[i].Key, output)

		i = j
	}
}

//
// Write intermediate key-value pairs to a file using JSON encoding
//
func writeIntermediateFile(filename string, kvs []KeyValue) {
	// Use temporary file for atomic writes
	tempFile, err := ioutil.TempFile("", "mr-tmp-*")
	if err != nil {
		log.Fatalf("cannot create temp file")
	}

	enc := json.NewEncoder(tempFile)
	for _, kv := range kvs {
		err := enc.Encode(&kv)
		if err != nil {
			log.Fatalf("cannot encode %v", kv)
		}
	}
	tempFile.Close()

	// Atomically rename to final filename
	os.Rename(tempFile.Name(), filename)
}

//
// Read intermediate key-value pairs from a file using JSON encoding
//
func readIntermediateFile(filename string) []KeyValue {
	var kva []KeyValue

	file, err := os.Open(filename)
	if err != nil {
		// File might not exist if map task hasn't created it yet
		return kva
	}
	defer file.Close()

	dec := json.NewDecoder(file)
	for {
		var kv KeyValue
		if err := dec.Decode(&kv); err != nil {
			break
		}
		kva = append(kva, kv)
	}
	return kva
}

//
// for sorting by key.
//
type ByKey []KeyValue

func (a ByKey) Len() int           { return len(a) }
func (a ByKey) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByKey) Less(i, j int) bool { return a[i].Key < a[j].Key }

//
// send an RPC request to the coordinator, wait for the response.
// usually returns true.
// returns false if something goes wrong.
//
func call(rpcname string, args interface{}, reply interface{}) bool {
	sockname := coordinatorSock()
	c, err := rpc.DialHTTP("unix", sockname)
	if err != nil {
		log.Fatal("dialing:", err)
	}
	defer c.Close()

	err = c.Call(rpcname, args, reply)
	if err == nil {
		return true
	}

	fmt.Println(err)
	return false
}