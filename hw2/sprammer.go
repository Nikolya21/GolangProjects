package main

import (
	"fmt"
	"sort"
	"sync"
)

func RunPipeline(cmds ...cmd) {
	if len(cmds) == 0 {
		return
	}

	chans := make([]chan interface{}, len(cmds)+1)
	for i := range chans {
		chans[i] = make(chan interface{})
	}

	var wg sync.WaitGroup
	for i := 0; i < len(cmds); i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			cmds[idx](chans[idx], chans[idx+1])
			close(chans[idx+1])
		}(i)
	}

	wg.Wait()
}

func SelectUsers(in, out chan interface{}) {
	setUsers := make(map[string]bool)
	var mu sync.Mutex
	var wg sync.WaitGroup

	for u := range in {
		email := u.(string)
		wg.Add(1)
		go func(emailStr string) {
			defer wg.Done()
			user := GetUser(emailStr)

			mu.Lock()
			if setUsers[user.Email] {
				mu.Unlock()
				return
			}
			setUsers[user.Email] = true
			mu.Unlock()

			out <- user
		}(email)
	}
	wg.Wait()
}

func SelectMessages(in, out chan interface{}) {
	var wg sync.WaitGroup
	batch := make([]User, 0, GetMessagesMaxUsersBatch)

	for u := range in {
		batch = append(batch, u.(User))
		if len(batch) == GetMessagesMaxUsersBatch {
			wg.Add(1)
			go func(usersToProcess []User) {
				defer wg.Done()
				res, err := GetMessages(usersToProcess...)
				if err != nil {
					return
				}
				for _, v := range res {
					out <- v
				}
			}(batch)
			batch = make([]User, 0, GetMessagesMaxUsersBatch)
		}
	}
	if len(batch) > 0 {
		wg.Add(1)
		go func(usersToProcess []User) {
			defer wg.Done()
			res, err := GetMessages(usersToProcess...)
			if err != nil {
				return
			}
			for _, v := range res {
				out <- v
			}
		}(batch)
	}
	wg.Wait()
}

func CheckSpam(in, out chan interface{}) {
	semaphora := make(chan struct{}, HasSpamMaxAsyncRequests)
	var wg sync.WaitGroup

	for u := range in {
		msgID := u.(MsgID)
		wg.Add(1)
		go func(id MsgID) {
			defer wg.Done()
			semaphora <- struct{}{}
			defer func() { <-semaphora }()

			res, err := HasSpam(id)
			if err != nil {
				return
			}
			out <- MsgData{
				ID:      id,
				HasSpam: res,
			}
		}(msgID)
	}
	wg.Wait()
}

func CombineResults(in, out chan interface{}) {
	var msgData []MsgData
	for u := range in {
		msgData = append(msgData, u.(MsgData))
	}
	sort.Slice(msgData, func(i, j int) bool {
		if msgData[i].HasSpam != msgData[j].HasSpam {
			return msgData[i].HasSpam
		}
		return msgData[i].ID < msgData[j].ID
	})
	for _, v := range msgData {
		out <- fmt.Sprintf("%t %d", v.HasSpam, v.ID)
	}
}
