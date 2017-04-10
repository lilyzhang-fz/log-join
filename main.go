package main

import (
	"fmt"
	"net/http"

	"github.com/davecgh/go-spew/spew"
	"github.com/spf13/viper"
)

var a int
var finish bool

var config Config

func main() {

	// i := 0
	// c := cron.New()

	// spec := "*/1 * * * * *"
	// c.AddFunc(spec, func() {
	// 	i++
	// 	time.Sleep(5 * time.Second)
	// 	log.Println("start1 - ", i)
	// 	c.Entries()
	// })

	// c.Start()

	// select {} // block forever
	spew.Dump(config)
}

func init() {
	getConfig()
	setDefaultValue()
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

func logJoin() {
	c := make(chan int)
	for index := 0; index < 20; index++ {

		go work(c)
	}
	for {
		a++
		c <- a
	}
}
func work(c chan int) {
	for {
		s := <-c

		get(s)

	}
}

func get(s int) {
	fmt.Printf("开始获取%v\n", s)
	http.Get("http://127.0.0.1:8080/sleep?time=3")
	fmt.Printf("结束获取%v\n", s)
}
