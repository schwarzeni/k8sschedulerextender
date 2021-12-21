自定义 k8s 打分策略 + 自实现节点资源采集上报

- agent/agent.go: agent 程序，收集系统的相关数据定期上报至 master
- collector/collector.go: 系统资源信息收集工具包，被 agent 调用
- master/master.go: master 程序主入口，用于接收存储 agent 上报的信息、处理 scheduler 调度请求等、修改自定义权重（API介绍略，详见 master/master.go 文件）
- model/node.go: 项目中涉及到的数据结构的定义
processor/processor.go: 对收集到的数据进行处理的模块，被 master 调用