package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/coreos/go-semver/semver"
	kuberlogicv1 "github.com/kuberlogic/operator/pkg/operator/api/v1"
	"io/ioutil"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/apimachinery/pkg/util/json"
	k8scheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"log"
	"net/http"
	"os"
	"sort"
	"time"
)

type RegistryProject struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	Path      string    `json:"path"`
	ProjectID int       `json:"project_id"`
	Location  string    `json:"location"`
	CreatedAt time.Time `json:"created_at"`
	Tags      []struct {
		Name     string `json:"name"`
		Path     string `json:"path"`
		Location string `json:"location"`
	} `json:"tags"`
}

type RegistryProjects []RegistryProject

func (rp RegistryProjects) get(name string) (*RegistryProject, error) {
	for _, project := range rp {
		if project.Name == name {
			return &project, nil
		}
	}
	return nil, errors.New(fmt.Sprintf("Project with %s is not found", name))
}

func (p *RegistryProject) getLatestTag() string {
	tags := p.Tags
	sort.SliceStable(tags, func(i, j int) bool {
		vA, vB := semver.New(tags[i].Name), semver.New(tags[j].Name)
		return vB.LessThan(*vA)
	})
	return tags[0].Name
}

var envRequired = []string{
	"REGISTRY_API",
	"REGISTRY_TOKEN",
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

func getProjects() RegistryProjects {
	url := os.Getenv("REGISTRY_API")

	spaceClient := http.Client{
		Timeout: time.Second * 2, // Timeout after 2 seconds
	}

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		log.Fatal(err)
	}

	req.Header.Set("PRIVATE-TOKEN", os.Getenv("REGISTRY_TOKEN"))

	res, getErr := spaceClient.Do(req)
	if getErr != nil {
		log.Fatal(getErr)
	}

	if res.Body != nil {
		defer res.Body.Close()
	}

	body, readErr := ioutil.ReadAll(res.Body)
	if readErr != nil {
		log.Fatal(readErr)
	}

	projects := RegistryProjects{}
	err = json.Unmarshal(body, &projects)
	if err != nil {
		log.Fatal(err)
	}

	return projects
}

func needUpgrade(a, b string) bool {
	vA, vB := semver.New(a), semver.New(b)
	return vA.Major == vB.Major && vA.Minor == vB.Minor && vB.Patch > vA.Patch
}

func main() {
	if err := checkEnv(); err != nil {
		log.Fatal(err)
	}

	// creates the in-cluster config
	config, err := rest.InClusterConfig()
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
	projects := getProjects()
	for _, item := range result.Items {
		log.Printf("Found %s with version %s\n",
			item.Spec.Type, item.Spec.Version)

		project, err := projects.get(item.Spec.Type)
		if err != nil {
			log.Fatal(err)
		}
		latest := project.getLatestTag()
		log.Printf("Latest tag was found %s", latest)

		if needUpgrade(item.Spec.Version, latest) {
			log.Printf("Version %s of %s should be upgraded onto %s\n",
				item.Spec.Version, item.Spec.Type, latest)

			updatedResult := kuberlogicv1.KuberLogicService{}
			item.Spec.Version = latest
			err = restClient.Put().
				Namespace("default").
				Resource("kuberlogics").
				Name(item.Name).
				Body(&item).
				Do(context.TODO()).
				Into(&updatedResult)
			if err != nil {
				log.Fatal(err.Error())
			}
			log.Println("KuberLogicService resource is successfully changed")

		} else {
			log.Println("No need upgrade")
		}
	}
}
