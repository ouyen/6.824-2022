# Lab1 MapReduce 入门指南

欢迎来到 MIT 6.824 分布式系统课程的 Lab1！这个实验要求你实现一个分布式的 MapReduce 系统。

## 📁 项目结构

```
6.824-2022/
├── LAB1_GETTING_STARTED.md      # 本指南
├── build_and_test.sh            # 构建和测试脚本
├── templates/                   # 实现模板
│   ├── rpc_example.go           # RPC 定义示例
│   ├── coordinator_template.go  # Coordinator 实现模板
│   └── worker_template.go       # Worker 实现模板
├── src/
│   ├── mr/                      # 你需要修改的核心文件
│   │   ├── coordinator.go       # 待实现：协调器
│   │   ├── worker.go           # 待实现：工作进程
│   │   └── rpc.go              # 待实现：RPC 定义
│   ├── main/                    # 主程序和测试
│   │   ├── mrcoordinator.go     # Coordinator 启动程序
│   │   ├── mrworker.go         # Worker 启动程序
│   │   ├── mrsequential.go     # 顺序版本（参考）
│   │   └── test-mr.sh          # 测试脚本
│   └── mrapps/                  # MapReduce 应用
│       ├── wc.go               # 词频统计
│       └── ...                 # 其他测试应用
```

## 🎯 实验目标

实现一个分布式的 MapReduce 系统，让多个 worker 进程并行处理大型数据集，就像 Google 的 MapReduce 论文中描述的那样。

## 🚀 快速开始

1. **运行构建脚本**：
```bash
./build_and_test.sh
```

2. **查看模板文件**：
参考 `templates/` 目录中的实现模板来理解需要实现的组件。

3. **开始实现**：
从 `src/mr/rpc.go` 开始，定义 RPC 接口，然后实现 coordinator 和 worker。

## 📋 核心组件

你需要实现以下三个主要组件：

### 1. Coordinator (协调器) - `src/mr/coordinator.go`
- 分配 map 和 reduce 任务给 workers
- 跟踪任务完成状态
- 处理 worker 故障
- 决定何时开始 reduce 阶段

### 2. Worker (工作进程) - `src/mr/worker.go`
- 从 coordinator 请求任务
- 执行 map 或 reduce 函数
- 读写中间文件
- 报告任务完成状态

### 3. RPC 通信 - `src/mr/rpc.go`
- 定义 coordinator 和 worker 之间的消息格式
- 任务请求和分配
- 任务完成通知

## 🚀 从哪里开始

### 第一步：理解顺序版本
```bash
cd src/main
go build mrsequential.go
./mrsequential ../mrapps/wc.so pg-*.txt
head mr-out-0
```

这个命令运行单机版的 MapReduce，处理所有的文本文件并生成词频统计。研究 `mrsequential.go` 的代码来理解：
- Map 和 Reduce 函数如何工作
- 键值对的数据结构
- 输出文件的格式

### 第二步：理解测试
```bash
cd src/main
bash test-mr.sh
```

现在这个测试会失败，因为分布式版本还没有实现。但是看看测试脚本来理解：
- 如何启动 coordinator 和 workers
- 期望的输出格式
- 不同的测试场景

### 第三步：设计 RPC 接口

在 `src/mr/rpc.go` 中添加你的 RPC 定义。考虑这些消息类型：

```go
// Worker 请求任务
type TaskRequest struct {
    WorkerID string  // 可选：worker 标识
}

type TaskReply struct {
    TaskType   string    // "map", "reduce", "wait", "done"
    TaskID     int       // 任务编号
    Filename   string    // Map 任务的输入文件
    NReduce    int       // Reduce 任务总数
    NMap       int       // Map 任务总数
}

// Worker 报告任务完成
type TaskCompleteRequest struct {
    TaskType string
    TaskID   int
}

type TaskCompleteReply struct {
    // 可以为空
}
```

### 第四步：实现基本的 Coordinator

在 `src/mr/coordinator.go` 中：

1. **数据结构设计**：
```go
type Coordinator struct {
    mu          sync.Mutex
    files       []string        // 输入文件列表
    nReduce     int            // reduce 任务数
    mapTasks    []Task         // map 任务状态
    reduceTasks []Task         // reduce 任务状态
    phase      string         // "map", "reduce", "done"
}

type Task struct {
    ID        int
    State     string    // "idle", "in-progress", "completed"
    Filename  string    // 对于 map 任务
    StartTime time.Time // 用于超时检测
}
```

2. **RPC 处理函数**：
```go
func (c *Coordinator) RequestTask(args *TaskRequest, reply *TaskReply) error {
    c.mu.Lock()
    defer c.mu.Unlock()
    
    // 找到一个空闲的任务并分配给 worker
    // 更新任务状态为 "in-progress"
    // 记录开始时间
}

func (c *Coordinator) CompleteTask(args *TaskCompleteRequest, reply *TaskCompleteReply) error {
    c.mu.Lock()
    defer c.mu.Unlock()
    
    // 标记任务为完成
    // 检查是否可以进入下一阶段
}
```

### 第五步：实现基本的 Worker

在 `src/mr/worker.go` 中：

```go
func Worker(mapf func(string, string) []KeyValue, reducef func(string, []string) string) {
    for {
        // 1. 请求任务
        task := requestTask()
        
        if task.TaskType == "done" {
            break
        } else if task.TaskType == "wait" {
            time.Sleep(time.Second)
            continue
        } else if task.TaskType == "map" {
            // 2. 执行 map 任务
            doMapTask(task, mapf)
        } else if task.TaskType == "reduce" {
            // 3. 执行 reduce 任务
            doReduceTask(task, reducef)
        }
        
        // 4. 报告任务完成
        reportTaskComplete(task)
    }
}
```

### 第六步：处理中间文件

**Map 任务**需要：
1. 读取输入文件
2. 调用 map 函数
3. 将结果按 reduce 任务分区
4. 写入中间文件：`mr-X-Y`（X=map任务号，Y=reduce任务号）

```go
func doMapTask(task TaskReply, mapf func(string, string) []KeyValue) {
    // 读取文件
    content := readFile(task.Filename)
    
    // 调用 map 函数
    kva := mapf(task.Filename, content)
    
    // 按 reduce 任务分区
    buckets := make([][]KeyValue, task.NReduce)
    for _, kv := range kva {
        bucket := ihash(kv.Key) % task.NReduce
        buckets[bucket] = append(buckets[bucket], kv)
    }
    
    // 写入中间文件
    for i := 0; i < task.NReduce; i++ {
        filename := fmt.Sprintf("mr-%d-%d", task.TaskID, i)
        writeIntermediateFile(filename, buckets[i])
    }
}
```

**Reduce 任务**需要：
1. 读取所有相关的中间文件：`mr-*-Y`（Y=reduce任务号）
2. 合并和排序键值对
3. 对每个唯一键调用 reduce 函数
4. 写入最终输出文件：`mr-out-Y`

## 🔧 实现技巧

### 1. 文件操作
```go
// 安全地写入文件（原子操作）
tempFile, _ := ioutil.TempFile("", "mr-tmp-*")
// 写入数据到 tempFile
tempFile.Close()
os.Rename(tempFile.Name(), finalFilename)
```

### 2. JSON 序列化中间文件
```go
import "encoding/json"

// 写入
enc := json.NewEncoder(file)
for _, kv := range kvs {
    enc.Encode(&kv)
}

// 读取
dec := json.NewDecoder(file)
for {
    var kv KeyValue
    if err := dec.Decode(&kv); err != nil {
        break
    }
    kva = append(kva, kv)
}
```

### 3. 处理并发和锁
- 在 Coordinator 中使用互斥锁保护共享状态
- 考虑任务超时和重新分配
- 处理 worker 崩溃的情况

## 🧪 测试和调试

### 运行测试
```bash
cd src/main
bash test-mr.sh
```

### 构建和运行
```bash
# 构建
cd src/main
go build -race mrcoordinator.go
go build -race mrworker.go
cd ../mrapps
go build -buildmode=plugin wc.go

# 运行
cd ../main
./mrcoordinator pg-*.txt &
./mrworker ../mrapps/wc.so &
./mrworker ../mrapps/wc.so &
./mrworker ../mrapps/wc.so &
```

### 调试技巧
1. 添加日志输出来跟踪任务分配和完成
2. 检查中间文件是否正确生成
3. 验证最终输出与顺序版本匹配
4. 使用 `go run -race` 检测竞争条件

## 📚 实现步骤总结

1. **第一阶段**：基本功能
   - [ ] 定义 RPC 接口
   - [ ] 实现基本的 coordinator 任务分配
   - [ ] 实现基本的 worker 任务执行
   - [ ] 处理中间文件生成和读取

2. **第二阶段**：并行性
   - [ ] 确保多个 workers 可以并行工作
   - [ ] 正确处理 map 和 reduce 阶段的转换
   - [ ] 通过并行性测试

3. **第三阶段**：容错性
   - [ ] 处理 worker 故障和任务超时
   - [ ] 实现任务重新分配
   - [ ] 通过崩溃测试

## 🎯 成功标准

当你的实现通过所有测试时，你会看到：
```
*** PASSED ALL TESTS
```

这意味着你的 MapReduce 实现能够：
- 正确处理词频统计和索引任务
- 支持多个 workers 并行执行
- 在 workers 崩溃时保持健壮性
- 产生与顺序版本相同的输出

祝你在 Lab1 中好运！记住，分布式系统编程需要仔细思考并发、通信和故障处理。慢慢来，一步一步实现，经常测试！