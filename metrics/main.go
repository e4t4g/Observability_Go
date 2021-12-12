package main

import (
	"log"
	"math/rand"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

const (
	LabelMethod = "method"
)

type Metric struct {
	opsProcessed *prometheus.CounterVec
}

func (m *Metric) NewMetricsMiddleware() error {
	m.opsProcessed = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "myapp_processed_ops_total",
			Help: "The total number of processed events",
		}, []string{"method", "status"})

	prometheus.MustRegister(m.opsProcessed)

	return nil

}

func main() {
	m := Metric{}

	go func() {
		for {
			m.opsProcessed.With(prometheus.Labels{"method": LabelMethod, "status": "OK"}).Inc()
			time.Sleep(2 * time.Second)
		}
	}()

	if err := m.NewMetricsMiddleware(); err != nil {
		log.Fatal(err)
	}

	time.Sleep(time.Duration(rand.Intn(1000)) * time.Millisecond)

	http.Handle("/metrics", promhttp.Handler())
	log.Fatal(http.ListenAndServe("0.0.0.0:2112", nil))
}


