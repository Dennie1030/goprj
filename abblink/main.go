package main

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
)

func main() {
	// ip地址 本都文件路径 上传到服务器的路径
	args := os.Args[1:] // 获取除了可执行文件路径之外的所有参数

	// 文件路径
	//filePath := `c:\aaa.txt`
	filePath := args[1]
	// 默认URL
	//baseURL := "https://192.168.125.1/fileservice/$home/"
	baseURL := "https://" + args[0] + "/fileservice/$home/"

	// 获取文件名
	fileName := args[2]

	// 构建完整URL
	fullURL := baseURL + fileName

	//fmt.Printf("fullURL: %v\n", fullURL)
	//fmt.Printf("localFile: %v\n", filePath)

	// 读取文件内容
	fileData, err := ioutil.ReadFile(filePath)
	if err != nil {
		fmt.Printf("Error reading file: %v\n", err)
		return
	}

	// 创建HTTP客户端，禁用SSL验证
	httpClient := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}

	// 创建HTTP请求
	req, err := http.NewRequest("PUT", fullURL, bytes.NewReader(fileData))
	if err != nil {
		fmt.Printf("Error creating request: %v\n", err)
		return
	}

	// 设置请求头
	req.Header.Set("Accept", "application/hal+json; v=2.0")
	req.Header.Set("Content-Type", "application/octet-stream; v=2.0")

	// 设置基本认证
	req.SetBasicAuth("Default User", "robotics")

	// 发送请求
	resp, err := httpClient.Do(req)
	if err != nil {
		fmt.Printf("Error uploading file: %v\n", err)
		return
	}
	defer resp.Body.Close()

	// 检查响应状态
	if resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusCreated {
		//fmt.Printf("File '%s' uploaded successfully to '%s'\n", fileName, fullURL)
		fmt.Printf("ok")
	} else {
		fmt.Printf("Failed to upload file. Status code: %d\n", resp.StatusCode)
	}
}
