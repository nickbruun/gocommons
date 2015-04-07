package zkutils

import (
	"fmt"
	"github.com/samuel/go-zookeeper/zk"
	"os"
	"testing"
	"time"
)

// Create a test cluster of a given size.
func CreateTestCluster(t *testing.T, size int) (testCluster *zk.TestCluster, serverAddresses []string) {
	var err error

	// Create the test cluster.
	testCluster, err = zk.StartTestCluster(size, os.Stdout, os.Stdout)
	if err != nil {
		t.Fatalf("Failed to create test cluster: %v", err)
	}

	// Construct server addresses and create a connection.
	serverAddresses = make([]string, len(testCluster.Servers))

	for i, s := range testCluster.Servers {
		serverAddresses[i] = fmt.Sprintf("127.0.0.1:%d", s.Port)
	}

	return
}

// Create a test cluster of a given size and a connection to the cluster.
func CreateTestClusterAndConn(t *testing.T, size int) (testCluster *zk.TestCluster, conn *zk.Conn) {
	var serverAddrs []string
	testCluster, serverAddrs = CreateTestCluster(t, size)

	// Create a connection.
	var err error
	conn, _, err = zk.Connect(serverAddrs, 10 * time.Second)
	if err != nil {
		testCluster.Stop()
		t.Fatalf("Failed to create connection to test cluster: %v", err)
	}

	return
}
