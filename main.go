package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	_ "github.com/Lensual/ServerBackup/backuper/rsyncmd"
	_ "github.com/Lensual/ServerBackup/compressor/zstd"
	_ "github.com/Lensual/ServerBackup/storage/local"

	"github.com/Lensual/ServerBackup/config"
	"github.com/Lensual/ServerBackup/worker"
	"github.com/imdario/mergo"
	"github.com/robfig/cron/v3"
	"github.com/spf13/viper"
)

var configPath = flag.String("config", "./config", "config path")
var runNow = flag.Bool("runnow", false, "run immediately")

func main() {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("$HOME/.serverbackup")
	viper.AddConfigPath("/etc/serverbackup/")
	viper.AddConfigPath(".")

	flag.Parse()

	if configPath != nil {
		viper.AddConfigPath(*configPath)
	}

	err := viper.ReadInConfig()
	if err != nil {
		panic(fmt.Errorf("fatal error config file: %w", err))
	}

	conf := config.Config{}
	err = viper.Unmarshal(&conf)
	if err != nil {
		panic(err)
	}

	c := cron.New(cron.WithParser(cron.NewParser(
		cron.SecondOptional | cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow | cron.Descriptor,
	)))
	for taskName, taskConf := range conf.Task {
		err := mergo.Merge(&taskConf, conf.Global)
		if err != nil {
			log.Printf("%v\n", err)
			continue
		}

		w := worker.New(
			worker.WithTaskName(taskName),
			worker.WithTaskConfig(taskConf),
			worker.WithLogger(
				log.New(os.Stdout, fmt.Sprintf("[%s] ", taskName), log.LstdFlags),
			),
		)

		if *runNow {
			w.Run()
			continue
		}
		_, err = c.AddJob(taskConf.Cron, w)
		if err != nil {
			log.Printf("%v\n", err)
			continue
		}

		log.Printf("Task Added: %v\n", taskName)
	}
	if *runNow {
		return
	}
	c.Run()
}
