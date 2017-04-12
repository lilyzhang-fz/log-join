package main

import (
	"fmt"
	"os"
	"path"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/gogap/logrus_mate"
	"github.com/robfig/cron"
	"github.com/spf13/viper"
)

var a int
var finish bool
var config Config
var filepath string
var logger *logrus.Logger

func main() {
	logger.Info("开始调度")
	for i := 0; i < len(config.Scenes); i++ {
		logger.Info("创建新 goroutine ", config.Scenes[i].Name)
		go newScene(config.Scenes[i])
	}
	select {}
}

func newScene(s Scene) {
	Running := false
	c := cron.New()
	logger.Info("cron = ", s.Cron)
	c.AddFunc(s.Cron, func() {
		if Running {
			logger.Info("正在执行中，放弃此批次")
		} else {
			logger.Info("开始执行")
			Running = true
			time.Sleep(time.Second * 3)
			Running = false
			logger.Info("执行结束")
		}
	})
	c.Start()
}

func init() {
	// 获取执行路径
	getFilePath()

	// 设置 logger
	setLogger()

	// 获取场景配置
	getConfig()

	// 设置默认值
	setDefaultValue()
}
func getFilePath() {
	ex, err := os.Executable()
	if err != nil {
		panic(err)
	}
	filepath = path.Dir(ex)
}

func setLogger() {
	mateConf, err := logrus_mate.LoadLogrusMateConfig(fmt.Sprintf("%s/%s", filepath, "config/log.config"))
	if err != nil {
		panic(fmt.Errorf("can not read log config fileL: %s", err))
	}
	newMate, err := logrus_mate.NewLogrusMate(mateConf)
	if err != nil {
		panic(fmt.Errorf("can not read log config fileL: %s", err))
	}
	logger = newMate.Logger("main")
}
func getConfig() {

	viper.SetConfigName("config") // name of config file (without extension)

	viper.AddConfigPath("./config") // optionally look for config in the working directory
	err := viper.ReadInConfig()     // Find and read the config file
	if err != nil {                 // Handle errors reading the config file
		panic(fmt.Errorf("fatal error config file: %s ", err))
	}
	viper.Unmarshal(&config)

}

func setDefaultValue() {
	setWorker()
}

// 默认 worker = 5
func setWorker() {
	for i := 0; i < len(config.Scenes); i++ {
		if config.Scenes[i].Worker <= 0 {
			config.Scenes[i].Worker = 5
		}
	}
}

// func logJoin() {
// 	c := make(chan int)
// 	for index := 0; index < 20; index++ {

// 		go work(c)
// 	}
// 	for {
// 		a++
// 		c <- a
// 	}
// }
// func work(c chan int) {
// 	for {
// 		s := <-c

// 		get(s)

// 	}
// }

// func get(s int) {
// 	fmt.Printf("开始获取%v\n", s)
// 	http.Get("http://127.0.0.1:8080/sleep?time=3")
// 	fmt.Printf("结束获取%v\n", s)
// }
