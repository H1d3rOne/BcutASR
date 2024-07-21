package asr

import (
	"crypto/tls"
	"encoding/json"
	"github.com/H1d3rOne/BcutASR/src"
	"log"
	"net/http"
	"os"
)

var bcut src.Bcut

//func main() {
//	src.RunCmd()
//}

func NewClient() *BcutClient {
	return &BcutClient{}
}

func (client BcutClient) Input(inputFile string) {
	client.InputFile = inputFile
	if bcut.InputPath == "" {
		bcut.InputPath = client.InputFile
	}
}

func (client BcutClient) Format(format string) {
	client.OutputFormat = format
	if bcut.Format == "" {
		bcut.Format = client.OutputFormat
	}
}

func (client BcutClient) Output(outputFile string) {
	client.OutputFile = outputFile
	if bcut.OutputPath == "" {
		bcut.OutputPath = client.OutputFile
	}
}

func (client BcutClient) Interval(interval int) {
	client.IntervalTime = interval
	if bcut.Interval == 0 {
		bcut.Interval = client.IntervalTime
	}
}

func (client BcutClient) Run() {
	if bcut.InputPath == "" {
		log.Println("请输入文件")
	}
	if bcut.Format == "" {
		bcut.Format = "srt"
	}
	if bcut.Interval == 0 {
		bcut.Interval = 300
	}

	//开始识别音频
	customTransport := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	bcut.Client = &http.Client{Transport: customTransport}
	bcut.OutputPath = src.DetermineOutputPath(bcut.InputPath, bcut.OutputPath, bcut.Format)
	bcut.OutputFmt = src.GetPathFmt(bcut.OutputPath)

	//处理输入文件
	inputData := bcut.SetData()

	//发起请求,获取解析结果
	resultResp := src.ParseBcut(bcut.Client, inputData, bcut.Interval)

	//转换结果保存到输出文件
	output, err := os.Create(bcut.OutputPath)
	if err != nil {
		panic(err)
	}
	//output.Write()
	defer output.Close()
	attData := src.ATTData{}
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
