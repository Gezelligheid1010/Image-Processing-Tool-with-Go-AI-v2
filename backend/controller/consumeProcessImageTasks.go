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

func ConsumeProcessTasks(queueName string) {
	msgs, err := mq.RabbitMQChannel.Consume(
		queueName,
		"",
		true,
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

	// 为每个任务启动一个 goroutine 来并行处理
	for d := range msgs {
		go func(d amqp.Delivery) {
			var task models.ProcessTaskPayload
			err := json.Unmarshal(d.Body, &task)
			if err != nil {
				taskResultChannel <- models.TaskResult{
					TaskID: task.TaskID,
					Status: "failed",
					Error:  fmt.Errorf("Failed to unmarshal process task: %v", err),
				}
				return
			}

			redis.UpdateTaskStatus(task.TaskID, "processing", "")

			// 调用图像处理 API
			result, err := api.InstructPix2Pix(task.Prompt, task.ImageURL)
			if err != nil {
				taskResultChannel <- models.TaskResult{
					TaskID: task.TaskID,
					Status: "failed",
					Error:  fmt.Errorf("Failed to process image: %v", err),
				}
				return
			}

			// 更新任务状态为完成
			err = redis.UpdateTaskStatus(task.TaskID, "completed", result)
			if err != nil {
				taskResultChannel <- models.TaskResult{
					TaskID: task.TaskID,
					Status: "failed",
					Error:  fmt.Errorf("Failed to update task status: %v", err),
				}
				return
			}

			taskResultChannel <- models.TaskResult{
				TaskID: task.TaskID,
				Status: "completed",
				Error:  nil,
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
