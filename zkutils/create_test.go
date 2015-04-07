package zkutils

import (
	"github.com/samuel/go-zookeeper/zk"
	"testing"
	"time"
)

func AssertCreateRecursivelyCreates(t *testing.T, conn *zk.Conn, path string) {
	if err := CreateRecursively(conn, path, zk.WorldACL(zk.PermAll)); err != nil {
		t.Errorf("Failed to create path %s recursively: %v", path, err)
		return
	}

	exists, _, err := conn.Exists(path)
	if err != nil {
		t.Errorf("Failed to test if path %s exists: %v", path, err)
		return
	}

	if !exists {
		t.Errorf("Expected path %s to exist", path)
	}
}

func TestCreateRecursively(t *testing.T) {
	testCluster, serverAddrs := CreateTestCluster(t, 1)
	defer testCluster.Stop()

	conn, _, err := zk.Connect(serverAddrs, 10 * time.Second)
	if err != nil {
		t.Fatalf("Failed to create ZooKeeper connection to test cluster")
	}
	defer conn.Close()

	AssertCreateRecursivelyCreates(t, conn, "/")
	AssertCreateRecursivelyCreates(t, conn, "/one")
	AssertCreateRecursivelyCreates(t, conn, "/one")
	AssertCreateRecursivelyCreates(t, conn, "/two/three/four/five/six")
	AssertCreateRecursivelyCreates(t, conn, "/two/three/four/five/six")
}
