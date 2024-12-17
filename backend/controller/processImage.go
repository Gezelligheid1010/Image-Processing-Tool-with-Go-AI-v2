package controller

import (
	"backend/dao/mq"
	"backend/models"
	"backend/pkg/snowflake"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"net/http"
	"strconv"
)

// ProcessImageHandler 异步处理图像任务
func ProcessImageHandler(c *gin.Context) {
	// 1、获取参数及校验参数
	var image models.Image
	if err := c.ShouldBindJSON(&image); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求数据格式错误"})
		zap.L().Error("Error binding JSON", zap.Error(err))
		return
	}

	//// 2. 获取作者ID，当前请求的UserID(从c取到当前发请求的用户ID)
	//userID, err := getCurrentUserID(c)
	//if err != nil {
	//	zap.L().Error("GetCurrentUserID() failed", zap.Error(err))
	//	ResponseError(c, CodeNotLogin)
	//	return
	//}
	//image.AuthorId = userID

	// 3. 为任务生成唯一 ID
	taskID, err := snowflake.GetID()
	if err != nil {
		zap.L().Error("snowflake.GetID() failed", zap.Error(err))
		ResponseError(c, CodeServerBusy)
		return
	}
	//fmt.Println("taskID:", taskID)

	// 4、构建任务数据
	task := map[string]string{
		"task_id":   strconv.FormatUint(taskID, 10),
		"author_id": strconv.FormatUint(image.CategoryId, 10), // 记录分组
		//"author_id": strconv.FormatUint(userID, 10), // 记录作者
		"task_type": "process_image", // 指定任务类型
		"prompt":    image.Prompt,    // 图像处理的 prompt
		"image_url": image.OriUrl,    // 上传后图像 URL
		"image":     image.OriImage,  // 要上传的图像base64编码
	}
	//fmt.Println("taskID:", taskID)
	taskJSON, err := json.Marshal(task)
	if err != nil {
		zap.L().Error("生成任务JSON失败", zap.Error(err))
		fmt.Println("生成任务JSON失败:", err)
		ResponseError(c, CodeServerBusy)
		return
	}
	//fmt.Println("taskID:", taskID)

	// 5、将任务推送到 RabbitMQ 队列
	err = mq.PublishMessage("ai_tasks", "image.upload", taskJSON)
	if err != nil {
		zap.L().Error("推送任务到队列失败", zap.Error(err))
		ResponseError(c, CodeServerBusy)
		return
	}

	// 6、返回任务 ID 给前端
	ResponseSuccess(c, strconv.FormatUint(taskID, 10))
}
