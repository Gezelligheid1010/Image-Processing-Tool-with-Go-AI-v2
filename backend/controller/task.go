package controller

import (
	"backend/dao/redis"
	"github.com/gin-gonic/gin"
)

// TaskStatusHandler 查询任务状态
func TaskStatusHandler(c *gin.Context) {
	// 获取任务ID
	taskID := c.Query("task_id")
	if taskID == "" {
		ResponseError(c, ErrTaskIdEmpty)
		//c.JSON(http.StatusBadRequest, gin.H{"error": "任务ID不能为空"})
	}

	// 从Redis查询任务状态
	//status, result, _ := redis.GetTaskStatus(taskID)
	status, _, _ := redis.GetTaskStatus(taskID)
	if status == "not_found" {
		ResponseError(c, TaskNotFound)
		//c.JSON(http.StatusOK, gin.H{"status": "not_found"})
		//return
	} else if status == "processing" {
		ResponseError(c, TaskProcessing)
		//c.JSON(http.StatusOK, gin.H{"status": "processing"})
		//return
	} else if status == "completed" {
		ResponseSuccess(c, "result")
	} else {
		ResponseError(c, CodeServerBusy)
	}

	//c.JSON(http.StatusInternalServerError, gin.H{"error": "未知错误"})
}
