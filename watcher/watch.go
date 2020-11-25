package main

import (
	"context"
	"flag"
	cloudlinuxv1 "gitlab.com/cloudmanaged/operator/api/v1"
	"gitlab.com/cloudmanaged/operator/watcher/api"
	"gitlab.com/cloudmanaged/operator/watcher/api/common"
	"gitlab.com/cloudmanaged/operator/watcher/k8s"
	"log"
	"time"
)

type Params struct {
	targetCluster string
	targetDb      string
	targetTable   string
	delay         common.Delay
	duration      common.Duration
}

func parseParams() Params {
	params := Params{}
	flag.StringVar(
		&params.targetCluster,
		"cluster",
		"",
		"target cluster")
	flag.StringVar(
		&params.targetDb,
		"db",
		"",
		"target db")
	flag.StringVar(
		&params.targetTable,
		"table",
		"",
		"target table")

	// Delays
	flag.Int64Var(
		&params.delay.MasterRead,
		"master-read-delay",
		1000,
		"master read delay (msec)")
	flag.Int64Var(
		&params.delay.ReplicaRead,
		"replica-read-delay",
		1000,
		"replica read delay (msec)")
	flag.Int64Var(
		&params.delay.MasterWrite,
		"master-write-delay",
		1000,
		"master write delay (msec)")

	// Durations
	flag.Int64Var(
		&params.duration.MasterRead,
		"master-read-duration",
		0,
		"master read duration (sec)")
	flag.Int64Var(
		&params.duration.ReplicaRead,
		"replica-read-duration",
		0,
		"replica read duration (sec)")
	flag.Int64Var(
		&params.duration.MasterWrite,
		"master-write-duration",
		0,
		"master write duration (sec)")

	flag.Parse()
	if params.targetCluster == "" || params.targetDb == "" || params.targetTable == "" {
		log.Fatal("Several variables are undefined")
	}
	return params
}

func main() {
	params := parseParams()

	config, err := k8s.GetConfig()
	if err != nil {
		log.Fatal(err)
	}

	client, err := k8s.GetBaseClient(config)
	if err != nil {
		log.Fatal(err)
	}

	crdClient, err := k8s.GetCloudmanagedClient(config)
	if err != nil {
		log.Fatal(err)
	}

	cluster := &cloudlinuxv1.CloudManaged{}
	err = crdClient.
		Get().
		Resource("cloudmanageds").
		Namespace("default").
		Name(params.targetCluster).
		Do(context.TODO()).
		Into(cluster)
	if err != nil {
		log.Fatal(err)
	}

	watcher, err := api.GetWatcher(cluster, client, params.targetDb, params.targetTable)
	if err != nil {
		log.Fatal(err)
	}

	log.Println(watcher)

	if err := watcher.SetupDDL(); err != nil {
		log.Fatal(err)
	}
	watcher.RunQueries(params.delay, params.duration)
	for {
		time.Sleep(time.Minute)
	}
}
