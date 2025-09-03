# Lab1 Implementation Templates

This directory contains template files to help you get started with implementing the MapReduce system for Lab1.

## Files

### rpc_example.go
- Shows how to define RPC structures for communication between coordinator and workers
- Includes examples for task requests, task assignments, and task completion messages
- Copy the relevant parts to `../src/mr/rpc.go`

### coordinator_template.go  
- Template implementation for the MapReduce coordinator
- Shows data structures for tracking tasks and workers
- Includes skeleton RPC handlers for task assignment and completion
- Copy and adapt this to `../src/mr/coordinator.go`

### worker_template.go
- Template implementation for MapReduce workers
- Shows how to request tasks, execute map/reduce operations, and handle intermediate files
- Includes file I/O utilities for reading/writing data
- Copy and adapt this to `../src/mr/worker.go`

## How to Use

1. Read through each template to understand the overall structure
2. Copy the relevant code to the actual implementation files in `../src/mr/`
3. Fill in the TODO sections with your own logic
4. Remove the template files once you're done (they're just for reference)

## Important Notes

- The templates are not complete implementations - you need to fill in the core logic
- Pay attention to the TODOs and comments for guidance
- Make sure to handle edge cases like worker failures and task timeouts
- Test your implementation frequently using the test script

Good luck with Lab1!