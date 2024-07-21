package src

import (
	"fmt"
	"strings"
)

type ATTDataWords struct {
	Label     string `json:"label"`
	StartTime int    `json:"start_time"`
	EndTime   int    `json:"end_time"`
}

type ATTDataSeg struct {
	StartTime  int            `json:"start_time"`
	EndTime    int            `json:"end_time"`
	Transcript string         `json:"transcript"`
	Words      []ATTDataWords `json:"words"`
}

type ATTData struct {
	Utterances []ATTDataSeg
	Version    string
}

// toSrtTS 转换为srt时间戳
func (seg *ATTDataSeg) toSrtTS() string {
	_conv := func(ms int) (int, int, int, int) {
		h := ms / 3600000
		m := (ms / 60000) % 60
		s := (ms / 1000) % 60
		ms1 := ms % 1000
		return h, m, s, ms1
	}

	s_h, s_m, s_s, s_ms := _conv(seg.StartTime)
	e_h, e_m, e_s, e_ms := _conv(seg.EndTime)

	return fmt.Sprintf("%02d:%02d:%02d,%03d --> %02d:%02d:%02d,%03d",
		s_h, s_m, s_s, s_ms, e_h, e_m, e_s, e_ms)
}

// toLrcTS 转换为lrc时间戳
func (seg *ATTDataSeg) toLrcTS() string {
	_conv := func(ms int) (int, int, int) {
		m := ms / 60000
		s := (ms / 1000) % 60
		ms1 := (ms % 1000) / 10
		return m, s, ms1
	}

	s_m, s_s, s_ms := _conv(seg.StartTime)

	return fmt.Sprintf("[%02d:%02d.%02d]", s_m, s_s, s_ms)
}

func (a *ATTData) HasData() bool {
	return len(a.Utterances) > 0
}

// ToTxt 转成 txt 格式字幕 (无时间标记)
func (a *ATTData) ToTxt() string {
	txtLines := make([]string, len(a.Utterances))
	for i, seg := range a.Utterances {
		txtLines[i] = seg.Transcript
	}
	return "\n" + strings.Join(txtLines, "\n")
}

// ToSrt 转成 srt 格式字幕
func (a *ATTData) ToSrt() string {
	var srtLines []string
	for n, seg := range a.Utterances {
		srtLines = append(srtLines, fmt.Sprintf(
			"%d\n%s\n%s\n",
			n+1,
			seg.toSrtTS(),
			seg.Transcript,
		))
	}
	return strings.Join(srtLines, "\n")
}

// ToLrc 转成 lrc 格式字幕
func (a *ATTData) ToLrc() string {
	lrcLines := make([]string, len(a.Utterances))
	for _, seg := range a.Utterances {
		lrcLines = append(lrcLines, fmt.Sprintf(
			"%s%s",
			seg.toLrcTS(),
			seg.Transcript,
		))
	}
	return strings.Join(lrcLines, "\n")
}

// ToAss 转换为 ass 格式
// TODO: 实现 ass 序列化
func (a *ATTData) ToAss() error {
	return fmt.Errorf("NotImplementedError: ASS 序列化功能尚未实现")
}
