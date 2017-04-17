package main

import (
	"fmt"
	"os"
	"os/signal"
	"path"
	"syscall"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/davecgh/go-spew/spew"
	"github.com/gogap/logrus_mate"
	"github.com/robfig/cron"
	"github.com/spf13/viper"
)

var a int
var finish bool
var config Config
var filepath string
var logger *logrus.Logger
var cronList []*cron.Cron

func main() {
	go listenDownSignal()
	logger.Info("开始调度")
	for i := 0; i < len(config.Scenes); i++ {
		logger.Info("创建新 goroutine ", config.Scenes[i].Name)
		go newScene(config.Scenes[i])
	}
	select {}
}

func listenDownSignal() {
	ch := make(chan os.Signal)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
	<-ch
	for _, v := range cronList {
		logger.Debug("开始停止定时任务")
		v.Stop()
		logger.Debug("定时任务停止完成")
	}

	for {
		running := 0
		for _, s := range config.Scenes {
			if s.Running == true {
				running++
				logger.Infof("场景 '%s' 还在运行中,还有 %d 条记录待合并", s.Name, len(s.Hits))
			}
		}
		if running == 0 {
			logger.Info("所有场景都已经结束，准备退出程序")
			os.Exit(0)
		}

		time.Sleep(2 * time.Second)
	}

}

func newScene(s *Scene) {

	c := cron.New()
	// 添加到列表中
	s.Running = false
	// fmt.Printf("%p", &s.Running)
	cronList = append(cronList, c)
	logger.Infof("%s 周期配置为 '%s' ", s.Name, s.Cron)
	c.AddFunc(s.Cron, func() {
		// fmt.Printf("%p", &s.Running)
		if s.Running {
			logger.Infof("正在执行中，放弃此批次-场景名称（%s）", s.Name)
		} else {
			logger.Infof("开始执行-场景名称（%s）", s.Name)
			s.Running = true
			s.Join()
			s.Running = false
			logger.Infof("执行结束-场景名称（%s）", s.Name)
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
	logger.Info(spew.Sdump(config))
	checkConfig()
}

func checkConfig() {
	logger.Debug("正在检查配置")
	ok := true
	for _, s := range config.Scenes {
		err := s.SetFirstTache()
		if err != nil {
			ok = false
		}
	}
	if !ok {
		logger.Panic("配置检查不通过，检查配置")
	}
	logger.Debug("配置检查结束")
}
func getFilePath() {
	ex, err := os.Executable()
	if err != nil {
		panic(err)
	}
	filepath = path.Dir(ex)
}

func setLogger() {
	os.Setenv("RUN_MODE", "production")
	mateConf, err := logrus_mate.LoadLogrusMateConfig(fmt.Sprintf("%s/%s", filepath, "config/log.config"))
	if err != nil {
		panic(fmt.Errorf("can not read log config file: %s", err))
	}
	newMate, err := logrus_mate.NewLogrusMate(mateConf)
	if err != nil {
		panic(fmt.Errorf("can not read log config file: %s", err))
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
