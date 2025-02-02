// Copyright 2019, Chef.  All rights reserved.
// https://github.com/q191201771/lal
//
// Use of this source code is governed by a MIT-style license
// that can be found in the License file.
//
// Author: Chef (191201771@qq.com)

package rtmp

import (
	"github.com/q191201771/lal/pkg/base"
	"github.com/q191201771/naza/pkg/connection"
)

type OnReadRTMPAVMsg func(msg base.RTMPMsg)

type PullSession struct {
	core *ClientSession
}

type PullSessionOption struct {
	ConnectTimeoutMS int
	PullTimeoutMS    int
	ReadAVTimeoutMS  int
}

var defaultPullSessionOption = PullSessionOption{
	ConnectTimeoutMS: 0,
	PullTimeoutMS:    0,
	ReadAVTimeoutMS:  0,
}

type ModPullSessionOption func(option *PullSessionOption)

func NewPullSession(modOptions ...ModPullSessionOption) *PullSession {
	opt := defaultPullSessionOption
	for _, fn := range modOptions {
		fn(&opt)
	}

	return &PullSession{
		core: NewClientSession(CSTPullSession, func(option *ClientSessionOption) {
			option.ConnectTimeoutMS = opt.ConnectTimeoutMS
			option.DoTimeoutMS = opt.PullTimeoutMS
			option.ReadAVTimeoutMS = opt.ReadAVTimeoutMS
		}),
	}
}

// 建立rtmp play连接
// 阻塞直到收到服务端返回的rtmp publish对应结果的信令，或发生错误
//
// @param onReadRTMPAVMsg: 注意，回调结束后，内存块会被PullSession重复使用
func (s *PullSession) Pull(rawURL string, onReadRTMPAVMsg OnReadRTMPAVMsg) error {
	s.core.onReadRTMPAVMsg = onReadRTMPAVMsg
	return s.core.doWithTimeout(rawURL)
}

func (s *PullSession) Done() <-chan error {
	return s.core.Done()
}

func (s *PullSession) Dispose() {
	s.core.Dispose()
}

func (s *PullSession) UniqueKey() string {
	return s.core.UniqueKey
}

func (s *PullSession) GetStat() base.StatSession {
	return s.core.GetStat()
}

// TODO chef: 默认每5秒调用一次
func (s *PullSession) UpdateStat(interval uint32) {
	s.core.UpdateStat(interval)
}

func (s *PullSession) IsAlive(interval uint32) (ret bool) {
	currStat := s.core.conn.GetStat()
	if s.core.staleStat == nil {
		s.core.staleStat = new(connection.Stat)
		*s.core.staleStat = currStat
		return true
	}

	ret = !(currStat.ReadBytesSum-s.core.staleStat.ReadBytesSum == 0)
	*s.core.staleStat = currStat
	return ret
}
