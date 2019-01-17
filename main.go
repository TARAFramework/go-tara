// Copyright 2017,2018 Lei Ni (nilei81@gmail.com).
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

/*
multigroup is an example program for dragonboat demonstrating how multiple
raft groups can be used in an user application.
*/
package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"
	"time"

	"github.com/lni/dragonboat"
	"github.com/lni/dragonboat-example/utils"
	"github.com/lni/dragonboat/config"
	"github.com/lni/dragonboat/logger"
)

//TODO: Cluster ID Array to facilitate dynamic cluster creation
const (
	// we use two raft groups in this example, they are identified by the cluster
	// ID values below
	clusterID1 uint64 = 100
	clusterID2 uint64 = 101
)

//TODO: NodeAddress Array to facilitate dynamic changes to cluster's membership
var (
	// initial nodes count is three, their addresses are also fixed
	// this is for simplicity
	addresses = []string{
		"localhost:63001",
		"localhost:63002",
		"localhost:63003",
	}
)

func main() {
	nodeID := flag.Int("nodeid", 1, "NodeID to use")
	flag.Parse()
	if *nodeID > 3 || *nodeID < 1 {
		fmt.Fprintf(os.Stderr, "invalid nodeid %d, it must be 1, 2 or 3", *nodeID)
		os.Exit(1)
	} //TODO: Eliminate this condition checking to allow more than 3 nodes in a raft cluster
	// https://github.com/golang/go/issues/17393
	if runtime.GOOS == "darwin" {
		signal.Ignore(syscall.Signal(0xd))
	}
	peers := make(map[uint64]string)
	for idx, v := range addresses {
		// key is the NodeID, NodeID is not allowed to be 0
		// value is the raft address
		peers[uint64(idx+1)] = v
	}
	nodeAddr := peers[uint64(*nodeID)]
	fmt.Fprintf(os.Stdout, "node address: %s\n", nodeAddr)
	// change the log verbosity
	logger.GetLogger("raft").SetLevel(logger.ERROR)
	logger.GetLogger("rsm").SetLevel(logger.WARNING)
	logger.GetLogger("transport").SetLevel(logger.WARNING)
	logger.GetLogger("grpc").SetLevel(logger.WARNING)
	// config for raft
	// note the ClusterID value is not specified here
	//TODO: Allow customization of the numerical attributes through setters
	rc := config.Config{
		NodeID:             uint64(*nodeID),
		ElectionRTT:        5,
		HeartbeatRTT:       1,
		CheckQuorum:        true,
		SnapshotEntries:    10,
		CompactionOverhead: 5,
	}
	//TODO: Migrate the folder to goenv.HOME/.multigroup-data
	datadir := filepath.Join(
		"example-data",
		"multigroup-data",
		fmt.Sprintf("node%d", *nodeID))
	// config for the nodehost
	// by default, insecure transport is used, you can choose to use Mutual TLS
	// Authentication to authenticate both servers and clients. To use Mutual
	// TLS Authentication, set the MutualTLS field in NodeHostConfig to true, set
	// the CAFile, CertFile and KeyFile fields to point to the path of your CA
	// file, certificate and key files.
	// by default, TCP based RPC module is used, set the RaftRPCFactory field in
	// NodeHostConfig to rpc.NewRaftGRPC (github.com/lni/dragonboat/plugin/rpc) to
	// use gRPC based transport. To use gRPC based RPC module, you need to install
	// the gRPC library first -
	//
	// $ go get -u google.golang.org/grpc
	//
	nhc := config.NodeHostConfig{
		WALDir:         datadir,
		NodeHostDir:    datadir,
		RTTMillisecond: 200,
		RaftAddress:    nodeAddr,
		// RaftRPCFactory: rpc.NewRaftGRPC,
	}
	// create a NodeHost instance. it is a facade interface allowing access to
	// all functionalities provided by dragonboat.
	nh := dragonboat.NewNodeHost(nhc)
	defer nh.Stop()
	// start the first cluster
	// we use ExampleStateMachine as the IStateMachine for this cluster, its
	// behaviour is identical to the one used in the Hello World example.
	//TODO: Automate the creation of clusters. Tagged to to-dos at L#40 & L#48
	rc.ClusterID = clusterID1
	if err := nh.StartCluster(peers, false, NewExampleStateMachine, rc); err != nil {
		fmt.Fprintf(os.Stderr, "failed to add cluster, %v\n", err)
		os.Exit(1)
	}
	// start the second cluster
	// we use SecondStateMachine as the IStateMachine for the second cluster
	//TODO: Automate the creation of clusters. Tagged to to-dos at L#40 & L#48
	rc.ClusterID = clusterID2
	if err := nh.StartCluster(peers, false, NewSecondStateMachine, rc); err != nil {
		fmt.Fprintf(os.Stderr, "failed to add cluster, %v\n", err)
		os.Exit(1)
	}
	raftStopper := utils.NewStopper()
	consoleStopper := utils.NewStopper()
	ch := make(chan string, 16)
	consoleStopper.RunWorker(func() {
		reader := bufio.NewReader(os.Stdin)
		for {
			s, err := reader.ReadString('\n')
			if err != nil {
				close(ch)
				return
			}
			if s == "exit\n" {
				raftStopper.Stop()
				// no data will be lost/corrupted if nodehost.Stop() is not called
				nh.Stop()
				return
			}
			ch <- s
		}
	})
	raftStopper.RunWorker(func() {
		// use NO-OP client session here
		// check the example in godoc to see how to use a regular client session
		cs1 := nh.GetNoOPSession(clusterID1)
		cs2 := nh.GetNoOPSession(clusterID2)
		for {
			select {
			case v, ok := <-ch:
				if !ok {
					return
				}
				ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
				// remove the \n char
				msg := strings.Replace(strings.TrimSpace(v), "\n", "", 1)
				var err error
				//TODO: Redefine the criteria to redirect the client requests with new mechanisms (?)
				if strings.HasSuffix(msg, "?") {
					// user message ends with "?", make a proposal to update the second
					// raft group
					_, err = nh.SyncPropose(ctx, cs2, []byte(msg))
				} else {
					// message not ends with "?", make a proposal to update the first
					// raft group
					_, err = nh.SyncPropose(ctx, cs1, []byte(msg))
				}
				cancel()
				if err != nil {
					fmt.Fprintf(os.Stderr, "SyncPropose returned error %v\n", err)
				}
			case <-raftStopper.ShouldStop():
				return
			}
		}
	})
	raftStopper.Wait()
}
