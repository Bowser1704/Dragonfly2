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

package service

import (
	"context"
	"fmt"
	"io"
	"time"

	"d7y.io/dragonfly/v2/internal/dfcodes"
	"d7y.io/dragonfly/v2/internal/dferrors"
	logger "d7y.io/dragonfly/v2/internal/dflog"
	"d7y.io/dragonfly/v2/internal/rpc/base"
	"d7y.io/dragonfly/v2/internal/rpc/base/common"
	"d7y.io/dragonfly/v2/internal/rpc/scheduler"
	"d7y.io/dragonfly/v2/pkg/util/net/urlutils"
	"d7y.io/dragonfly/v2/pkg/util/stringutils"
	"d7y.io/dragonfly/v2/scheduler/config"
	"d7y.io/dragonfly/v2/scheduler/core"
	"d7y.io/dragonfly/v2/scheduler/daemon"
	"d7y.io/dragonfly/v2/scheduler/types"
	ants "github.com/panjf2000/ants/v2"
	"github.com/pkg/errors"
)

type SchedulerServer struct {
	service     *core.SchedulerService
	worker      *ants.Pool
	config      config.SchedulerConfig
	peerManager daemon.PeerMgr
	taskManager daemon.TaskMgr
}

// Option is a functional option for configuring the scheduler
type Option func(p *SchedulerServer) *SchedulerServer

// WithSchedulerService sets the *service.SchedulerService
func WithSchedulerService(service *core.SchedulerService) Option {
	return func(s *SchedulerServer) *SchedulerServer {
		s.service = service

		return s
	}
}


// NewSchedulerServer returns a new transparent scheduler server from the given options
func NewSchedulerServer(cfg *config.Config, options ...Option) *SchedulerServer {
	return NewSchedulerWithOptions(cfg, options...)
}

// NewSchedulerWithOptions constructs a new instance of a scheduler server with additional options.
func NewSchedulerWithOptions(cfg *config.Config, options ...Option) *SchedulerServer {
	scheduler := &SchedulerServer{
		config: cfg.Scheduler,
	}

	for _, opt := range options {
		opt(scheduler)
	}

	return scheduler
}

func (s *SchedulerServer) RegisterPeerTask(ctx context.Context, request *scheduler.PeerTaskRequest) (resp *scheduler.RegisterResult, err error) {
	if err := validateParams(request); err != nil {
		// todo return
	}
	resp = new(scheduler.RegisterResult)

	// get or create task
	//var isCdn = false
	taskID := s.service.GenerateTaskID(request.Url, request.Filter, request.UrlMeta, request.BizId, request.PeerId)
	task, ok := s.service.GetTask(taskID)
	if !ok {
		task, err = s.service.AddTask(types.NewTask(resp.TaskId, request.Url, request.Filter, request.BizId, request.UrlMeta))
		if err != nil {
			dferror, _ := err.(*dferrors.DfError)
			if dferror != nil && dferror.Code == dfcodes.SchedNeedBackSource {
				isCdn = true
			} else {
				return
			}
		}
	}

	if task.CDNError != nil {
		err = task.CDNError
		return
	}

	// get or create host
	reqPeerHost := request.PeerHost
	if host, ok := s.service.GetHost(reqPeerHost.Uuid); !ok {
		host = &types.NodeHost{
			Type: types.NodeHost,
			PeerHost: scheduler.PeerHost{
				Uuid:           reqPeerHost.Uuid,
				Ip:             reqPeerHost.Ip,
				RpcPort:        reqPeerHost.RpcPort,
				DownPort:       reqPeerHost.DownPort,
				HostName:       reqPeerHost.HostName,
				SecurityDomain: reqPeerHost.SecurityDomain,
				Location:       reqPeerHost.Location,
				Idc:            reqPeerHost.Idc,
				NetTopology:    reqPeerHost.NetTopology,
			},
		}
		//if isCdn {
		//	host.Type = types.HostTypeCdn
		//}
		host, err = s.service.AddHost(host)
		if err != nil {
			return
		}
	}

	resp.TaskId = task.GetTaskID()
	resp.SizeScope = task.SizeScope

	// case base.SizeScope_TINY
	if resp.SizeScope == base.SizeScope_TINY {
		resp.DirectPiece = task.DirectPiece
		return
	}

	// get or creat PeerTask
	if peerTask, ok := s.service.GetPeerTask(request.PeerId); !ok {
		peerTask, err = s.service.AddPeerTask(pid, task, host)
		if err != nil
	} else if peerTask.Host == nil {
		peerTask.Host = host
	}

	if isCdn {
		peerTask.SetDown()
		err = dferrors.New(dfcodes.SchedNeedBackSource, "there is no cdn")
		return
	} else if peerTask.IsDown() {
		peerTask.SetUp()
	}

	if resp.SizeScope == base.SizeScope_NORMAL {
		return
	}

	// case base.SizeScope_SMALL
	// do scheduler piece
	parent, _, err := s.service.ScheduleParent(peerTask)
	if err != nil {
		return
	}

	if parent == nil {
		resp.SizeScope = base.SizeScope_NORMAL
		return
	}

	resp.DirectPiece = &scheduler.RegisterResult_SinglePiece{
		SinglePiece: &scheduler.SinglePiece{
			// destination peer id
			DstPid: parent.Pid,
			// download address(ip:port)
			DstAddr: fmt.Sprintf("%s:%d", parent.Host.Ip, parent.Host.DownPort),
			// one piece task
			PieceInfo: &task.PieceList[0].PieceInfo,
		},
	}

	return
}

func (s *SchedulerServer) ReportPieceResult(stream scheduler.Scheduler_ReportPieceResultServer) error {
	for {
		select {
		case <-stream.Context().Done():
			logger.Infof()
		default:
			pieceResult, err := stream.Recv()
			if err == io.EOF || pieceResult.PieceNum == common.EndOfPiece {
				logger.Infof("read all piece result")
				return nil
			}
			if err != nil {
				// 处理piece error
				return err
			}
		}
	}
	//err = worker.NewClient(stream, s.worker, s.service).Serve()
	//return
}

func (s *SchedulerServer) ReportPeerResult(ctx context.Context, result *scheduler.PeerResult) (err error) {
	peerTask, err := s.service.GetPeerTask(result.PeerId)
	if err != nil {
		return
	}
	peerTask.SetStatus(result.Traffic, result.Cost, result.Success, result.Code)

	if peerTask.Success {
		peerTask.Status = types.PeerStatusDone
		s.worker.ReceiveJob(peerTask)
	} else {
		peerTask.Status = types.PeerStatusLeaveNode
		s.worker.ReceiveJob(peerTask)
	}

	return
}

func (s *SchedulerServer) LeaveTask(ctx context.Context, target *scheduler.PeerTarget) error {
	startTime := time.Now()
	defer func() {
		e := recover()
		if e != nil {
			err = dferrors.New(dfcodes.SchedError, fmt.Sprintf("%v", e))
			return
		}
		if err != nil {
			if _, ok := err.(*dferrors.DfError); !ok {
				err = dferrors.New(dfcodes.SchedError, err.Error())
			}
		}
		logger.Debugf("ReportPeerResult [%s] cost time: [%d]", target.PeerId, time.Now().Sub(startTime))
		return
	}()

	peerNode, ok := s.service.GetPeerTask(target.PeerId)
	if !ok {
		return nil
	}

	peerNode.Status = types.PeerStatusLeaveNode
	s.worker.ReceiveJob(peerNode)

	return
}

// validateParams validates the params of scheduler.PeerTaskRequest.
func validateParams(req *scheduler.PeerTaskRequest) error {
	if !urlutils.IsValidURL(req.Url) {
		return errors.Wrapf(errortypes.ErrInvalidValue, "raw url: %s", req.Url)
	}

	if stringutils.IsEmpty(req.PeerId) {
		return errors.Wrapf(errortypes.ErrEmptyValue, "path")
	}
	return nil
}
