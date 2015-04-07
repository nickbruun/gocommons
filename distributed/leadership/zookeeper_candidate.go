package leadership

import (
	"fmt"
	log "github.com/nickbruun/gocommons/logging"
	"github.com/nickbruun/gocommons/zkutils"
	"github.com/samuel/go-zookeeper/zk"
	"path"
	"strings"
	"sync"
	"time"
)

// ZooKeeper candidate.
type zkCandidate struct {
	// ZooKeeper connection.
	c *zk.Conn

	// ZooKeeper path prefix.
	pp string

	// ZooKeeper ACL.
	acl []zk.ACL

	// Leadership handler.
	lh LeadershipHandler

	// Stop channel.
	stop chan struct{}

	// Done.
	done sync.WaitGroup
}

// Test if the candidature has been stopped.
func (c *zkCandidate) isStopped() bool {
	select {
	case <-c.stop:
		return true
	default:
		return false
	}
}

// Test if the candidature has been stopped before a certain amount of time.
func (c *zkCandidate) isStoppedBefore(d time.Duration) bool {
	select {
	case <-c.stop:
		return true
	case <-time.After(d):
		return false
	}
}

// Get candidate node path from name.
func (c *zkCandidate) candidateNodePath(name string) string {
	return fmt.Sprintf("%s/%s", c.pp, name)
}

// List all candidates.
func (c *zkCandidate) listCandidateNodes() ([]zkutils.SequenceNode, error) {
	// List the child nodes of the prefix.
	children, _, err := c.c.Children(c.pp)
	if err == zk.ErrNoNode {
		return []zkutils.SequenceNode{}, nil
	} else if err != nil {
		return nil, err
	}

	return zkutils.ParseSequenceNodes(children, "candidate"), nil
}

// Create candidate node.
func (c *zkCandidate) createCandidateNode() (n zkutils.SequenceNode, err error) {
	// Attempt to create the candidate node.
	candidatePath := fmt.Sprintf("%s/candidate", c.pp)

	var path string
	path, err = c.c.CreateProtectedEphemeralSequential(candidatePath, nil, c.acl)

	if err == nil {
		n, err = zkutils.ParseSequenceNode(path, "candidate")
		return
	} else if err == zk.ErrNoNode {
		if err = zkutils.CreateRecursively(c.c, c.pp, c.acl); err != nil {
			return
		}
		return c.createCandidateNode()
	} else {
		log.Warnf("Error creating epehermeral sequential node for candidate: %v", err)
		return
	}
}

// Safely delete a node.
func (c *zkCandidate) safelyDeleteNode(path string) {
	for {
		err := c.c.Delete(path, -1)
		if err == nil {
			log.Debugf("Removed candidate node: %s", path)
			break
		} else if err == zk.ErrNoNode {
			log.Debugf("Candidate node no longer exists: %s", path)
			break
		} else {
			log.Warnf("Failed to remove leadership candidate node, waiting 100 ms to retry: %v", err)
			time.Sleep(100 * time.Millisecond)
		}
	}
}

// Assume leadership.
func (c *zkCandidate) assumeLeadership(candidateNode zkutils.SequenceNode) (stopped bool) {
	log.Info("Became leader")

	// Invoke the leadership handler.
	end := make(chan struct{}, 1)
	resigned := make(chan struct{}, 1)

	go func() {
		c.lh(end)
		resigned <- struct{}{}
	}()

	// Watch for changes in the candidate node.
	//
	// If an error occurs, we need to assume that it's virtually impossible for
	// us to correctly assume leadership.
	removed := make(chan struct{}, 1)

	go func() {
		path := c.candidateNodePath(candidateNode.Name)

		for {
			exists, _, err := c.c.Exists(path)

			if err != nil || !exists {
				if err != nil {
					log.Debugf("Error checking for existence of candidate node %s: %v", path, err)
				} else {
					log.Debugf("Candidate node no longer exists: %s", path)
				}

				removed <- struct{}{}
			}
		}
	}()

	// Wait for the leadership handler to stop, the candidate to be stopped
	// externally or the candidate node to disappear for some reason.
	stopped = false

	select {
	case <-resigned:
		break

	case <-c.stop:
		// If we are stopped, send an end message to the leadership handler
		// and wait for the handler to resign.
		stopped = true
		end <- struct{}{}
		log.Debug("Candidate stopped, awaiting leadership handler return")
		<-resigned

	case <-removed:
		// If the candidate node is removed (or presumed removed due to a
		// timeout), send an end message to the leadership handler and wait for
		// the handler to resign.
		end <- struct{}{}
		log.Debug("Candidate node removed or check failed, awaiting leadership handler return")
		<-resigned
	}

	log.Debug("Leadership handler returned, relieving leadership role")

	// For good meassure, remove the candidate node.
	c.safelyDeleteNode(c.candidateNodePath(candidateNode.Name))

	return stopped || c.isStopped()
}

// Wait to become the leader.
func (c *zkCandidate) awaitLeadership(candidateNode zkutils.SequenceNode) (stopped bool) {
	for {
		// Test if the candidate has been stopped.
		if c.isStopped() {
			return true
		}

		// List all candidate nodes.
		candidateNodes, err := c.listCandidateNodes()
		if err != nil {
			log.Warn("Error retrieving leadership candidates, waiting 100 ms to retry: %v", err)

			if c.isStoppedBefore(100 * time.Millisecond) {
				return true
			} else {
				continue
			}
		}

		// Sort the list of candidate nodes.
		zkutils.SortSequenceNodes(candidateNodes)

		// Ensure that our candidate node is present in the list of
		// candidates.
		candidateIdx := -1

		for i, n := range candidateNodes {
			if n.SequenceNumber == candidateNode.SequenceNumber {
				candidateIdx = i
				break
			}
		}

		if candidateIdx == -1 {
			log.Warn("Candidate node has gone away")
			return false
		}

		// If the first candidate is our candidate, we have obtained
		// leadership.
		if candidateIdx == 0 {
			return c.assumeLeadership(candidateNode)
		} else {
			// Wait for the candidate node prior to our candidate to disappear.
			path := c.candidateNodePath(candidateNodes[candidateIdx-1].Name)
			log.Debugf("Waiting for change in %s", path)

			exists, _, eventChan, err := c.c.ExistsW(path)
			if !exists || err != nil {
				continue
			}

			select {
			case <-eventChan:
				break
			case <-c.stop:
				return true
			}
		}
	}
}

// Run the election process until stopped.
func (c *zkCandidate) run() {
	for {
		// Test if the candidate has been stopped.
		if c.isStopped() {
			break
		}

		// Create a candidate node.
		candidateNode, err := c.createCandidateNode()
		if err != nil {
			log.Warnf("Error creating candidate node, waiting 100 ms to retry: %v", err)
			if c.isStoppedBefore(100 * time.Millisecond) {
				break
			} else {
				continue
			}
		}

		_, candidateNode.Name = path.Split(candidateNode.Name)

		// Await leadership.
		if stopped := c.awaitLeadership(candidateNode); stopped {
			break
		}
	}

	c.done.Done()
	log.Debug("Done running candidate")
}

func (c *zkCandidate) Stop() {
	c.stop <- struct{}{}
	log.Debug("Sent stop signal")
	c.done.Wait()
	log.Debug("Done waiting for running candidate")
}

// New ZooKeeper candidate.
func NewZooKeeperCandidate(zkConn *zk.Conn, pathPrefix string, leadershipHandler LeadershipHandler) (Candidate, error) {
	if pathPrefix == "" || pathPrefix == "/" {
		return nil, fmt.Errorf("invalid path prefix: %s", pathPrefix)
	}

	zc := &zkCandidate{
		c:    zkConn,
		pp:   strings.TrimRight(pathPrefix, "/"),
		acl:  zk.WorldACL(zk.PermAll),
		lh:   leadershipHandler,
		stop: make(chan struct{}, 10),
	}

	zc.done.Add(1)
	go zc.run()

	return zc, nil
}