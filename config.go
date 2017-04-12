package main

// Scene 场景
type Scene struct {
	Name            string
	IndexNamePerfix string `mapstructure:"index_name_perfix" json:"index_name_perfix"`
	Cron            string
	TimeRange       int `mapstructure:"time_range" json:"time_range"`
	Worker          int
	Taches          map[string]Tache
	Links           []Link
}

// Link 描述环节之间的连接
type Link struct {
	From struct {
		Tache string `json:"tache"`
		Field string `json:"field"`
	} `json:"from"`
	To struct {
		Tache string `json:"tache"`
		Field string `json:"field"`
	} `json:"to"`
}

// Tache 描述环节信息
type Tache struct {
	IndexNamePerfix string `mapstructure:"index_name_perfix" json:"index_name_perfix"`
	TimeField       string `mapstructure:"time_field" json:"time_field"`
	IDField         string `mapstructure:"id_field" json:"id_field"`
}

// Config 用于存储配置
type Config struct {
	Scenes []Scene
}
