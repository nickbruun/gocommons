package zkutils

import (
	"github.com/samuel/go-zookeeper/zk"
	log "github.com/nickbruun/logging"
)

// Await the existence of a node.
//
// Emits a nil object, or an error, on the channel, when the node at the given
// path exists or an error occurs.
func AwaitExists(conn *zk.Conn, path string) <- chan error {
	await := make(chan error, 1)

	go func() {
		for {
			exists, _, event, err := conn.ExistsW(path)

			if err != nil {
				log.Debugf("Error testing for existence of %s: %v", path, err)
				await <- err
				return
			} else if exists {
				log.Debugf("Node at path %s exists", path)
				await <- nil
				return
			} else {
				log.Debugf("Node %s does not exist, awaiting event", path)
				<- event
			}
		}
	}()

	return await
}
