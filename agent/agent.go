package main

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"systeminfoagent/collector"
	"systeminfoagent/model"
	"time"
)

type Config struct {
	NodeID     string
	MasterAddr string
}

func main() {
	config := &Config{
		// NodeID: fmt.Sprintf("%d", time.Now().UnixNano()),
		// NodeID:     "aa",
		NodeID:     os.Args[1],
		MasterAddr: "http://10.211.55.52:8080",
	}
	c := collector.NewDefaultCollector()
	httpClient := http.DefaultClient
	metricURL := config.MasterAddr + "/api/v1/agenthealth/" + config.NodeID
	for {
		time.Sleep(time.Second)
		nodeMetric := &model.NodeMetric{}
		if err := c.Collect(nodeMetric); err != nil {
			log.Println("[err] collect metric:", err)
			continue
		}
		nodeMetric.NodeInfo = model.NodeInfo{ID: config.NodeID}
		binaryData, err := json.Marshal(nodeMetric)
		if err != nil {
			log.Println("[err] marshal json data:", err)
			continue
		}
		nodeMetric.Timestamp = time.Now()
		req, err := http.NewRequest(http.MethodPut, metricURL, bytes.NewBuffer(binaryData))
		if err != nil {
			log.Println("[err] gen http request:", err)
			continue
		}
		resp, err := httpClient.Do(req)
		if err != nil {
			log.Println(err)
			continue
		}
		_ = resp.Body.Close()
	}
}
