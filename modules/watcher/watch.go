package main

import (
	"context"
	"flag"
	kuberlogicv1 "github.com/kuberlogic/operator/modules/operator/api/v1"
	"github.com/kuberlogic/operator/modules/watcher/api"
	"github.com/kuberlogic/operator/modules/watcher/api/common"
	"github.com/kuberlogic/operator/modules/watcher/k8s"
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
		log.Fatalf("error when receiving base client: %s", err)
	}

	crdClient, err := k8s.GetKuberLogicClient(config)
	if err != nil {
		log.Fatalf("error when receiving rest client: %s", err)
	}

	cluster := &kuberlogicv1.KuberLogicService{}
	err = crdClient.
		Get().
		Resource("kuberlogicservices").
		Namespace("default").
		Name(params.targetCluster).
		Do(context.TODO()).
		Into(cluster)
	if err != nil {
		log.Fatalf("Error receiving resource: %s", err)
	}

	session, err := api.GetSession(cluster, client, params.targetDb, params.targetTable)
	if err != nil {
		log.Fatalf("error when receiving session: %s", err)
	}

	log.Println(session)

	if err := session.SetupDDL(); err != nil {
		log.Fatal(err)
	}
	session.RunQueries(params.delay, params.duration)
	for {
		time.Sleep(time.Minute)
	}
}
