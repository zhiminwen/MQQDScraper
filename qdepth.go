package main

import (
	"fmt"
	"pkg/k8sExec"
	"regexp"
	"strconv"
	"strings"

	"github.com/sirupsen/logrus"
)

func monitor_queue_depth() (map[string]int, error) {
	cmds := []string{
		"sh",
		"-c",
		fmt.Sprintf(`echo "display ql(%s) curdepth" | runmqsc -e %s`, conf.MqQueueName, conf.MqManager),
	}
	k8s := k8sExec.New(gClientSet, gRestConfig, conf.MqPodName, conf.MqContainer, conf.MqManager)
	stdout, stderr, err := k8s.Exec(cmds)
	if err != nil {
		logrus.Errorf("Failed to exec:%v", err)
		return map[string]int{}, err
	}

	logrus.Infof("out:%s", stdout)
	logrus.Infof("err:%s", stderr)

	qDepth := parse_queue_depth(string(stdout))
	return qDepth, nil
}

func parse_queue_depth(output string) map[string]int {
	/*
			AMQ8409I: Display Queue details.
		   QUEUE(QueueName)               TYPE(QLOCAL)
			 CURDEPTH(0)
	*/
	qDepthMap := map[string]int{}
	qExp := regexp.MustCompile(`^QUEUE\((.*)\)\s*TYPE`)
	depthExp := regexp.MustCompile(`^CURDEPTH\((.*)\)`)
	var qName, qDepth string
	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimSpace(line)
		queue := qExp.FindStringSubmatch(line)
		if len(queue) > 0 {
			qName = queue[1]
			continue
		}

		queueDepth := depthExp.FindStringSubmatch(line)
		if len(queueDepth) > 0 {
			qDepth = queueDepth[1]
			d, err := strconv.Atoi(qDepth)
			if err != nil {
				continue
			}
			qDepthMap[qName] = d
		}
	}

	return qDepthMap
}
