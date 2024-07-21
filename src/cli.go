package src

import (
	"crypto/tls"
	"encoding/json"
	"flag"
	"log"
	"net/http"
	"os"
)

func RunCmd() {
	var (
		//inputFmt = ""
		inputFile  = flag.String("i", "", "输入文件路径")
		format     = flag.String("f", "srt", "输出字幕格式")
		outputFile = flag.String("o", "", "输出文件路径")
		interval   = flag.Int("t", 300, "请求发送时间间隔，单位为毫秒")
	)

	flag.Parse()

	//开始识别音频
	bcut := NewBcut()
	customTransport := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	bcut.Client = &http.Client{Transport: customTransport}

	if *inputFile == "" {
		log.Fatalln("请输入文件")
	} else {
		bcut.InputPath = *inputFile
	}

	bcut.OutputPath = DetermineOutputPath(*inputFile, *outputFile, *format)
	bcut.OutputFmt = GetPathFmt(bcut.OutputPath)
	//bcut.Format = *format
	bcut.Interval = *interval
	//处理输入文件
	inputData := bcut.SetData()

	//发起请求,获取解析结果
	resultResp := ParseBcut(bcut.Client, inputData, bcut.Interval)

	//转换结果保存到输出文件
	output, err := os.Create(bcut.OutputPath)
	if err != nil {
		panic(err)
	}
	//output.Write()
	defer output.Close()
	attData := ATTData{}
	err = json.Unmarshal([]byte(resultResp.Result), &attData)
	if err != nil {
		log.Fatal(err)
	}
	if attData.HasData() {
		switch bcut.OutputFmt {
		case "srt":
			_, err := output.WriteString(attData.ToSrt())
			if err != nil {
				log.Fatalf("保存到srt文件失败：%s", err)
			}
		case "lrc":
			_, err := output.WriteString(attData.ToLrc())
			if err != nil {
				log.Fatalf("保存到lrc文件失败：%s", err)
			}
		case "text":
			_, err := output.WriteString(attData.ToTxt())
			if err != nil {
				log.Fatalf("保存到text文件失败：%s", err)
			}
		}
		log.Printf("保存到文件：%s", bcut.OutputPath)
	} else {
		log.Fatal("未识别出语音")
	}
}
