// Copyright 2021 The casbin Authors. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package discuzx

import (
	"fmt"
	"sync"
	"testing"

	"github.com/casbin/casnode/controllers"
	"github.com/casbin/casnode/object"
)

var AddThreadsConcurrency = 20

func TestAddThreads(t *testing.T) {
	object.InitConfig()
	InitAdapter()
	object.InitAdapter()
	controllers.InitAuthConfig()

	attachmentMap := getAttachmentMap()
	fmt.Printf("Loaded attachments: %d\n", len(attachmentMap))
	forumMap := getForumMap()
	fmt.Printf("Loaded forums: %d\n", len(forumMap))
	classMap := getClassMap()
	fmt.Printf("Loaded classes: %d\n", len(classMap))
	threads := getThreads()
	fmt.Printf("Loaded threads: %d\n", len(threads))
	threadPostsMap, postCount := getThreadPostsMap()
	fmt.Printf("Loaded posts: %d\n", postCount)

	arrayMutex := sync.RWMutex{}

	var wg sync.WaitGroup
	wg.Add(len(threads))

	sem := make(chan int, SyncAvatarsConcurrency)
	topics := []*object.Topic{}
	replies := []*object.Reply{}
	for i, thread := range threads {
		sem <- 1
		go func(i int, thread *Thread) {
			defer wg.Done()

			attachments := attachmentMap[thread.Tid]
			forum := forumMap[thread.Fid]
			topic, replies2 := addThread(thread, threadPostsMap, attachments, forum, classMap)
			arrayMutex.Lock()
			topics = append(topics, topic)
			replies = append(replies, replies2...)
			arrayMutex.Unlock()
			fmt.Printf("[%d/%d]: Added thread: tid = %d, fid = %d, replies = %d\n", i+1, len(threads), thread.Tid, thread.Fid, len(replies2))
			<-sem
		}(i, thread)
	}

	wg.Wait()

	object.AddTopicsInBatch(topics)
	object.AddRepliesInBatch(replies)
}
