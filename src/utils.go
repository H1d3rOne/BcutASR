package src

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// 上传请求
func httpRequest(client *http.Client, method string, targetUrl string, body io.Reader, headers map[string]string) *http.Response {

	request, err := http.NewRequest(method, targetUrl, body)
	if err != nil {
		panic(err)
	}
	for k, v := range headers {
		request.Header.Set(k, v)
	}

	response, err := client.Do(request)
	if err != nil {
		panic(err)

	}
	//defer response.Body.Close()
	return response
}

// 提取视频伴音并转码为aac格式
func ffmpegRender(mediaFile string) ([]byte, error) {
	cmd := exec.Command(
		"ffmpeg",
		"-v", "warning",
		"-i", mediaFile,
		"-ac", "1",
		"-f", "adts",
		"-",
	)

	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		return nil, fmt.Errorf("ffmpeg command failed: %w", err)
	}
	return out.Bytes(), nil
}

// 确定输出文件路径
func DetermineOutputPath(inputPath, outputPath, defaultFormat string) string {
	if outputPath != "" {
		return outputPath
	}
	baseName := strings.TrimSuffix(filepath.Base(inputPath), filepath.Ext(inputPath))
	return fmt.Sprintf("%s.%s", baseName, defaultFormat)
}

// 获取路径的文件格式
func GetPathFmt(path string) string {
	fileName := filepath.Base(path)
	ext := filepath.Ext(fileName)
	fileFmt := strings.TrimPrefix(ext, ".")
	return fileFmt
}

// 获取项目根路径
func getRoot() string {
	dir, err := os.Getwd()
	if err != nil {
		log.Fatalf("获取当前工作目录失败：%s", err)
	}
	return dir
}

// 获取路径是绝对路径还是相对路径，或者是文件名
func getFileType(path string) string {
	var fileType string
	if filepath.IsAbs(path) {
		fmt.Printf("'%s' 是绝对路径。\n", path)
		fileType = "ABS"
	} else {
		// 检查是否包含路径分隔符来区分相对路径和文件名
		if strings.ContainsAny(path, string(filepath.Separator)) {
			fmt.Printf("'%s' 是相对路径。\n", path)
			fileType = "REL"
		} else {
			fmt.Printf("'%s' 更倾向于是文件名。\n", path)
			fileType = "FILE"
		}
	}
	return fileType
}

// 从视频中提取音频
func extractAudio(videoPath string, outputPath string) error {
	// 构建FFmpeg命令
	cmd := exec.Command("ffmpeg",
		"-i", videoPath, // 输入视频文件
		"-vn",                // 不输出视频流
		"-c:a", "libmp3lame", // 复制音频流，如果需要转码可以替换为aac等编码器
		"-y",
		outputPath, // 输出音频文件路径
	)

	// 获取命令的标准输出和错误输出，用于日志或错误处理
	var stderr, stdout bytes.Buffer
	cmd.Stderr = &stderr
	cmd.Stdout = &stdout

	// 执行命令
	err := cmd.Run()
	if err != nil {
		log.Printf("执行FFmpeg命令时出错: %v\n", err)
		log.Printf("Stderr: %s\n", stderr.String())
		return err
	}

	log.Printf("提取音频完成，输出信息:\n%s\n", stdout.String())
	return nil
}
