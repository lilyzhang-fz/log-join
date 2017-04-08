package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/robfig/cron"
	"github.com/spf13/viper"
)

var a int
var finish bool

func main() {
	i := 0
	c := cron.New()
	spec := "*/1 * * * * *"
	
	c.AddFunc(spec, func() {
		i++
		time.Sleep(5 * time.Second)
		log.Println("start1 - ", i)
		c.Entries()
	})

	c.Start()

	select {} // block forever
}
func getConfig() {
	viper.SetDefault("ContentDir", "content")
	viper.SetDefault("LayoutDir", "layouts")
	viper.SetDefault("Taxonomies", map[string]string{"tag": "tags", "category": "categories"})

	viper.SetConfigName("config")          // name of config file (without extension)
	viper.AddConfigPath("/etc/log-join/")  // path to look for the config file in
	viper.AddConfigPath("$HOME/.log-join") // call multiple times to add many search paths
	viper.AddConfigPath(".")               // optionally look for config in the working directory
	err := viper.ReadInConfig()            // Find and read the config file
	if err != nil {                        // Handle errors reading the config file
		panic(fmt.Errorf("fatal error config file: %s ", err))
	}

	viper.WatchConfig()
	viper.OnConfigChange(func(e fsnotify.Event) {
		fmt.Println("Config file changed:", e.Name)
	})

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
