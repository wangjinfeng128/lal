// Copyright 2020, Chef.  All rights reserved.
// https://github.com/q191201771/lal
//
// Use of this source code is governed by a MIT-style license
// that can be found in the License file.
//
// Author: Chef (191201771@qq.com)

package main

import (
	"os"

	"github.com/q191201771/lal/pkg/rtprtcp"

	"github.com/q191201771/lal/pkg/base"

	"github.com/q191201771/lal/pkg/aac"

	"github.com/q191201771/lal/pkg/rtsp"
	"github.com/q191201771/naza/pkg/nazalog"
)

type Obs struct {
}

var obs Obs
var aacFp *os.File
var avcFp *os.File
var a aac.ADTS

func (obs *Obs) OnRTPPacket(pkt rtprtcp.RTPPacket) {

}

func (obs *Obs) OnAVConfig(asc, vps, sps, pps []byte) {
	if asc != nil {
		_ = a.InitWithAACAudioSpecificConfig(asc)
	}
	if vps != nil {
		_, _ = avcFp.Write([]byte{0, 0, 0, 1})
		_, _ = avcFp.Write(vps)
	}
	if sps != nil && pps != nil {
		_, _ = avcFp.Write([]byte{0, 0, 0, 1})
		_, _ = avcFp.Write(sps)
		_, _ = avcFp.Write([]byte{0, 0, 0, 1})
		_, _ = avcFp.Write(pps)
	}
}

func (obs *Obs) OnAVPacket(pkt base.AVPacket) {
	nazalog.Debugf("type=%d, ts=%d, len=%d", pkt.PayloadType, pkt.Timestamp, len(pkt.Payload))

	switch pkt.PayloadType {
	case base.RTPPacketTypeAVCOrHEVC:
		// TODO chef: 由于存在多nalu情况，需要进行拆分
		_, _ = avcFp.Write([]byte{0, 0, 0, 1})
		_, _ = avcFp.Write(pkt.Payload)
		_ = avcFp.Sync()
	case base.RTPPacketTypeAAC:
		h, _ := a.CalcADTSHeader(uint16(len(pkt.Payload)))
		_, _ = aacFp.Write(h)
		_, _ = aacFp.Write(pkt.Payload)
		_ = aacFp.Sync()
	}
}

func (obs *Obs) OnNewRTSPPubSession(session *rtsp.PubSession) bool {
	nazalog.Debugf("OnNewRTSPPubSession. %+v", session)

	var err error
	avcFp, err = os.Create("/tmp/rtsp.h264")
	nazalog.Assert(nil, err)

	aacFp, err = os.Create("/tmp/rtsp.aac")
	nazalog.Assert(nil, err)

	session.SetObserver(obs)

	return true
}

func (obs *Obs) OnDelRTSPPubSession(session *rtsp.PubSession) {
	nazalog.Debugf("OnDelRTSPPubSession. %+v", session)
}

func (obs *Obs) OnNewRTSPSubSession(session *rtsp.SubSession) bool {
	return true
}

func (obs *Obs) OnDelRTSPSubSession(session *rtsp.SubSession) {

}

func main() {
	var obs Obs
	s := rtsp.NewServer(":5544", &obs)
	err := s.Listen()
	nazalog.Assert(nil, err)
	err = s.RunLoop()
	nazalog.Error(err)
}
