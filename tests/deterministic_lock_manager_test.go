package main

import (
	"github.com/cbthchbc/determisticExecution/common"
	"github.com/cbthchbc/determisticExecution/proto_"
	"github.com/cbthchbc/determisticExecution/scheduler"
	"sync"
	"testing"
	"time"
)

func ThroughputTest(t *testing.T) {
	var readyTxns []*proto_.TxnProto
	var config common.Configuration
	lm := scheduler.NewDeterministicLockManager(&readyTxns, &config)

	var txns []*proto_.TxnProto

	for i := 0; i < 100000; i++ {
		txns = append(txns, &proto_.TxnProto{})
	}

	start := time.Now()

	next := 0
	var wg sync.WaitGroup
	for i := 0; i < 1000; i++ {
		for j := 0; j < 100; j++ {
			wg.Add(1)
			go func(next int) {
				defer wg.Done()
				lm.Lock(txns[next])
			}(next)
			next++
		}

		wg.Wait()

		for len(readyTxns) > 0 {
			txn := readyTxns[0]
			readyTxns = readyTxns[1:]
			lm.Release(txn)
		}
	}

	//elapsed := time.Since(start)
	//txnsPerSec := float64(100000) / elapsed.Seconds()
	//fmt.Printf("%.2f txns/sec\n", txnsPerSec)
}

func TestThroughputTest(t *testing.T) {
	ThroughputTest(t)
}

func main() {
	// Run the test
	testing.Main(func(pat, str string) (bool, error) { return true, nil }, nil, nil, nil)
}
