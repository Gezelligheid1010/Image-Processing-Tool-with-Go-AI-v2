package controller

import (
	"backend/dao/api"
	"backend/dao/mq"
	"backend/dao/redis"
	"backend/models"
	"encoding/json"
	"fmt"
	"github.com/streadway/amqp"
	"log"
)

func ConsumeUploadTasks(queueName string) {
	msgs, err := mq.RabbitMQChannel.Consume(
		queueName, // 队列名称
		"",        // 消费者名称
		true,      // true：自动应答；false：手动应答
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		log.Printf("Failed to consume messages from queue %s: %v", queueName, err)
		return
	}

	log.Printf("Consumer started for queue: %s", queueName)

	taskResultChannel := make(chan models.TaskResult) // 用于接收任务处理结果

	for d := range msgs {
		go func(d amqp.Delivery) {
			var task models.UploadTaskPayload
			err := json.Unmarshal(d.Body, &task)
			if err != nil {
				taskResultChannel <- models.TaskResult{
					TaskID: task.TaskID,
					Status: "failed",
					Error:  fmt.Errorf("Failed to unmarshal upload task: %v", err),
				}
				return
			}

			// 调用图像上传 API
			imageURL, err := api.UploadImageToOSS(task.Image)
			if err != nil {
				taskResultChannel <- models.TaskResult{
					TaskID: task.TaskID,
					Status: "failed",
					Error:  fmt.Errorf("Failed to upload image: %v", err),
				}
				return
			}

			// 更新任务状态为 "uploaded"
			err = redis.UpdateTaskStatus(task.TaskID, "uploaded", imageURL)
			if err != nil {
				taskResultChannel <- models.TaskResult{
					TaskID: task.TaskID,
					Status: "failed",
					Error:  fmt.Errorf("Failed to update task status: %v", err),
				}
				return
			}

			// 推送后续处理任务
			switch task.TaskType {
			case "process_image":
				newTask := map[string]string{
					"task_id":   task.TaskID,
					"prompt":    task.Prompt,
					"image_url": imageURL,
				}
				taskJSON, _ := json.Marshal(newTask)
				err := mq.PublishMessage("ai_tasks", "image.process", taskJSON)
				if err != nil {
					taskResultChannel <- models.TaskResult{
						TaskID: task.TaskID,
						Status: "failed",
						Error:  fmt.Errorf("Failed to publish process_image task: %v", err),
					}
					return
				}
			default:
				taskResultChannel <- models.TaskResult{
					TaskID: task.TaskID,
					Status: "completed",
					Error:  nil,
				}
			}
		}(d)
	}
	// 处理每个任务的结果
	go func() {
		for result := range taskResultChannel {
			if result.Error != nil {
				log.Printf("Task %s failed: %v", result.TaskID, result.Error)
				continue
			}
			log.Printf("Task %s completed with status: %s", result.TaskID, result.Status)
		}
	}()
}
