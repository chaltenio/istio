// Copyright Istio Authors
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

package ctrlz

import (
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"testing"
	"time"
)

func TestStartStopEnabled(t *testing.T) {
	server := startAndWaitForServer(t)
	done := make(chan struct{}, 1)
	go func() {
		server.Close()
		done <- struct{}{}
	}()
	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("Timed out waiting for listeningTestProbe to be called")
	}
}

func TestSignals(t *testing.T) {
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGUSR1)
	server := startAndWaitForServer(t)
	defer server.Close()
	reloadURL := fmt.Sprintf("http://%v/signalj/SIGUSR1", server.Address())
	resp, err := http.DefaultClient.Post(reloadURL, "text/plain", nil)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if resp.StatusCode != http.StatusAccepted {
		t.Fatalf("Got unexpected status code: %v", resp.StatusCode)
	}
	select {
	case <-c:
	case <-time.After(5 * time.Second):
		t.Fatal("Timed out waiting for SIGUSR1")
	}
}

func startAndWaitForServer(t *testing.T) *Server {
	ready := make(chan struct{}, 1)
	listeningTestProbe = func() {
		ready <- struct{}{}
	}
	defer func() { listeningTestProbe = nil }()

	// Start and wait for server
	o := DefaultOptions()
	s, err := Run(o, nil)
	if err != nil {
		t.Fatalf("Failed to start server: %v", err)
	}
	select {
	case <-ready:
	case <-time.After(5 * time.Second):
		t.Fatal("Timed out waiting for server start")
	}
	return s
}
