package processor

import (
	"log"
	"math"
	"sync/atomic"
	"systeminfoagent/model"
)

func debugLogF(s1 string, s2 ...interface{}) {
	log.Printf(s1, s2...)
}

// Processor 用于在打分的时候计算相关指标的分数
// 返回 rawscore 和 weight
// 同时给出计算平均值和方差的接口
// extraWeight: 在计算的时候会 / 100
type Processor interface {
	Score(*model.NodeFullMetric) (float64, float64)
	ExtraWeight(int32)
	N(*model.NodeInfoRecord)
	Even(*model.NodeInfoRecord)
	Variance(*model.NodeInfoRecord)
}

type ProcessorType int

const (
	TCPUPROCESSOR ProcessorType = iota
	TMEMORYPROCESSOR
	TDISKUSAGEPROCESSOR
	TNETWORKPROCESSOR
)

var defaultextraweight int32 = 100

var ProcessorMap map[ProcessorType]Processor = map[ProcessorType]Processor{
	TCPUPROCESSOR:       &CPUProcessor{extraWeight: defaultextraweight},
	TMEMORYPROCESSOR:    &MemoryProcessor{extraWeight: defaultextraweight},
	TDISKUSAGEPROCESSOR: &DiskUsageProcessor{extraWeight: defaultextraweight},
	TNETWORKPROCESSOR:   &NetworkProcessor{MaxRxPerSecond: 1 << 20, extraWeight: defaultextraweight},
}

type ProcessorMapV struct {
	processor Processor
	postFunc  func(ProcessorType, float64, float64)
}

type ScoreProcessor struct {
	processorMap map[ProcessorType]Processor
}

func NewScoreProcessor() *ScoreProcessor {
	return &ScoreProcessor{
		processorMap: ProcessorMap,
	}
}

func (sp *ScoreProcessor) Score(nfm *model.NodeFullMetric, postFuncs ...func(ProcessorType, float64, float64)) (float64, float64) {
	var totalScore, totalWeight float64
	// TODO: 查询需要使用到的 processor 以及设置的 weight
	for processorType, processor := range sp.processorMap {
		score, weight := processor.Score(nfm)
		totalScore += score
		totalWeight += weight
		for _, fn := range postFuncs {
			fn(processorType, score, weight)
		}
	}
	return totalScore / totalWeight, totalWeight
}

type CPUProcessor struct {
	extraWeight int32
}

func (cp *CPUProcessor) ExtraWeight(w int32) {
	atomic.StoreInt32(&cp.extraWeight, w)
}

func (cp *CPUProcessor) Score(nfm *model.NodeFullMetric) (float64, float64) {
	raw := nfm.RawMetric.CPU
	rawScore := (float64(raw.Idle) / float64(raw.System+raw.User+raw.Idle)) * 100.0
	weight := calWeight(nfm.Statistics.CPU) * float64(cp.extraWeight) / 100.0
	debugLogF("[CPU] %s\t%.2f\t%.2f", nfm.NodeInfo.ID, rawScore, weight)
	return rawScore, weight
}

func (*CPUProcessor) N(record *model.NodeInfoRecord) {
	idx := len(record.Metrics) - 1
	if len(record.Metrics) == 0 {
		return
	}
	if len(record.Metrics) == 1 {
		record.Metrics[idx].Statistics.CPU.N = 1
		return
	}
	if !record.Metrics[idx].RawMetric.CPU.Valid {
		record.Metrics[idx].Statistics.CPU.N = record.Metrics[idx-1].Statistics.CPU.N
		return
	}
	record.Metrics[idx].Statistics.CPU.N = record.Metrics[idx-1].Statistics.CPU.N + 1
}

func (*CPUProcessor) Even(record *model.NodeInfoRecord) {
	idx := len(record.Metrics) - 1
	if len(record.Metrics) == 0 {
		return
	}
	if len(record.Metrics) == 1 {
		record.Metrics[idx].Statistics.CPU.Mean = float64(record.Metrics[idx].RawMetric.CPU.Idle)
		return
	}
	prevMean := record.Metrics[idx-1].Statistics.CPU.Mean
	n := float64(record.Metrics[idx-1].Statistics.CPU.N) + 1
	currIdle := float64(record.Metrics[idx].RawMetric.CPU.Idle)
	if !record.Metrics[idx].RawMetric.CPU.Valid {
		record.Metrics[idx].Statistics.CPU.Mean = prevMean
		return
	}
	record.Metrics[idx].Statistics.CPU.Mean = calEven(prevMean, currIdle, n)
}

func (*CPUProcessor) Variance(record *model.NodeInfoRecord) {
	idx := len(record.Metrics) - 1
	if len(record.Metrics) == 0 {
		return
	}
	if len(record.Metrics) == 1 {
		record.Metrics[idx].Statistics.CPU.Variance = 0
		return
	}
	prevVariance := record.Metrics[idx-1].Statistics.CPU.Variance
	n := float64(record.Metrics[idx-1].Statistics.CPU.N) + 1
	currIdle := float64(record.Metrics[idx].RawMetric.CPU.Idle)
	prevMean := record.Metrics[idx-1].Statistics.CPU.Mean
	currMean := record.Metrics[idx].Statistics.CPU.Mean
	if !record.Metrics[idx].RawMetric.CPU.Valid {
		record.Metrics[idx].Statistics.CPU.Variance = prevVariance
		return
	}
	record.Metrics[idx].Statistics.CPU.Variance = calVariance(prevVariance, currIdle, prevMean, currMean, n)
}

type MemoryProcessor struct {
	extraWeight int32
}

func (mp *MemoryProcessor) ExtraWeight(w int32) {
	atomic.StoreInt32(&mp.extraWeight, w)
}

func (mp *MemoryProcessor) Score(nfm *model.NodeFullMetric) (float64, float64) {
	raw := nfm.RawMetric.Memory
	rawScore := (float64(raw.Free) / float64(raw.Total)) * 100.0
	weight := calWeight(nfm.Statistics.Memory) * float64(mp.extraWeight) / 100.0
	debugLogF("[memory] %s\t%.2f\t%.2f", nfm.NodeInfo.ID, rawScore, weight)
	return rawScore, weight
}

func (*MemoryProcessor) N(record *model.NodeInfoRecord) {
	idx := len(record.Metrics) - 1
	if len(record.Metrics) == 0 {
		return
	}
	if len(record.Metrics) == 1 {
		record.Metrics[idx].Statistics.Memory.N = 1
		return
	}
	if !record.Metrics[idx].RawMetric.Memory.Valid {
		record.Metrics[idx].Statistics.Memory.N = record.Metrics[idx-1].Statistics.Memory.N
		return
	}
	record.Metrics[idx].Statistics.Memory.N = record.Metrics[idx-1].Statistics.Memory.N + 1
}

func (*MemoryProcessor) Even(record *model.NodeInfoRecord) {
	idx := len(record.Metrics) - 1
	if len(record.Metrics) == 0 {
		return
	}
	if len(record.Metrics) == 1 {
		record.Metrics[idx].Statistics.Memory.Mean = float64(record.Metrics[idx].RawMetric.Memory.Free)
		return
	}
	prevMean := record.Metrics[idx-1].Statistics.Memory.Mean
	n := float64(record.Metrics[idx-1].Statistics.Memory.N) + 1
	currFree := float64(record.Metrics[idx].RawMetric.Memory.Free)
	if !record.Metrics[idx].RawMetric.Memory.Valid {
		record.Metrics[idx].Statistics.Memory.Mean = prevMean
		return
	}
	record.Metrics[idx].Statistics.Memory.Mean = calEven(prevMean, currFree, n)
}

func (*MemoryProcessor) Variance(record *model.NodeInfoRecord) {
	idx := len(record.Metrics) - 1
	if len(record.Metrics) == 0 {
		return
	}
	if len(record.Metrics) == 1 {
		record.Metrics[idx].Statistics.Memory.Variance = 0
		return
	}
	prevVariance := record.Metrics[idx-1].Statistics.Memory.Variance
	n := float64(record.Metrics[idx-1].Statistics.Memory.N) + 1
	currFree := float64(record.Metrics[idx].RawMetric.Memory.Free)
	prevMean := record.Metrics[idx-1].Statistics.Memory.Mean
	currMean := record.Metrics[idx].Statistics.Memory.Mean
	if !record.Metrics[idx].RawMetric.Memory.Valid {
		record.Metrics[idx].Statistics.Memory.Variance = prevVariance
		return
	}
	record.Metrics[idx].Statistics.Memory.Variance = calVariance(prevVariance, currFree, prevMean, currMean, n)
}

type DiskUsageProcessor struct {
	extraWeight int32
}

func (dp *DiskUsageProcessor) ExtraWeight(w int32) {
	atomic.StoreInt32(&dp.extraWeight, w)
}

func (dup *DiskUsageProcessor) Score(nfm *model.NodeFullMetric) (float64, float64) {
	raw := nfm.RawMetric.Disk
	rawScore := (float64(raw.Free) / float64(raw.Size)) * 100.0
	weight := calWeight(nfm.Statistics.Disk) * float64(dup.extraWeight) / 100.0
	debugLogF("[diskusage] %s\t%.2f\t%.2f", nfm.NodeInfo.ID, rawScore, weight)
	return rawScore, weight
}

func (*DiskUsageProcessor) N(record *model.NodeInfoRecord) {
	idx := len(record.Metrics) - 1
	if len(record.Metrics) == 0 {
		return
	}
	if len(record.Metrics) == 1 {
		record.Metrics[idx].Statistics.Disk.N = 1
		return
	}
	if !record.Metrics[idx].RawMetric.Disk.Valid {
		record.Metrics[idx].Statistics.Disk.N = record.Metrics[idx-1].Statistics.Disk.N
		return
	}
	record.Metrics[idx].Statistics.Disk.N = record.Metrics[idx-1].Statistics.Disk.N + 1
}

func (*DiskUsageProcessor) Even(record *model.NodeInfoRecord) {
	idx := len(record.Metrics) - 1
	if len(record.Metrics) == 0 {
		return
	}
	if len(record.Metrics) == 1 {
		record.Metrics[idx].Statistics.Disk.Mean = float64(record.Metrics[idx].RawMetric.Disk.Free)
		return
	}
	prevMean := record.Metrics[idx-1].Statistics.Disk.Mean
	n := float64(record.Metrics[idx-1].Statistics.Disk.N) + 1
	currFree := float64(record.Metrics[idx].RawMetric.Disk.Free)
	if !record.Metrics[idx].RawMetric.Disk.Valid {
		record.Metrics[idx].Statistics.Disk.Mean = prevMean
		return
	}
	record.Metrics[idx].Statistics.Disk.Mean = calEven(prevMean, currFree, n)
}

func (*DiskUsageProcessor) Variance(record *model.NodeInfoRecord) {
	idx := len(record.Metrics) - 1
	if len(record.Metrics) == 0 {
		return
	}
	if len(record.Metrics) == 1 {
		record.Metrics[idx].Statistics.Disk.Variance = 0
		return
	}
	prevVariance := record.Metrics[idx-1].Statistics.Disk.Variance
	n := float64(record.Metrics[idx-1].Statistics.Disk.N) + 1
	currFree := float64(record.Metrics[idx].RawMetric.Disk.Free)
	prevMean := record.Metrics[idx-1].Statistics.Disk.Mean
	currMean := record.Metrics[idx].Statistics.Disk.Mean
	if !record.Metrics[idx].RawMetric.Disk.Valid {
		record.Metrics[idx].Statistics.Disk.Variance = prevVariance
		return
	}
	record.Metrics[idx].Statistics.Disk.Variance = calVariance(prevVariance, currFree, prevMean, currMean, n)
}

type NetworkProcessor struct {
	MaxRxPerSecond float64
	extraWeight    int32
}

func (np *NetworkProcessor) ExtraWeight(w int32) {
	atomic.StoreInt32(&np.extraWeight, w)
}

func (np *NetworkProcessor) Score(nfm *model.NodeFullMetric) (float64, float64) {
	raw := nfm.RawMetric.Network
	rawScore := ((np.MaxRxPerSecond - float64(raw.RxBytes)) / np.MaxRxPerSecond) * 100.0
	weight := calWeight(nfm.Statistics.Network) * float64(np.extraWeight) / 100.0
	debugLogF("[network] %s\t%.2f\t%.2f", nfm.NodeInfo.ID, rawScore, weight)
	return rawScore, weight
}

func (*NetworkProcessor) N(record *model.NodeInfoRecord) {
	idx := len(record.Metrics) - 1
	if len(record.Metrics) == 0 {
		return
	}
	if len(record.Metrics) == 1 {
		record.Metrics[idx].Statistics.Network.N = 1
		return
	}
	if !record.Metrics[idx].RawMetric.Network.Valid {
		record.Metrics[idx].Statistics.Network.N = record.Metrics[idx-1].Statistics.Network.N
		return
	}
	record.Metrics[idx].Statistics.Network.N = record.Metrics[idx-1].Statistics.Network.N + 1
}

func (*NetworkProcessor) Even(record *model.NodeInfoRecord) {
	idx := len(record.Metrics) - 1
	if len(record.Metrics) == 0 {
		return
	}
	if len(record.Metrics) == 1 {
		record.Metrics[idx].Statistics.Network.Mean = float64(record.Metrics[idx].RawMetric.Network.RxBytes)
		return
	}
	prevMean := record.Metrics[idx-1].Statistics.Network.Mean
	n := float64(record.Metrics[idx-1].Statistics.Network.N) + 1
	currRx := float64(record.Metrics[idx].RawMetric.Network.RxBytes)
	if !record.Metrics[idx].RawMetric.Network.Valid {
		record.Metrics[idx].Statistics.Network.Mean = prevMean
		return
	}
	record.Metrics[idx].Statistics.Network.Mean = calEven(prevMean, currRx, n)
}

func (*NetworkProcessor) Variance(record *model.NodeInfoRecord) {
	idx := len(record.Metrics) - 1
	if len(record.Metrics) == 0 {
		return
	}
	if len(record.Metrics) == 1 {
		record.Metrics[idx].Statistics.Network.Variance = 0
		return
	}
	prevVariance := record.Metrics[idx-1].Statistics.Network.Variance
	n := float64(record.Metrics[idx-1].Statistics.Network.N) + 1
	currRx := float64(record.Metrics[idx].RawMetric.Network.RxBytes)
	prevMean := record.Metrics[idx-1].Statistics.Network.Mean
	currMean := record.Metrics[idx].Statistics.Network.Mean
	if !record.Metrics[idx].RawMetric.Network.Valid {
		record.Metrics[idx].Statistics.Network.Variance = prevVariance
		return
	}
	record.Metrics[idx].Statistics.Network.Variance = calVariance(prevVariance, currRx, prevMean, currMean, n)
}

func calWeight(ms model.MetricStatistics) float64 {
	if ms.Mean == 0 || ms.Variance == 0 {
		return 1
	}
	res := math.Sqrt(ms.Variance) / ms.Mean
	if res >= 1 {
		return 0.001
	}
	return 1 - res
}

func calEven(prevEven, currValue, n float64) float64 {
	return prevEven + (currValue-prevEven)/n
}

func calVariance(prevVariance, currValue, prevEven, currEven, n float64) float64 {
	return (prevVariance*(n-1) + (currValue-prevEven)*(currValue-currEven)) / n
}
