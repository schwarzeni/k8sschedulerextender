package main

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"systeminfoagent/model"
	"systeminfoagent/processor"
	"time"

	"github.com/gin-gonic/gin"
	schedulerapi "k8s.io/kube-scheduler/extender/v1"
)

var metricProcessor = processor.NewScoreProcessor()

var offlineTimeBound = time.Second * 5

func processdata(rawMetric *model.NodeMetric) {
	// 查询往期的记录
	record, _, err := getRecord(rawMetric.NodeInfo.ID)
	if err != nil {
		log.Println("[err] get record", err)
		return
	}
	defer func() {
		// 存储数据
		if err := saveRecord(rawMetric.NodeInfo.ID, record); err != nil {
			log.Println("[err] save record", err)
			return
		}
	}()

	record.Metrics = append(record.Metrics, model.NodeFullMetric{
		RawMetric:  *rawMetric,
		NodeInfo:   rawMetric.NodeInfo,
		Statistics: model.Statistics{},
	})

	// 判断节点是否下线过，如果是，则计算时长
	checkIfOffline(record)

	// 判断数据是否合法，如果是，计算计算标准差平均值；如果不是，则使用上一次的数据
	for _, processor := range processor.ProcessorMap {
		processor.N(record)
		processor.Even(record)
		processor.Variance(record)
	}
	// fmt.Printf("%+v, %+v\n", record.Metrics[len(record.Metrics)-1].Statistics.CPU, record.Metrics[len(record.Metrics)-1].RawMetric.CPU.Idle)
	// fmt.Printf("%+v, %+v, %+v\n", record.DownDuration.Seconds(), record.Metrics[len(record.Metrics)-1].Statistics, record.Metrics[len(record.Metrics)-1].RawMetric)
}

func checkIfOffline(record *model.NodeInfoRecord) {
	if len(record.Metrics) == 1 {
		return
	}
	now := time.Now()
	prevRecord := record.Metrics[len(record.Metrics)-2]
	if v := now.Sub(prevRecord.RawMetric.Timestamp); v > offlineTimeBound {
		record.DownDuration += v
	}
}

func priorityFunc(c *gin.Context) {
	var buf bytes.Buffer
	var extendArgs schedulerapi.ExtenderArgs
	var hostPriorityList *schedulerapi.HostPriorityList
	body := io.TeeReader(c.Request.Body, &buf)
	if err := json.NewDecoder(body).Decode(&extendArgs); err != nil {
		log.Printf("[err] json decode err:%v", err)
		hostPriorityList = &schedulerapi.HostPriorityList{}
	} else {
		hostPriorityList = prioritize(extendArgs)
	}
	log.Println("[debug] priority list: ", hostPriorityList)
	if response, err := json.Marshal(&hostPriorityList); err != nil {
		log.Fatal(err)
	} else {
		c.Header("Content-Type", "application/json")
		c.Status(http.StatusOK)
		_, _ = c.Writer.Write(response)
		c.Writer.Flush()
	}
}
