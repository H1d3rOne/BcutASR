package src

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"
)

//var client *http.Client

const (
	baseUrl           = "https://member.bilibili.com/x/bcut/rubick-interface"
	uploadReqUrl      = baseUrl + "/resource/create"
	uploadCompleteUrl = baseUrl + "/resource/create/complete"
	uploadTaskUrl     = baseUrl + "/task"
)

//func init() {
//	customTransport := &http.Transport{
//		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
//	}
//	client = &http.Client{Transport: customTransport}
//
//}

func ParseBcut(clent *http.Client, inputData []byte, interval int) ResultResp {
	//1、读取音频文件
	//data, err := ioutil.ReadFile("audio.mp3")
	//data, err := ioutil.ReadFile(filename)
	//if err != nil {
	//	panic(err)
	//}
	//1、发起请求上传
	log.Println("开始解析语音")
	form := url.Values{}
	form.Add("model_id", "8")
	form.Add("name", "audio.mp3")
	form.Add("resource_file_type", "mp3")
	form.Add("size", fmt.Sprintf("%d", len(inputData)))
	form.Add("type", "2")
	body := strings.NewReader(form.Encode())
	headers := map[string]string{
		"Content-Type": "application/x-www-form-urlencoded",
		"User-Agent":   "cpp-httplib/0.9",
	}
	resp := httpRequest(clent, "POST", uploadReqUrl, body, headers)
	bodyText, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("读取响应体错误:%v", err)
	}
	//fmt.Println(string(bodyText))
	var respJson map[string]interface{}
	err = json.Unmarshal(bodyText, &respJson)
	if err != nil {
		panic(err)
	}
	// 解析响应内容
	inBossKey := respJson["data"].(map[string]interface{})["in_boss_key"].(string)
	uploadId := respJson["data"].(map[string]interface{})["upload_id"].(string)
	resourceID := respJson["data"].(map[string]interface{})["resource_id"].(string)
	uploadURL := respJson["data"].(map[string]interface{})["upload_urls"].([]interface{})[0].(string)
	//关闭
	resp.Body.Close()

	//2、上传文件
	headers = map[string]string{
		//"Content-Type": "application/x-www-form-urlencoded",
		"User-Agent": "cpp-httplib/0.9",
	}
	bodyBytes := bytes.NewReader(inputData)
	resp = httpRequest(clent, "PUT", uploadURL, bodyBytes, headers)
	etag := resp.Header.Get("Etag")
	resp.Body.Close()

	//3、完成上传
	form = url.Values{}
	form.Set("etags", etag)
	form.Add("in_boss_key", inBossKey)
	form.Add("model_id", "8")
	form.Add("resource_id", resourceID)
	form.Add("upload_id", uploadId)
	body = strings.NewReader(form.Encode())
	headers = map[string]string{
		"Content-Type": "application/x-www-form-urlencoded",
		"User-Agent":   "cpp-httplib/0.9",
	}

	resp = httpRequest(clent, "POST", uploadCompleteUrl, body, headers)
	headers = map[string]string{
		"User-Agent": "cpp-httplib/0.9",
	}
	bodyText, err = io.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("读取响应体错误:%v", err)
	}
	//fmt.Println(string(bodyText))

	err = json.Unmarshal(bodyText, &respJson)
	if err != nil {
		panic(err)
	}
	downloadURL := respJson["data"].(map[string]interface{})["download_url"].(string)
	resp.Body.Close()

	//4、创建任务
	type TaskData struct {
		Resource string `json:"resource"`
		ModelID  string `json:"model_id"`
	}
	taskData := TaskData{
		Resource: downloadURL,
		ModelID:  "8",
	}
	jsonTaskData, err := json.Marshal(taskData)
	if err != nil {
		panic(err)
	}
	bodyBytes = bytes.NewReader(jsonTaskData)
	resp = httpRequest(clent, "POST", uploadTaskUrl, bodyBytes, headers)
	bodyText, err = io.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}
	//fmt.Println(string(bodyText))

	err = json.Unmarshal(bodyText, &respJson)
	if err != nil {
		panic(err)
	}
	taskId := respJson["data"].(map[string]interface{})["task_id"].(string)
	resp.Body.Close()

	//5、获取结果
	var result ResultResp
	//fmt.Println("开始获取结果")

loop:
	for {
		resultUrl := "https://member.bilibili.com/x/bcut/rubick-interface/task/result?model_id=8&task_id=" + taskId
		headers = map[string]string{
			"User-Agent": "cpp-httplib/0.9",
		}
		resp = httpRequest(clent, "GET", resultUrl, nil, headers)
		defer resp.Body.Close()
		bodyText, err = io.ReadAll(resp.Body)
		if err != nil {
			panic(err)
		}

		result = resultWrap(bodyText)
		//fmt.Println(result.State)
		switch result.State {
		case 0:
			log.Println("等待识别开始")
		case 1:
			log.Println("识别中...")
		case 3:
			log.Println("识别失败")
		case 4:
			log.Println("识别成功")
			//fmt.Println(result)
			break loop
		}
		time.Sleep(time.Duration(interval) * time.Millisecond)
	}
	log.Println("语音解析完成")
	return result
}

//type State float64
//
//const (
//	STOP     State = 0 // 未开始
//	RUNING         = 1 // 运行中
//	ERROR          = 3 // 错误
//	COMPLETE       = 4 // 完成
//)

type ResultResp struct {
	TaskID string  `json:"task_id"` // 任务id
	Result string  `json:"result"`  // 结果数据-json
	Remark string  `json:"remark"`  // 任务状态详情
	State  float64 `json:"state"`   // 任务状态
}

func resultWrap(respBody []byte) ResultResp {
	resultResp := ResultResp{}
	var respJson map[string]interface{}
	err := json.Unmarshal(respBody, &respJson)
	if err != nil {
		panic(err)
	}
	//fmt.Println(string(respBody))
	//fmt.Println(respJson)
	resultResp.TaskID = respJson["data"].(map[string]interface{})["task_id"].(string)
	resultResp.Result = respJson["data"].(map[string]interface{})["result"].(string)
	resultResp.Remark = respJson["data"].(map[string]interface{})["remark"].(string)
	resultResp.State = respJson["data"].(map[string]interface{})["state"].(float64)

	return resultResp
}
