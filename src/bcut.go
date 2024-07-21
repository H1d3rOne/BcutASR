package src

import (
	"io/ioutil"
	"log"
	"net/http"
	"path/filepath"
	"strings"
)

type Bcut struct {
	Client     *http.Client
	InputPath  string
	InputFmt   string
	OutputPath string
	OutputFmt  string
	Format     string
	Interval   int
}

func NewBcut() *Bcut {
	return &Bcut{}
}

func (bcut *Bcut) SetData() []byte {
	//处理输入文件
	var inputData []byte
	if bcut.InputPath != "" {
		inputExt := filepath.Ext(bcut.InputPath)
		if bcut.InputFmt == "" {
			bcut.InputFmt = strings.TrimPrefix(inputExt, ".")
		}
		//InputFileName := filepath.Base(bcut.InputPath)
		InputPathNoExt := strings.TrimSuffix(bcut.InputPath, inputExt)

		switch bcut.InputFmt {
		case "flac", "aac", "m4a", "mp3", "wav":
			var err error
			// 直接读取文件
			inputData, err = ioutil.ReadFile(bcut.InputPath)
			if err != nil {
				log.Println("读取输入文件出错")
			}
		case "mp4", "avi", "mov", "flv":
			err := extractAudio(bcut.InputPath, InputPathNoExt+".mp3")
			if err != nil {
				log.Fatalf("提取音频失败：%v", err)
			}
			inputData, err = ioutil.ReadFile(InputPathNoExt + ".mp3")
			if err != nil {
				log.Println("读取输入文件出错")
			}
		default:
			var err error
			log.Println("非标准音频文件, 尝试调用ffmpeg转码")
			inputData, err = ffmpegRender(bcut.InputPath)
			if err != nil {
				log.Fatalf("ffmpeg转码失败或为非音频文件: %v", err)
			}
			log.Println("ffmpeg转码完成")
			bcut.InputFmt = "acc"
		}
		//fmt.Println(inputData)
	} else {
		log.Fatalln("需要填写输入文件路径")
	}

	//处理输出文件
	//bcut.OutputPath = determineOutputPath(bcut.InputPath, bcut.OutputPath, bcut.DeafaultFmt)
	//bcut.OutputFmt = getPathFmt(bcut.OutputPath)
	return inputData
}
