#!/bin/bash

# Lab1 Build and Test Script
# 用于构建和测试 MapReduce 实现

echo "🔨 Building MapReduce components..."

# Navigate to the correct directory
cd "$(dirname "$0")/src/main"

# Clean up any old binaries
rm -f mrcoordinator mrworker mrsequential

# Build the MapReduce components
echo "Building coordinator..."
go build -race mrcoordinator.go
if [ $? -ne 0 ]; then
    echo "❌ Failed to build coordinator"
    exit 1
fi

echo "Building worker..."
go build -race mrworker.go
if [ $? -ne 0 ]; then
    echo "❌ Failed to build worker"
    exit 1
fi

echo "Building sequential version..."
go build -race mrsequential.go
if [ $? -ne 0 ]; then
    echo "❌ Failed to build sequential version"
    exit 1
fi

# Build the plugins
echo "Building plugins..."
cd ../mrapps
for plugin in wc.go indexer.go mtiming.go rtiming.go jobcount.go early_exit.go crash.go nocrash.go; do
    if [ -f "$plugin" ]; then
        echo "Building $plugin..."
        go build -race -buildmode=plugin "$plugin"
        if [ $? -ne 0 ]; then
            echo "❌ Failed to build $plugin"
            exit 1
        fi
    fi
done

cd ../main

echo "✅ Build completed successfully!"

# Test if sequential version works
echo ""
echo "🧪 Testing sequential version..."
./mrsequential ../mrapps/wc.so pg-*.txt
if [ $? -eq 0 ] && [ -f "mr-out-0" ]; then
    echo "✅ Sequential version works!"
    word_count=$(head -1 mr-out-0)
    echo "First line of output: $word_count"
    rm -f mr-out-*
else
    echo "❌ Sequential version failed"
    exit 1
fi

echo ""
echo "🚀 Ready to start implementing your distributed MapReduce!"
echo ""
echo "Next steps:"
echo "1. Implement the RPC definitions in src/mr/rpc.go"
echo "2. Implement the coordinator logic in src/mr/coordinator.go"  
echo "3. Implement the worker logic in src/mr/worker.go"
echo "4. Test with: bash test-mr.sh"
echo ""
echo "Templates are provided in the templates/ directory:"
echo "- templates/rpc_example.go"
echo "- templates/coordinator_template.go"
echo "- templates/worker_template.go"
echo ""
echo "Copy and adapt these templates to implement your solution!"