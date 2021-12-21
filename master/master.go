package main

import (
	"log"
	"math"
	"net/http"
	"strconv"
	"systeminfoagent/model"
	"systeminfoagent/processor"

	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.New()
	ch := make(chan *model.NodeMetric)
	r.PUT("/api/v1/agenthealth/:nodeid", func(c *gin.Context) {
		// nodeID := c.Param("nodeid")
		rawMetric := &model.NodeMetric{}
		if err := c.BindJSON(rawMetric); err != nil {
			log.Println("[err] parse json:", err)
			c.Status(http.StatusInternalServerError)
			return
		}
		// log.Println(nodeID, *rawMetric)
		ch <- rawMetric
	})
	r.POST("/api/v1/k8sextension/prioritize", func(c *gin.Context) {
		log.Println("[debug] access priority")
		priorityFunc(c)
	})
	r.PUT("/api/v1/processor/:id/:weight", func(c *gin.Context) {
		processorID := c.Param("id")
		newWeight := c.Param("weight")

		w, err := strconv.Atoi(newWeight)
		id, err2 := strconv.Atoi(processorID)
		_, ok := processor.ProcessorMap[processor.ProcessorType(id)]
		if err != nil || err2 != nil || w < 0 || w > math.MaxInt32 || !ok {
			c.Status(http.StatusBadRequest)
			return
		}
		processor.ProcessorMap[processor.ProcessorType(id)].ExtraWeight(int32(w))
	})
	go func() {
		for metric := range ch {
			processdata(metric)
		}
	}()
	if err := r.Run(":8080"); err != nil {
		log.Fatal(err)
	}
}
