package main

import (
	"sync"
	"systeminfoagent/model"
)

var dataMap = map[string]*model.NodeInfoRecord{}
var lock sync.Mutex

func getAllNodeLatestMetrics() []*model.NodeFullMetric {
	lock.Lock()
	defer lock.Unlock()
	var res []*model.NodeFullMetric
	for _, v := range dataMap {
		res = append(res, &(v.Metrics[len(v.Metrics)-1]))
	}
	return res
}

func getRecord(nodeid string) (*model.NodeInfoRecord, bool, error) {
	lock.Lock()
	defer lock.Unlock()
	record, ok := dataMap[nodeid]
	if !ok {
		record = &model.NodeInfoRecord{ID: nodeid}
	}
	return record, ok, nil
}

func saveRecord(nodeid string, record *model.NodeInfoRecord) error {
	lock.Lock()
	defer lock.Unlock()
	dataMap[nodeid] = record
	return nil
}
