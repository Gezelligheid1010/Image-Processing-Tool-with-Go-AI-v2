package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

type ImageRequest struct {
	URL    string `json:"url"`
	Prompt string `json:"prompt"`
}

type ImageResponse struct {
	Image string `json:"image"`
}

// InstructPix2Pix 发送请求到 Python API 生成图像
func InstructPix2Pix(prompt, url string) (string, error) {
	// 将请求参数序列化为 JSON
	requestBody, err := json.Marshal(ImageRequest{
		URL:    url,
		Prompt: prompt,
	})
	if err != nil {
		return "", fmt.Errorf("序列化请求数据失败: %v", err)
	}

	// 向 Python API 发送 POST 请求
	resp, err := http.Post("http://localhost:8001/generate", "application/json", bytes.NewBuffer(requestBody))
	if err != nil {
		return "", fmt.Errorf("请求 Python API 失败: %v", err)
	}
	defer resp.Body.Close()

	// 读取响应体
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("读取响应失败: %v", err)
	}

	// 检查响应状态码
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("图像生成失败: %s", body)
	}

	// 解析响应中的图像数据
	var imageResponse ImageResponse
	err = json.Unmarshal(body, &imageResponse)
	if err != nil {
		return "", fmt.Errorf("解析响应数据失败: %v", err)
	}

	return imageResponse.Image, nil
}
