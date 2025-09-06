package mr

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"
)
import "log"
import "net/rpc"
import "hash/fnv"

// Map functions return a slice of KeyValue.
type KeyValue struct {
	Key   string
	Value string
}

// use ihash(key) % NReduce to choose the reduce
// task number for each KeyValue emitted by Map.
func ihash(key string) int {
	h := fnv.New32a()
	h.Write([]byte(key))
	return int(h.Sum32() & 0x7fffffff)
}

// main/mrworker.go calls this function.
func Worker(mapf func(string, string) []KeyValue,
	reducef func(string, []string) string) {

	// Your worker implementation here.
	for {
		// 1. 请求任务
		task := RequestTaskFromCoordinator()
		switch task.Type {
		case Map:
			DoMapTask(task, mapf)
			ReportTaskDone(task)
		case Reduce:
			DoReduceTask(task, reducef)
			ReportTaskDone(task)
		case Wait:
			time.Sleep(time.Second)
		case Exit:
			return
		}
	}
	// uncomment to send the Example RPC to the coordinator.
	//CallExample()
}
func RequestTaskFromCoordinator() Task {
	args := RequestTaskArgs{}
	reply := RequestTaskReply{}
	ok := call("Coordinator.RequestTask", &args, &reply)
	if !ok {
		// Coordinator unreachable，退出
		log.Printf("Coordinator.RequestTask fail %v\n", ok)
		return Task{Type: Exit}
	}
	return reply.Task
}

// 执行 Map 任务
func DoMapTask(task Task, mapf func(string, string) []KeyValue) {
	// 读取文件、调用 mapf、写入中间文件
	content, err := os.ReadFile(task.FileName)
	if err != nil {
		log.Fatalf("cannot read %v: %v", task.FileName, err)
	}
	kva := mapf(task.FileName, string(content))
	// 将 kva 分桶写入中间文件
	buckets := make([][]KeyValue, task.NReduce)
	for _, kv := range kva {
		hash := ihash(kv.Key) % task.NReduce
		buckets[hash] = append(buckets[hash], kv)
	}
	for y, kvs := range buckets {
		intermediateFileName := fmt.Sprintf("mr-%d-%d", task.ID, y)
		//intermediateFileName := fmt.Sprintf("%s/mr-%d-%d", os.TempDir(), task.ID, y)
		f, err := os.Create(intermediateFileName)
		if err != nil {
			log.Fatalf("cannot create %v", intermediateFileName)
		}
		defer f.Close()
		enc := json.NewEncoder(f)
		for _, kv := range kvs {
			if err := enc.Encode(&kv); err != nil {
				log.Fatalf("cannot encode kv to %v: %v", intermediateFileName, err)
			}
		}
	}
}

// 执行 Reduce 任务
func DoReduceTask(task Task, reducef func(string, []string) string) {
	// 读取中间文件、调用 reducef、写入 mr-out-X
	pattern := fmt.Sprintf("mr-*-%d", task.ID)
	files, err := filepath.Glob(pattern)
	if err != nil {
		log.Fatalf("cannot find reduce files: %v", err)
	}
	kvMap := make(map[string][]string)
	for _, filename := range files {
		file, err := os.Open(filename)
		if err != nil {
			log.Fatalf("cannot open file %v: %v", filename, err)
		}
		defer file.Close()
		dec := json.NewDecoder(file)
		for {
			var kv KeyValue
			if err := dec.Decode(&kv); err != nil {
				break
			}
			kvMap[kv.Key] = append(kvMap[kv.Key], kv.Value)
		}
	}
	// sort keys
	keys := make([]string, 0, len(kvMap))
	for k := range kvMap {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	// 4. 写入临时输出文件
	outTmp := fmt.Sprintf("mr-out-%d-%d", task.ID, time.Now().UnixNano())
	outFinal := fmt.Sprintf("mr-out-%d", task.ID)
	outFile, err := os.Create(outTmp)
	if err != nil {
		log.Fatalf("cannot create output file %v: %v", outTmp, err)
	}
	defer outFile.Close()
	for _, k := range keys {
		s := reducef(k, kvMap[k])
		// 格式: key value
		fmt.Fprintf(outFile, "%v %v\n", k, s)
	}

	// 5. 原子移动到最终输出文件
	os.Rename(outTmp, outFinal)
}

// 汇报任务完成
func ReportTaskDone(task Task) {
	args := ReportTaskArgs{TaskID: task.ID, Type: task.Type}
	reply := ReportTaskReply{}
	ok := call("Coordinator.ReportTask", &args, &reply)
	if !ok {
		log.Fatalf("Coordinator.ReportTask failed")
	}
}

// send an RPC request to the coordinator, wait for the response.
// usually returns true.
// returns false if something goes wrong.
func call(rpcname string, args interface{}, reply interface{}) bool {
	// c, err := rpc.DialHTTP("tcp", "127.0.0.1"+":1234")
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
