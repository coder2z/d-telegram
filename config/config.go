package config

import (
	"github.com/mitchellh/mapstructure"
	"github.com/natefinch/lumberjack"
	"github.com/spf13/viper"
	"sync"
	"sync/atomic"
)

var (
	configD atomic.Value
	one     sync.Once
)

type Logger struct {
	WriteSyncer lumberjack.Logger `yaml:"write_syncer"`
	Level       string            `yaml:"level"`
	AddCaller   bool              `yaml:"add_caller"`
	CallerSkip  int               `yaml:"caller_skip"`
	Debug       bool              `yaml:"debug"`
}

type Config struct {
	AppID              int      `yaml:"app_id"`
	AppHash            string   `yaml:"app_hash"`
	SessionFile        string   `yaml:"session_file"`
	SuperChannelIDList []int64  `yaml:"super_channel_id_list"`
	WatchChannelIDList []int64  `yaml:"watch_channel_id_list"`
	Log                Logger   `yaml:"log"`
	DownloadPool       int      `yaml:"download_pool"`
	DownloadDir        string   `yaml:"download_dir"`
	MaxFileSize        int64    `yaml:"max_file_size"`
	WatchFileKeyWord   []string `yaml:"watch_file_key_word"`
	Phone              string   `yaml:"phone"`
}

func Get() *Config {
	one.Do(func() {
		vp := viper.New()             //创建viper实例
		vp.SetConfigName("config")    //设置配置文件名
		vp.SetConfigType("yaml")      //设置配置文件类型
		vp.AddConfigPath("./configs") //设置配置文件所在的目录
		err := vp.ReadInConfig()      // Find and read the config file
		if err != nil {
			panic(err)
		}
		var conf = defaultConfig()
		//注释---start
		err = vp.Unmarshal(conf, func(config *mapstructure.DecoderConfig) {
			config.TagName = "yaml" //设置对应标签为"json"
		})
		if err != nil {
			panic(err)
		}
		configD.Store(conf)
	})
	return configD.Load().(*Config)
}
func defaultConfig() *Config {
	return &Config{
		AppID:       0,
		AppHash:     "",
		SessionFile: "./session/session",
		WatchChannelIDList: []int64{
			1778419548,
			1549179925,
		},
		Log: Logger{
			WriteSyncer: lumberjack.Logger{
				Filename:   "./log/TGUpdate.log",
				MaxSize:    100,
				MaxBackups: 20,
				MaxAge:     7,
			},
			Level:      "info",
			AddCaller:  true,
			CallerSkip: 0,
			Debug:      true,
		},
		DownloadPool: 1000,
		DownloadDir:  "./data",
		// 200MB
		MaxFileSize: 200 * 1024 * 1024,
	}
}
