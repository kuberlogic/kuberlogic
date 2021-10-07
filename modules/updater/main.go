/*
 * CloudLinux Software Inc 2019-2021 All Rights Reserved
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *    http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/coreos/go-semver/semver"
	kuberlogicv1 "github.com/kuberlogic/kuberlogic/modules/operator/api/v1"
	"io/ioutil"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/apimachinery/pkg/util/json"
	k8scheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"log"
	"net/http"
	"os"
	"sort"
	"strings"
	"time"
)

type TagList struct {
	Name string   `json:"name"`
	Tags []string `json:"tags"`
}

var envRequired = []string{
	"REGISTRY_API",
}

var daysOfWeek = map[string]time.Weekday{
	"Sunday":    time.Sunday,
	"Monday":    time.Monday,
	"Tuesday":   time.Tuesday,
	"Wednesday": time.Wednesday,
	"Thursday":  time.Thursday,
	"Friday":    time.Friday,
	"Saturday":  time.Saturday,
}

func parseWeek(weekday string) (time.Weekday, error) {
	value, ok := daysOfWeek[weekday]
	if !ok {
		return time.Sunday, fmt.Errorf("unknown weekday %s", weekday)
	}
	return value, nil
}

func (p *TagList) getLatestTag() string {
	tags := p.Tags
	sort.SliceStable(tags, func(i, j int) bool {
		vA, vB := semver.New(tags[i]), semver.New(tags[j])
		return vB.LessThan(*vA)
	})
	return tags[0]
}

func checkEnv() error {
	for _, envVariable := range envRequired {
		if value := os.Getenv(envVariable); value == "" {
			return errors.New(fmt.Sprintf(
				"required env variable is undefined: %s", envVariable,
			))
		}
	}
	return nil
}

func fetchTags(name string) (*TagList, error) {
	baseUrl := os.Getenv("REGISTRY_API")

	spaceClient := http.Client{
		Timeout: time.Second * 2, // Timeout after 2 seconds
	}

	baseUrl = strings.TrimSuffix(baseUrl, "/")
	endpoint := fmt.Sprintf("%s/kuberlogic/%s/tags/list", baseUrl, name)
	req, err := http.NewRequest(http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}

	res, err := spaceClient.Do(req)
	if err != nil {
		return nil, err
	}

	if res.Body != nil {
		defer res.Body.Close()
	}

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	tagList := &TagList{}
	err = json.Unmarshal(body, tagList)
	if err != nil {
		return nil, err
	}

	return tagList, nil
}

func needUpgrade(a, b string) bool {
	vA, vB := semver.New(a), semver.New(b)
	return vA.Major == vB.Major && vA.Minor == vB.Minor && vB.Patch > vA.Patch
}

func GetConfig() (*rest.Config, error) {
	// check in-cluster usage
	if cfg, err := rest.InClusterConfig(); err == nil {
		return cfg, nil
	}

	// use the current context in kubeconfig
	conf, err := clientcmd.BuildConfigFromFlags("", "/root/.kube/config")
	if err != nil {
		return nil, err
	}
	return conf, err
}

func main() {
	if err := checkEnv(); err != nil {
		log.Fatal(err)
	}

	// creates the in-cluster config
	config, err := GetConfig()
	if err != nil {
		log.Fatal(err)
	}

	err = kuberlogicv1.AddToScheme(k8scheme.Scheme)
	if err != nil {
		log.Fatal(err)
	}

	crdConfig := *config
	crdConfig.ContentConfig.GroupVersion = &kuberlogicv1.GroupVersion
	crdConfig.APIPath = "/apis"
	crdConfig.NegotiatedSerializer = serializer.NewCodecFactory(k8scheme.Scheme)
	crdConfig.UserAgent = rest.DefaultKubernetesUserAgent()

	restClient, err := rest.UnversionedRESTClientFor(&crdConfig)
	if err != nil {
		panic(err)
	}

	result := kuberlogicv1.KuberLogicServiceList{}
	err = restClient.Get().
		Resource("kuberlogicservices").
		Do(context.TODO()).
		Into(&result)
	if err != nil {
		log.Fatal(err)
	}
	if len(result.Items) == 0 {
		log.Println("KuberLogicService resources was not found")
		return
	}

	for _, item := range result.Items {
		// TBD: check maintenance window before
		now := time.Now()
		weekday, err := parseWeek(item.Spec.MaintenanceWindow.Weekday)
		if err != nil {
			log.Fatal(err)
		}
		startHour := item.Spec.MaintenanceWindow.StartHour
		duration := item.Spec.MaintenanceWindow.DurationHours
		if weekday != now.Weekday() || !(startHour <= now.Hour() && startHour+duration > now.Hour()) {
			log.Println("Not time to upgrade")
			break
		}

		log.Printf("Found %s with version %s\n",
			item.Spec.Type, item.Spec.Version)

		tags, err := fetchTags(item.Spec.Type)
		if err != nil {
			log.Fatal(err)
		}
		latest := tags.getLatestTag()
		log.Printf("Latest tag was found %s\n", latest)

		if needUpgrade(item.Spec.Version, latest) {
			log.Printf("Version %s of %s should be upgraded onto %s\n",
				item.Spec.Version, item.Spec.Type, latest)

			updatedResult := kuberlogicv1.KuberLogicService{}
			item.Spec.Version = latest
			err = restClient.Put().
				Namespace("default").
				Resource("kuberlogicservices").
				Name(item.Name).
				Body(&item).
				Do(context.TODO()).
				Into(&updatedResult)
			if err != nil {
				log.Fatal(err)
			}
			log.Println("KuberLogicService resource is successfully changed")

		} else {
			log.Println("No need upgrade")
		}
	}
}
