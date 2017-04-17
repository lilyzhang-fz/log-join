package main

import (
	"encoding/json"
	"fmt"
	"io"
	"sync"
	"time"

	"github.com/siddontang/go/log"

	"github.com/davecgh/go-spew/spew"
	"github.com/tidwall/gjson"
	"golang.org/x/net/context"
	elastic "gopkg.in/olivere/elastic.v5"
)

// Scene 场景
type Scene struct {
	Name            string
	IndexNamePerfix string `mapstructure:"index_name_perfix" json:"index_name_perfix"`
	Cron            string
	TimeRange       int `mapstructure:"time_range" json:"time_range"`
	Worker          int
	Taches          map[string]Tache
	Links           []Link
	FirstTache      string   `mapstructure:"first_tache" json:"first_tache"`
	ESUrl           []string `mapstructure:"es_url" json:"es_url"`
	Running         bool
	Hits            chan elastic.SearchHit
	ESClinet        *elastic.Client
} // Scene 场景

// Join 开始日志合并
func (s *Scene) Join() {
	defer func() {
		if r := recover(); r != nil {
			log.Errorf("场景 '%s' 执行异常", s.Name)
		}
	}()

	client, err := elastic.NewClient(elastic.SetURL(s.ESUrl...), elastic.SetSniff(false))
	if err != nil {
		// Handle error
		log.Errorf("场景 '%s' 执行异常:%s", s.Name, err)
		return
	}
	s.ESClinet = client

	s.Hits = make(chan elastic.SearchHit, 200)
	go s.GetAllRecodes()
	s.CheckAll()

}

// CheckAll 并发确认每条记录
func (s *Scene) CheckAll() {
	var wg sync.WaitGroup
	logger.Debugf("场景 '%s' 共有确认线程 %d 个", s.Name, s.Worker)
	for w := 0; w < s.Worker; w++ {
		wn := w

		wg.Add(1)
		go func() {
			logger.Debugf("场景 '%s' 确认线程 %d 启动", s.Name, wn)
			for hit := range s.Hits {

				s.Check(hit)
			}
			logger.Debugf("场景 '%s' 确认线程 %d 完毕", s.Name, wn)
			wg.Done()

		}()
	}
	wg.Wait()
	logger.Infof("完成场景 '%s' 所有工作", s.Name)
}

// Check 检验一条记录是否完成，如 完成返回
func (s *Scene) Check(hit elastic.SearchHit) {
	var firstHitJSON map[string]interface{}

	firstHitChar, err := hit.Source.MarshalJSON()

	if err != nil {
		logger.Errorf("场景 '%s' 返序列化开始环节日志失败，ES-id= '%s'", s.Name, hit.Id)
		// return nil, false
		return
	}
	json.Unmarshal(firstHitChar, &firstHitJSON)
	if err != nil {
		logger.Errorf("场景 '%s' 返序列化开始环节日志失败，ES-id= '%s'", s.Name, hit.Id)
		// return nil, false
		return
	}

	checkedHits := map[string][]*elastic.SearchHit{}

	s.AddHits(&checkedHits, s.FirstTache, []*elastic.SearchHit{&hit})
	// var sdfsdf Hits =
	// checkedHits = append(checkedHits, Hits{Tache: s.FirstTache, Hit: []elastic.SearchHit{hit}})
	// 把首环节日志放入以获取的日志列表
	// lastTaches := s.FirstTache
	links, taches := s.GetNextTache([]string{s.FirstTache})

	logger.Debugf("场景 '%s' 获取到起始环节 '%s' 的下属环节 %d 个分别为：%s", s.Name, s.FirstTache, len(links), spew.Sprint(links))
	for len(links) > 0 {
		// 一条线一条线处理
		for _, link := range links {
			// 每条线获取源头有多少条记录
			termQureys := []elastic.Query{}
			for _, fromTache := range checkedHits[link.From.Tache] {
				// 由于无法确认环节一对多的情况下暂时无法判断对应关系，所以为每条记录添加一个 termQureys 有匹配到就算是有记录
				// 默认前端为 logstash 5.x 需要添加 keyword 才能全文匹配
				tmpHitChar, _ := fromTache.Source.MarshalJSON()
				tmpHitKey := gjson.Get(string(tmpHitChar), link.From.Field).String()
				if tmpHitKey == "null" {
					return
				}
				termQureys = append(termQureys, elastic.NewTermQuery(link.To.Field+".keyword", tmpHitKey))
			}

			q := elastic.NewBoolQuery().Should(termQureys...)
			// termQuery := elastic.NewTermQuery(link.To.Field, hitJSON[link.From.Field])
			ctx := context.Background()
			searchquery := s.ESClinet.Search(s.Taches[link.To.Tache].IndexNamePerfix + "*").Query(q)
			searchResult, err := searchquery.Do(ctx)
			if err != nil {
				logger.Errorf("场景 '%s' 查询失败 ，错误内容为%s", s.Name, err)
			}
			// 如果没有记录就马上返回
			if searchResult.TotalHits() == 0 {
				return
			}
			// 如果有记录就加入到 checkedHits
			s.AddHits(&checkedHits, link.To.Tache, searchResult.Hits.Hits)

		}
		links, taches = s.GetNextTache(taches)
	}
	spew.Dump(checkedHits)
	return
}

func (s *Scene) AddHits(checkedHits *map[string][]*elastic.SearchHit, tacheName string, newHits []*elastic.SearchHit) {
	// 校验每一天新记录
	for _, newHit := range newHits {
		same := false
		for _, oldHit := range (*checkedHits)[tacheName] {
			if newHit.Id == oldHit.Id {
				same = true
				break
			}
		}
		// 如果没有找到匹配项目 就加入新 Hits
		if same == false {
			hitList, exit := (*checkedHits)[tacheName]
			if exit {
				(*checkedHits)[tacheName] = append(hitList, newHit)
			} else {
				(*checkedHits)[tacheName] = []*elastic.SearchHit{newHit}
			}
		}
	}

}

func (s *Scene) GetNextTache(ts []string) ([]Link, []string) {
	retLink := []Link{}
	retTache := []string{}
	for _, t := range ts {
		for _, link := range s.Links {
			if link.From.Tache == t {
				retLink = append(retLink, link)
				retTache = append(retTache, link.To.Tache)
			}
		}
	}
	return retLink, retTache
}

// GetAllRecodes 获取所有的开始环节记录
func (s *Scene) GetAllRecodes() {

	logger.Debugf("连接 ES URL = %s", s.ESUrl)

	ctx := context.Background()
	logger.Debugf("正在为场景 '%s' 获取第一环节 '%s' 的未串联日志，获取的时间段为，%s 到 %s ",
		s.Name, s.FirstTache, time.Now().Add(time.Minute*time.Duration(s.TimeRange*-1)), time.Now())
	q := elastic.NewRangeQuery(s.Taches[s.FirstTache].TimeField).
		Gte(time.Now().Add(time.Minute * time.Duration(s.TimeRange*-1)).Format("2006-01-02T15:04:05-07:00")).
		Lte("now")

	scroll := s.ESClinet.Scroll(s.Taches[s.FirstTache].IndexNamePerfix + "*").
		Size(10000).
		Query(q)

	i := 0
	for {
		results, err := scroll.Do(ctx)
		if err == io.EOF {
			if i == 0 {
				logger.Infof("场景 '%s' 没找到未串联记录记录失败：", s.Name)
			}
			break
		}

		if err != nil {
			// Handle error
			logger.Errorf("场景 '%s' 寻找为串联记录记录失败：%s ", s.Name, err)
			break
		}
		i = i + len(results.Hits.Hits)
		logger.Debugf("场景 '%s' 找到 %d 条记录", s.Name, len(results.Hits.Hits))
		for _, hit := range results.Hits.Hits {

			s.Hits <- *hit
		}
		logger.Infof("场景 '%s' 共找到 %d 条记录", s.Name, i)

	}
	close(s.Hits)
}

// SetFirstTache 获取开始环节，开始环节只能有一个
func (s *Scene) SetFirstTache() error {
	// 如果有配置，检查一下配置，通过检查就直接返回配置的值
	if s.FirstTache != "" {
		if s.CheckFirstTache(s.FirstTache) {
			return nil
		}

		logger.Warningf("场景 %s 配置了开始环节为 %s，但经检查 %s 不满足开始环节条件，开始尝试搜索开始环节",
			s.Name, s.FirstTache, s.FirstTache)

	}
	logger.Infof("场景 %s 未配置开始环节，尝试自动搜索开始环节", s.Name)
	for k := range s.Taches {
		if s.CheckFirstTache(k) {
			(*s).FirstTache = k
			logger.Infof("将场景 %s 的开始环节配置为 %s", s.Name, s.FirstTache)
			// logger.Info(spew.Sdump(s))
			return nil
		}
	}
	logger.Errorf("无法为场景 %s 找到开始环节，请检查配置文件。", s.Name)
	return fmt.Errorf("无法为场景 %s 找到开始环节，请检查配置文件。", s.Name)
}

// CheckFirstTache 用于判断是否为初始环节
func (s *Scene) CheckFirstTache(name string) bool {
	if _, ok := s.Taches[name]; !ok {
		return false
	}

	for _, v := range s.Links {
		if v.To.Tache == name {
			return false
		}
	}
	return true
}

// // GetNextTaches 获取下一环节，下一环节可能会有多个
// func (s *Scene) GetNextTaches(tacheName string) []string {

// }
