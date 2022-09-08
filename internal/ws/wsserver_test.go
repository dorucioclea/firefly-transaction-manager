// Copyright © 2022 Kaleido, Inc.
//
// SPDX-License-Identifier: Apache-2.0
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

package ws

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"sync"
	"testing"
	"time"

	ws "github.com/gorilla/websocket"

	"github.com/stretchr/testify/assert"
)

func newTestWebSocketServer() (*webSocketServer, *httptest.Server) {
	s := NewWebSocketServer(context.Background()).(*webSocketServer)
	ts := httptest.NewServer(http.HandlerFunc(s.Handler))
	return s, ts
}

func TestConnectSendReceiveCycle(t *testing.T) {
	assert := assert.New(t)

	w, ts := newTestWebSocketServer()
	defer ts.Close()

	u, err := url.Parse(ts.URL)
	u.Scheme = "ws"
	u.Path = "/ws"
	c, _, err := ws.DefaultDialer.Dial(u.String(), nil)
	assert.NoError(err)

	c.WriteJSON(&WebSocketCommandMessage{
		Type: "listen",
	})

	s, _, r := w.GetChannels("")

	s <- "Hello World"

	var val string
	c.ReadJSON(&val)
	assert.Equal("Hello World", val)

	c.WriteJSON(&WebSocketCommandMessage{
		Type: "ignoreme",
	})

	c.WriteJSON(&WebSocketCommandMessage{
		Type: "ack",
	})
	msgOrErr := <-r
	assert.NoError(msgOrErr.Err)

	s <- "Don't Panic!"

	c.ReadJSON(&val)
	assert.Equal("Don't Panic!", val)

	c.WriteJSON(&WebSocketCommandMessage{
		Type:    "error",
		Message: "Panic!",
	})

	msgOrErr = <-r
	assert.Regexp("Error received from WebSocket client: Panic!", msgOrErr.Err)

	w.Close()

}

func TestConnectTopicIsolation(t *testing.T) {
	assert := assert.New(t)

	w, ts := newTestWebSocketServer()
	defer ts.Close()

	u, err := url.Parse(ts.URL)
	u.Scheme = "ws"
	u.Path = "/ws"
	c1, _, err := ws.DefaultDialer.Dial(u.String(), nil)
	assert.NoError(err)
	c2, _, err := ws.DefaultDialer.Dial(u.String(), nil)
	assert.NoError(err)

	c1.WriteJSON(&WebSocketCommandMessage{
		Type:  "listen",
		Topic: "topic1",
	})

	c2.WriteJSON(&WebSocketCommandMessage{
		Type:  "listen",
		Topic: "topic2",
	})

	s1, _, r1 := w.GetChannels("topic1")
	s2, _, r2 := w.GetChannels("topic2")

	s1 <- "Hello Number 1"
	s2 <- "Hello Number 2"

	var val string
	c1.ReadJSON(&val)
	assert.Equal("Hello Number 1", val)
	c1.WriteJSON(&WebSocketCommandMessage{
		Type:  "ack",
		Topic: "topic1",
	})
	msgOrErr := <-r1
	assert.NoError(msgOrErr.Err)

	c2.ReadJSON(&val)
	assert.Equal("Hello Number 2", val)
	c2.WriteJSON(&WebSocketCommandMessage{
		Type:  "ack",
		Topic: "topic2",
	})
	msgOrErr = <-r2
	assert.NoError(msgOrErr.Err)

	w.Close()

}

func TestConnectAbandonRequest(t *testing.T) {
	assert := assert.New(t)

	w, ts := newTestWebSocketServer()
	defer ts.Close()

	u, err := url.Parse(ts.URL)
	u.Scheme = "ws"
	u.Path = "/ws"
	c, _, err := ws.DefaultDialer.Dial(u.String(), nil)
	assert.NoError(err)

	c.WriteJSON(&WebSocketCommandMessage{
		Type: "listen",
	})
	_, _, r := w.GetChannels("")

	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		select {
		case <-r:
			break
		}
		wg.Done()
	}()

	// Close the client while we've got an active read stream
	c.Close()

	// We whould find the read stream closes out
	wg.Wait()
	w.Close()

}

func TestSpuriousAckProcessing(t *testing.T) {
	assert := assert.New(t)

	w, ts := newTestWebSocketServer()
	defer ts.Close()
	w.processingTimeout = 1 * time.Millisecond

	u, err := url.Parse(ts.URL)
	u.Scheme = "ws"
	u.Path = "/ws"
	c, _, err := ws.DefaultDialer.Dial(u.String(), nil)
	assert.NoError(err)

	c.WriteJSON(&WebSocketCommandMessage{
		Type:  "ack",
		Topic: "mytopic",
	})
	c.WriteJSON(&WebSocketCommandMessage{
		Type:  "ack",
		Topic: "mytopic",
	})

	for len(w.connections) > 0 {
		time.Sleep(1 * time.Millisecond)
	}
	w.Close()
}

func TestSpuriousNackProcessing(t *testing.T) {
	assert := assert.New(t)

	w, ts := newTestWebSocketServer()
	defer ts.Close()
	w.processingTimeout = 1 * time.Millisecond

	u, err := url.Parse(ts.URL)
	u.Scheme = "ws"
	u.Path = "/ws"
	c, _, err := ws.DefaultDialer.Dial(u.String(), nil)
	assert.NoError(err)

	c.WriteJSON(&WebSocketCommandMessage{
		Type:  "ack",
		Topic: "mytopic",
	})
	c.WriteJSON(&WebSocketCommandMessage{
		Type:  "error",
		Topic: "mytopic",
	})

	for len(w.connections) > 0 {
		time.Sleep(1 * time.Millisecond)
	}
	w.Close()
}

func TestConnectBadWebsocketHandshake(t *testing.T) {
	assert := assert.New(t)

	w, ts := newTestWebSocketServer()
	defer ts.Close()

	u, _ := url.Parse(ts.URL)
	u.Path = "/ws"

	res, err := http.Get(u.String())
	assert.NoError(err)
	assert.Equal(400, res.StatusCode)

	w.Close()

}

func TestBroadcast(t *testing.T) {
	assert := assert.New(t)

	w, ts := newTestWebSocketServer()
	defer ts.Close()

	u, _ := url.Parse(ts.URL)
	u.Scheme = "ws"
	u.Path = "/ws"
	topic := "banana"
	c, _, err := ws.DefaultDialer.Dial(u.String(), nil)
	assert.NoError(err)

	c.WriteJSON(&WebSocketCommandMessage{
		Type:  "listen",
		Topic: topic,
	})

	// Wait until the client has subscribed to the topic before proceeding
	for len(w.topicMap[topic]) == 0 {
		time.Sleep(10 * time.Millisecond)
	}

	_, b, _ := w.GetChannels(topic)
	b <- "Hello World"

	var val string
	c.ReadJSON(&val)
	assert.Equal("Hello World", val)

	b <- "Hello World Again"

	c.ReadJSON(&val)
	assert.Equal("Hello World Again", val)

	w.Close()
}

func TestBroadcastDefaultTopic(t *testing.T) {
	assert := assert.New(t)

	w, ts := newTestWebSocketServer()
	defer ts.Close()

	u, _ := url.Parse(ts.URL)
	u.Scheme = "ws"
	u.Path = "/ws"
	topic := ""
	c, _, err := ws.DefaultDialer.Dial(u.String(), nil)
	assert.NoError(err)

	c.WriteJSON(&WebSocketCommandMessage{
		Type: "listen",
	})

	// Wait until the client has subscribed to the topic before proceeding
	for len(w.topicMap[topic]) == 0 {
		time.Sleep(10 * time.Millisecond)
	}

	_, b, _ := w.GetChannels(topic)
	b <- "Hello World"

	var val string
	c.ReadJSON(&val)
	assert.Equal("Hello World", val)

	b <- "Hello World Again"

	c.ReadJSON(&val)
	assert.Equal("Hello World Again", val)

	w.Close()
}

func TestRecvNotOk(t *testing.T) {
	assert := assert.New(t)

	w, ts := newTestWebSocketServer()
	defer ts.Close()

	u, _ := url.Parse(ts.URL)
	u.Scheme = "ws"
	u.Path = "/ws"
	topic := ""
	c, _, err := ws.DefaultDialer.Dial(u.String(), nil)
	assert.NoError(err)

	c.WriteJSON(&WebSocketCommandMessage{
		Type: "listen",
	})

	// Wait until the client has subscribed to the topic before proceeding
	for len(w.topicMap[topic]) == 0 {
		time.Sleep(10 * time.Millisecond)
	}

	_, b, _ := w.GetChannels(topic)
	close(b)
	w.Close()
}

func TestSendReply(t *testing.T) {
	assert := assert.New(t)

	w, ts := newTestWebSocketServer()
	defer ts.Close()

	u, _ := url.Parse(ts.URL)
	u.Scheme = "ws"
	u.Path = "/ws"
	c, _, err := ws.DefaultDialer.Dial(u.String(), nil)
	assert.NoError(err)

	c.WriteJSON(&WebSocketCommandMessage{
		Type: "listenReplies",
	})

	// Wait until the client has subscribed to the topic before proceeding
	for len(w.replyMap) == 0 {
		time.Sleep(10 * time.Millisecond)
	}

	w.SendReply("Hello World")

	var val string
	c.ReadJSON(&val)
	assert.Equal("Hello World", val)
}

func TestListenTopicClosing(t *testing.T) {

	w, ts := newTestWebSocketServer()
	defer ts.Close()
	w.getTopic("test")

	c := &webSocketConnection{
		server:   w,
		topics:   make(map[string]*webSocketTopic),
		closing:  make(chan struct{}),
		newTopic: make(chan bool),
	}
	close(c.closing)
	c.listenTopic(&webSocketTopic{
		topic: "test",
	})
}

func TestBroadcastClosing(t *testing.T) {

	w, ts := newTestWebSocketServer()
	defer ts.Close()
	w.getTopic("test")

	c := &webSocketConnection{
		server:   w,
		topics:   make(map[string]*webSocketTopic),
		closing:  make(chan struct{}),
		newTopic: make(chan bool),
	}
	close(c.closing)
	// Check this doesn't block
	c.server.broadcastToConnections([]*webSocketConnection{c}, "anything")
}
