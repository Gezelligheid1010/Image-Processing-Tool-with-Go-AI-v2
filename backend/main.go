package main

import (
	"backend/controller"
	"backend/dao/mq"
	"backend/dao/mysql"
	"backend/dao/redis"
	"backend/logger"
	"backend/pkg/snowflake"
	"backend/routers"
	"backend/settings"
	"flag"
	"fmt"
	"os"
	"runtime/trace"
)

// @host 127.0.0.1:8081
// @BasePath /api/v1/
func main() {
	f, _ := os.Create("trace.out")
	defer f.Close()
	trace.Start(f)
	defer trace.Stop()

	var confFile string
	flag.StringVar(&confFile, "conf", "./conf/config.yaml", "配置文件")
	flag.Parse()
	// 加载配置
	if err := settings.Init(); err != nil {
		fmt.Printf("load config failed, err:%v\n", err)
		return
	}

	if err := logger.Init(settings.Conf.LogConfig, settings.Conf.Mode); err != nil {
		fmt.Printf("init logger failed, err:%v\n", err)
		return
	}

	if err := mysql.Init(settings.Conf.MySQLConfig); err != nil {
		fmt.Printf("init mysql failed, err:%v\n", err)
		return
	}
	defer mysql.Close() // 程序退出关闭数据库连接

	if err := redis.Init(settings.Conf.RedisConfig); err != nil {
		fmt.Printf("init redis failed, err:%v\n", err)
		return
	}
	defer redis.Close()

	// 雪花算法生成分布式ID
	if err := snowflake.Init(1); err != nil {
		fmt.Printf("init snowflake failed, err:%v\n", err)
		return
	}

	if err := controller.InitTrans("zh"); err != nil {
		fmt.Printf("init validator Trans failed,err:%v\n", err)
		return
	}

	// 初始化 RabbitMQ
	rabbitMQConfig := settings.Conf.RabbitMQConfig
	if err := mq.InitRabbitMQ(*rabbitMQConfig); err != nil {
		fmt.Printf("Failed to initialize RabbitMQ: %v\n", err)
		return
	}
	defer mq.CloseRabbitMQ()

	fmt.Println("RabbitMQ initialized successfully")

	//for i := 0; i < 5; i++ { // 启动多个图像处理消费者
	//	go controller.ConsumeProcessTasks("process_image")
	//}

	// 注册路由
	r := routers.SetupRouter(settings.Conf.Mode)

	err := r.Run(fmt.Sprintf(":%d", settings.Conf.Port))
	if err != nil {
		fmt.Printf("run server failed, err:%v\n", err)
		return
	}

	//go controller.ConsumeUploadTasks("upload_image")
	//go controller.ConsumeProcessTasks("process_image")

	//select {} // 阻止主线程退出
}
