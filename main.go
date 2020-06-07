package main

import (
	"fmt"
	"net/http"
	"pkg/k8sDiscovery"
	"strconv"
	"sync"
	"time"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sirupsen/logrus"
)

type Settings struct {
	Port        string `default:"8080"`
	MqPodName   string `split_words:"true"`
	MqNamespace string `split_words:"true"`
	MqContainer string `split_words:"true"`
	MqQueueName string `split_words:"true"`
	MqManager   string `split_words:"true"`

	PollInterval int `default:"30" split_words:"true"`
}

var conf Settings
var gClientSet kubernetes.Interface
var gRestConfig *rest.Config
var testQueue string
var testQueueDepth int
var mutex sync.Mutex

func testQueueHandler(w http.ResponseWriter, r *http.Request) {
	queries := r.URL.Query()
	queue, ok := queries["q"]
	if !ok {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "queue name is missing.")
		return
	}
	qName := queue[0]
	depth, ok := queries["d"]
	if !ok {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "queue name is missing. Example: q=queue1")
		return
	}
	qDepth, err := strconv.Atoi(depth[0])
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "queue depth is missing. Example: d=100")
		return
	}

	mutex.Lock()
	testQueue = qName
	testQueueDepth = qDepth
	mutex.Unlock()
	fmt.Fprintf(w, "queue is set as %s %d", qName, qDepth)
}

func main() {
	err := godotenv.Load()
	if err != nil {
		logrus.Errorf("Error loading .env file:%v. Ignore and continue", err)
	}

	err = envconfig.Process("EXP", &conf)
	if err != nil {
		logrus.Fatalf("Faile to parse envconfig")
	}

	logrus.Infof("conf: %v", conf)

	gClientSet, gRestConfig, err = k8sDiscovery.K8s()
	if err != nil {
		logrus.Fatalf("Could not get K8s.")
	}

	http.Handle("/metrics", promhttp.Handler())
	http.HandleFunc("/testQueue", testQueueHandler)

	q_depth_guage_vec := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "mq",
			Subsystem: "mon",
			Name:      "queue_depth",
			Help:      "MQ queue depth",
		},
		[]string{
			"queue_name",
		},
	)
	prometheus.MustRegister(q_depth_guage_vec)

	go func() {
		for {
			dict, err := monitor_queue_depth()
			if err != nil {
				logrus.Warning("Failed to get queue depth")
			}
			for qn, depth := range dict {
				q_depth_guage_vec.With(prometheus.Labels{"queue_name": qn}).Set(float64(depth))
			}

			if testQueue != "CLEAR" && testQueue != "" {
				q_depth_guage_vec.With(prometheus.Labels{"queue_name": testQueue}).Set(float64(testQueueDepth))
			}

			time.Sleep(time.Duration(conf.PollInterval) * time.Second)
		}
	}()

	logrus.Infof("Listening on port: %s", conf.Port)
	err = http.ListenAndServe(fmt.Sprintf(":%s", conf.Port), nil)
	if err != nil {
		logrus.Fatalf("Failed to start server:%v", err)
	}
}
