package logic

import (
	"backend/dao/mq"
	"backend/dao/mysql"
	"backend/models"
	"backend/pkg/jwt"
	"backend/pkg/snowflake"
	"encoding/json"
	"fmt"
	"go.uber.org/zap"
)

// SignUp 注册业务逻辑
func SignUp(p *models.RegisterForm) (error error) {
	// 1、判断用户存不存在
	err := mysql.CheckUserExist(p.UserName)
	if err != nil {
		fmt.Println("CheckUserExist:", err, p.UserName)
		// 数据库查询出错
		return err
	}

	// 2、生成UID
	userId, err := snowflake.GetID()
	if err != nil {
		return mysql.ErrorGenIDFailed
	}

	// 3. 保存用户到数据库（头像 URL 设置为 pending）
	u := models.User{
		UserID:   userId,
		UserName: p.UserName,
		Password: p.Password,
		Email:    p.Email,
		Gender:   p.Gender,
		Avatar:   "pending", // 上传任务完成后更新
	}
	err = mysql.InsertUser(u)
	if err != nil {
		fmt.Println("InsertUser:", err)
		zap.L().Error("InsertUser failed", zap.Error(err))
		return err
	}

	// 4. 构建上传任务
	taskID, err := snowflake.GetID()
	if err != nil {
		zap.L().Error("snowflake.GetID() failed", zap.Error(err))
		return err
	}
	task := map[string]string{
		"task_id":   fmt.Sprintf("%d", taskID),
		"task_type": "avatar_upload",
		"author_id": fmt.Sprintf("%d", userId),
		"image":     p.Avatar, // Base64 编码的图像
	}

	// 5. 推送任务到 RabbitMQ
	taskJSON, _ := json.Marshal(task)
	err = mq.PublishMessage("ai_tasks", "image.upload", taskJSON)
	if err != nil {
		zap.L().Error("PublishMessage to queue failed", zap.Error(err))
		return err
	}

	fmt.Printf("User created with pending avatar upload: %d\n", userId)
	return nil
}

// Login 登录业务逻辑代码
func Login(p *models.LoginForm) (user *models.User, error error) {
	user = &models.User{
		UserName: p.UserName,
		Password: p.Password,
	}
	if err := mysql.Login(user); err != nil {
		return nil, err
	}
	// 生成JWT
	//return jwt.GenToken(user.UserID,user.UserName)
	accessToken, refreshToken, err := jwt.GenToken(user.UserID, user.UserName)
	if err != nil {
		return
	}
	user.AccessToken = accessToken
	user.RefreshToken = refreshToken
	return
}
