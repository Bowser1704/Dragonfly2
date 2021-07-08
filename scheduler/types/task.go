/*
 *     Copyright 2020 The Dragonfly Authors
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *      http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package types

import (
	"sync"
	"time"

	"d7y.io/dragonfly/v2/internal/dferrors"
	"d7y.io/dragonfly/v2/internal/rpc/base"
)

const (
	TaskStatusWaiting = iota + 1
	TaskStatusRunning
	TaskStatusFailed
	TaskStatusSuccess
	TaskStatusSourceError
)

// isSuccessCDN determines that whether the CDNStatus is success.
func IsSuccess(status int) bool {
	return status == TaskStatusSuccess
}

func IsFrozen(status int) bool {
	return status == TaskStatusFailed ||
		status == TaskStatusWaiting ||
		status == TaskStatusSourceError
}

func IsWait(status int) bool {
	return status == TaskStatusWaiting
}

func IsFailedTask(status int) bool {
	return status == TaskStatusFailed
}

type Task struct {
	taskID         string
	url            string
	filter         string
	bizID          string
	urlMeta        *base.UrlMeta
	SizeScope      base.SizeScope
	DirectPiece    []byte
	CreateTime     time.Time
	LastAccessTime time.Time
	PieceList      map[int32]*Piece
	PieceTotal     int32
	ContentLength  int64
	Statistic      *TaskStatistic
	CDNError       *dferrors.DfError
	Status         int
}

func NewTask(taskID, url, filter, bizID string, meta *base.UrlMeta) *Task {
	return &Task{
		taskID:         taskID,
		url:            url,
		filter:         filter,
		bizID:          bizID,
		urlMeta:        meta,
		CreateTime:     time.Now(),
		LastAccessTime: time.Now(),
		PieceList:      nil,
		PieceTotal:     0,
		ContentLength:  0,
		Statistic:      nil,
		CDNError:       nil,
	}
}
func (task *Task) GetTaskID() string {
	return task.taskID
}

func (task *Task) GetUrl() string {
	return task.url
}

func (task *Task) GetFilter() string {
	return task.filter
}

func (task *Task) GetUrlMeta() *base.UrlMeta {
	return task.urlMeta
}
func (task *Task) InitProps() {
	if task.PieceList == nil {
		task.CreateTime = time.Now()
		task.LastAccessTime = task.CreateTime
		task.SizeScope = base.SizeScope_NORMAL
		task.Statistic = &TaskStatistic{
			StartTime: time.Now(),
		}
	}
}

func (task *Task) GetPiece(pieceNum int32) *Piece {
	return task.PieceList[pieceNum]
}

func (task *Task) GetOrCreatePiece(pieceNum int32) *Piece {
	task.rwLock.RLock()
	p := task.PieceList[pieceNum]
	if p == nil {
		task.rwLock.RUnlock()
		p = newEmptyPiece(pieceNum, task)
		task.rwLock.Lock()
		task.PieceList[pieceNum] = p
		task.rwLock.Unlock()
	} else {
		task.rwLock.RUnlock()
	}
	return p
}

func (t *Task) AddPiece(p *Piece) {
	t.PieceList[p.PieceNum] = p
}

type TaskStatistic struct {
	lock          sync.RWMutex
	StartTime     time.Time
	EndTime       time.Time
	PeerCount     int32
	FinishedCount int32
	CostList      []int32
}

type StatisticInfo struct {
	StartTime     time.Time
	EndTime       time.Time
	PeerCount     int32
	FinishedCount int32
	Costs         map[int32]int32
}

func (t *TaskStatistic) SetStartTime(start time.Time) {
	t.lock.Lock()
	t.StartTime = start
	t.lock.Unlock()
}

func (t *TaskStatistic) SetEndTime(end time.Time) {
	t.lock.Lock()
	t.EndTime = end
	t.lock.Unlock()
}

func (t *TaskStatistic) AddPeerTaskStart() {
	t.lock.Lock()
	t.PeerCount++
	t.lock.Unlock()
}

func (t *TaskStatistic) AddPeerTaskDown(cost int32) {
	t.lock.Lock()
	t.CostList = append(t.CostList, cost)
	t.lock.Unlock()
}

func (t *TaskStatistic) GetStatistic() (info *StatisticInfo) {
	t.lock.RLock()
	defer t.lock.RUnlock()
	info = &StatisticInfo{
		StartTime:     t.StartTime,
		EndTime:       t.EndTime,
		PeerCount:     t.PeerCount,
		FinishedCount: t.FinishedCount,
		Costs:         make(map[int32]int32),
	}

	if info.EndTime.IsZero() {
		info.EndTime = time.Now()
	}

	count := len(t.CostList)
	count90 := count * 90 / 100
	count95 := count * 95 / 100

	totalCost := int64(0)

	for i, cost := range t.CostList {
		totalCost += int64(cost)
		switch i {
		case count90:
			info.Costs[90] = int32(totalCost / int64(count90))
		case count95:
			info.Costs[95] = int32(totalCost / int64(count95))
		}
	}
	if count > 0 {
		info.Costs[100] = int32(totalCost / int64(count))
	}

	return
}

type Piece struct {
	PieceNum    int32
	RangeStart  uint64
	RangeSize   int32
	PieceMd5    string
	PieceOffset uint64
	PieceStyle  PieceStyle
}

type PieceStyle int32

type PeerRegisterInfo struct {
	PeerId string
	// peer host info
	PeerHost *NodeHost
	// current host load
	HostLoad    *base.HostLoad
	IsMigrating bool
}
