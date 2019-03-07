// Copyright 2015 The etcd Authors
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

package main

import (
	"bytes"  // to create bytes buffers for communication
	"encoding/gob"	// to encode and decode structs into byte buffers
	"encoding/json"  // to encode and decode jsons
	"log"
	"sync"  // to use a reader/writer mutex

	"github.com/coreos/etcd/snap" // handles raft nodes states with snapshots
)

// a Petri Net node store backed by raft
type pnstore struct {
	proposeC    chan<- string // channel for proposing updates
	mu          sync.RWMutex
	pnStore     map[string]petriNodeInfo // current committed id-petriNodeInfo pairs
	snapshotter *snap.Snapshotter
}

type PetriNodeInfo struct {
	ip string
	leader bool
	available bool
}

type pn struct {
	Key string
	Val petriNodeInfo
}

func newPNStore(snapshotter *snap.Snapshotter, proposeC chan<- string, commitC <-chan *string, errorC <-chan error) *pnstore {
	s := &pnstore{proposeC: proposeC, kvStore: make(map[string]PetriNodeInfo), snapshotter: snapshotter}
	// replay log into key-value map
	s.readCommits(commitC, errorC)
	// read commits from raft into kvStore map until error
	go s.readCommits(commitC, errorC)
	return s
}

func (s *pnstore) Lookup(key string) (PetriNodeInfo, bool) {
	s.mu.RLock()
	v, ok := s.pnStore[key]
	s.mu.RUnlock()
	return v, ok
}

func (s *pnstore) Propose(id string, nodeIp string, isLeader bool, isAvailable bool) {
	var buf bytes.Buffer
	if err := gob.NewEncoder(&buf).Encode(pn{id, PetriNodeInfo{nodeIp, isLeader, isAvailable}}); err != nil {
		log.Fatal(err)
	}
	s.proposeC <- buf.String()
}

func (s *pnstore) readCommits(commitC <-chan *string, errorC <-chan error) {
	for data := range commitC {
		if data == nil {
			// done replaying log; new data incoming
			// OR signaled to load snapshot
			snapshot, err := s.snapshotter.Load()
			if err == snap.ErrNoSnapshot {
				return
			} else if err != nil {
				log.Panic(err)
			}
			log.Printf("loading snapshot at term %d and index %d", snapshot.Metadata.Term, snapshot.Metadata.Index)
			if err := s.recoverFromSnapshot(snapshot.Data); err != nil {
				log.Panic(err)
			}
			continue
		}

		var dataPn pn
		dec := gob.NewDecoder(bytes.NewBufferString(*data))
		if err := dec.Decode(&dataPn); err != nil {
			log.Fatalf("raftexample: could not decode message (%v)", err)
		}
		s.mu.Lock()
		s.pnStore[dataPn.Key] = dataPn.Val
		s.mu.Unlock()
	}
	if err, ok := <-errorC; ok {
		log.Fatal(err)
	}
}

func (s *pnstore) getSnapshot() ([]byte, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return json.Marshal(s.pnStore)
}

func (s *kvstore) recoverFromSnapshot(snapshot []byte) error {
	var store map[string]PetriNodeInfo
	if err := json.Unmarshal(snapshot, &store); err != nil {
		return err
	}
	s.mu.Lock()
	s.pnStore = store
	s.mu.Unlock()
	return nil
}
