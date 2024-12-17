package logic

import (
	"backend/dao/mq"
	"backend/dao/mysql"
	"backend/models"
	"backend/pkg/snowflake"
	"encoding/json"
	"fmt"
	"go.uber.org/zap"
)

// CreateCategory 创建Category
func CreateCategory(c *models.Category) (err error) {
	// 1、 生成post_id(生成分类ID)
	CategoryId, err := snowflake.GetID()
	if err != nil {
		zap.L().Error("snowflake.GetID() failed", zap.Error(err))
		return
	}
	c.CategoryId = CategoryId

	// 2. 插入分组
	c.CoverUrl = "pending"
	err = mysql.CreateCategory(c)
	if err != nil {
		zap.L().Error("CreateCategory failed", zap.Error(err))
		return err
	}

	// 3. 构建上传任务
	taskID, err := snowflake.GetID()
	if err != nil {
		zap.L().Error("snowflake.GetID() failed", zap.Error(err))
		return err
	}
	task := map[string]string{
		"task_id":   fmt.Sprintf("%d", taskID),
		"task_type": "category_cover_upload",
		"author_id": fmt.Sprintf("%d", c.CategoryId),
		"image":     c.Cover, // Base64 编码的图像
	}

	// 4. 推送任务到 RabbitMQ
	taskJSON, _ := json.Marshal(task)
	err = mq.PublishMessage("ai_tasks", "image.upload", taskJSON)
	if err != nil {
		zap.L().Error("PublishMessage to queue failed", zap.Error(err))
		return err
	}

	fmt.Printf("Category created with pending cover upload: %d\n", c.CategoryId)
	return nil
}

func DeleteCategory(categoryID string) error {

	// 1. 删除该分类下的所有作品
	err := mysql.DeleteWorksByCategoryID(categoryID)
	if err != nil {
		return err
	}

	// 3. 删除分类本身
	err = mysql.DeleteCategory(categoryID)
	if err != nil {
		return err
	}

	return nil
}
