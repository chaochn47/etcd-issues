package main

import (
	"context"
	"fmt"
	"log"
	"runtime"
	"time"

	clientv3 "go.etcd.io/etcd/client/v3"
)

func main() {
	// 1. Establish etcd connection
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   []string{"127.0.0.1:2379"},
		DialTimeout: 5 * time.Second,
	})
	if err != nil {
		log.Fatalf("failed to connect: %v", err)
	}
	defer cli.Close()

	// 2. Pre-fill etcd with data if not already present
	// (Simulate a large response: ~10MB of data)
	ctx := context.Background()
	fmt.Println("Ensuring test data exists...")
	resp, _ := cli.Get(ctx, "bench/", clientv3.WithPrefix(), clientv3.WithCountOnly())
	if resp.Count != 1000 {
		for i := range 1000 {
			_, _ = cli.Put(ctx, fmt.Sprintf("bench/%04d", i), string(make([]byte, 10240))) // 10KB each
		}
		fmt.Println("Test data created.")
	} else {
		fmt.Printf("Test data already exists (%d keys).\n", resp.Count)
	}

	// Create 20 million nodes. This creates a massive pointer graph
	// that the GC MUST traverse, even if the total size is only ~200MB.
	fmt.Println("Applying memory pressure (ballast)...")
	ballast := createPressure(30_500_000)

	// 4. Perform Range Request under pressure
	fmt.Println("Executing Range request under GOMEMLIMIT pressure...")
	start := time.Now()
	resp, err = cli.Get(ctx, "bench/", clientv3.WithPrefix())
	duration := time.Since(start)

	if err != nil {
		log.Fatalf("Range request failed: %v", err)
	}

	fmt.Printf("Request took: %v (Received %d keys)\n", duration, len(resp.Kvs))

	// Keep ballast alive so it's not GC'd during the request
	runtime.KeepAlive(ballast)
}

type Node struct {
	Next *Node
	Data [8]byte
}

func createPressure(count int) *Node {
	var root *Node
	for range count {
		// Each node is small, but there are millions of them
		root = &Node{Next: root}
	}
	return root
}
