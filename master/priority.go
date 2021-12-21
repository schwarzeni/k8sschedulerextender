package main

import schedulerapi "k8s.io/kube-scheduler/extender/v1"

func prioritize(args schedulerapi.ExtenderArgs) *schedulerapi.HostPriorityList {
	nodes := args.Nodes.Items
	hostPriorityList := make(schedulerapi.HostPriorityList, len(nodes))
	for i, node := range nodes {
		record, ok, _ := getRecord(node.Name)
		var score float64
		if ok && len(record.Metrics) > 0 {
			latestMetric := record.Metrics[len(record.Metrics)-1]
			score, _ = metricProcessor.Score(&latestMetric)
		} else {
			score = 0.0
		}
		hostPriorityList[i] = schedulerapi.HostPriority{
			Host:  node.Name,
			Score: int64(score),
		}
	}
	return &hostPriorityList
}
