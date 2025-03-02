/*
 * Licensed to the Apache Software Foundation (ASF) under one or more
 * contributor license agreements.  See the NOTICE file distributed with
 * this work for additional information regarding copyright ownership.
 * The ASF licenses this file to You under the Apache License, Version 2.0
 * (the "License"); you may not use this file except in compliance with
 * the License.  You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package rpc_client

import (
	"sync"
	"sync/atomic"
	"time"
)

import (
	getty "github.com/apache/dubbo-getty"
)

var (
	MAX_CHECK_ALIVE_RETRY = 600

	CHECK_ALIVE_INTERNAL = 100

	allSessions = sync.Map{}

	// serverAddress -> rpc_client.Session -> bool
	serverSessions = sync.Map{}

	sessionSize int32 = 0

	clientSessionManager = &GettyClientSessionManager{}
)

type GettyClientSessionManager struct{}

func (sessionManager *GettyClientSessionManager) AcquireGettySession() getty.Session {
	// map 遍历是随机的
	var session getty.Session
	allSessions.Range(func(key, value interface{}) bool {
		session = key.(getty.Session)
		if session.IsClosed() {
			sessionManager.ReleaseGettySession(session)
		} else {
			return false
		}
		return true
	})
	if session != nil {
		return session
	}
	if sessionSize == 0 {
		ticker := time.NewTicker(time.Duration(CHECK_ALIVE_INTERNAL) * time.Millisecond)
		defer ticker.Stop()
		for i := 0; i < MAX_CHECK_ALIVE_RETRY; i++ {
			<-ticker.C
			allSessions.Range(func(key, value interface{}) bool {
				session = key.(getty.Session)
				if session.IsClosed() {
					sessionManager.ReleaseGettySession(session)
				} else {
					return false
				}
				return true
			})
			if session != nil {
				return session
			}
		}
	}
	return nil
}

func (sessionManager *GettyClientSessionManager) AcquireGettySessionByServerAddress(serverAddress string) getty.Session {
	m, _ := serverSessions.LoadOrStore(serverAddress, &sync.Map{})
	sMap := m.(*sync.Map)

	var session getty.Session
	sMap.Range(func(key, value interface{}) bool {
		session = key.(getty.Session)
		if session.IsClosed() {
			sessionManager.ReleaseGettySession(session)
		} else {
			return false
		}
		return true
	})
	return session
}

func (sessionManager *GettyClientSessionManager) ReleaseGettySession(session getty.Session) {
	allSessions.Delete(session)
	if !session.IsClosed() {
		m, _ := serverSessions.LoadOrStore(session.RemoteAddr(), &sync.Map{})
		sMap := m.(*sync.Map)
		sMap.Delete(session)
		session.Close()
	}
	atomic.AddInt32(&sessionSize, -1)
}

func (sessionManager *GettyClientSessionManager) RegisterGettySession(session getty.Session) {
	allSessions.Store(session, true)
	m, _ := serverSessions.LoadOrStore(session.RemoteAddr(), &sync.Map{})
	sMap := m.(*sync.Map)
	sMap.Store(session, true)
	atomic.AddInt32(&sessionSize, 1)
}
