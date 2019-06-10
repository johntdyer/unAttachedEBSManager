package main

import (
	"os"

	// "v/github.com/aws/aws-sdk-go@v1.14.7/models/endpoints"
	runtime "github.com/banzaicloud/logrus-runtime-formatter"

	"github.com/sirupsen/logrus"

	// "github.com/aws/aws-lambda-go/lambda"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/external"
	// "github.com/aws/aws-sdk-go/aws/awserr"
	// "github.com/aws/aws-sdk-go/aws/session"
)

// Region ...
type Region struct {
	Name string `json:"region"`
}

// Counter ...
type Counter struct {
	VolumesFound   int
	VolumesDeleted int
}

// Application ...
type Application struct {
	Client  aws.Config
	Counter map[string]*Counter
	Regions []Region
	// Regions2 map[string]*Region
}

var (
	defaulSaveVolumeTag  = "CCE_Meta_dont_delete_when_unmounted"
	log                  = logrus.New()
	volumesDeleted       int
	volumesTagged        int
	ebsCleaner           = &Application{}
	totalWasted          float64
	futureSavingsPerYear float64
)

func init() {

	// Load AWS creds from env
	cfg, err := external.LoadDefaultAWSConfig()
	errorCheck(err)
	ebsCleaner.Client = cfg
	ebsCleaner.Client.Region = "us-east-1"
	childFormatter := logrus.TextFormatter{
		DisableColors: false,
		FullTimestamp: true,
	}
	runtimeFormatter := &runtime.Formatter{ChildFormatter: &childFormatter}
	log.Formatter = runtimeFormatter

	if os.Getenv("DEBUG_MODE_ENABLED") == "" {
		log.Level = logrus.InfoLevel

	} else {
		log.Level = logrus.DebugLevel
	}

	if os.Getenv("SAVE_VOLUME_TAG") != "" {
		defaulSaveVolumeTag = os.Getenv("SAVE_VOLUME_TAG")

		log.WithFields(logrus.Fields{
			"SAVE_VOLUME_TAG": os.Getenv("SAVE_VOLUME_TAG"),
		}).Info("Using tag from environment variable")

	} else {
		log.WithFields(logrus.Fields{
			"SAVE_VOLUME_TAG": os.Getenv("SAVE_VOLUME_TAG"),
		}).Info("Using default save environment variable")
	}

	log.Info("Application initialized")
}

func handler() error {

	// Get list of all regions
	r, err := ebsCleaner.getRegions(ebsCleaner.Client)
	errorCheck(err)

	ebsCleaner.Counter = make(map[string]*Counter)

	for _, region := range r {

		log.WithFields(logrus.Fields{
			"Region": region.Name,
		}).Info("Starting region")

		// Set region name
		ebsCleaner.Client.Region = region.Name

		// Get list of ebs volumes that are unattached
		volumes, err := listAvailableVolumes(ebsCleaner.Client)
		errorCheck(err)

		// Update structs
		ebsCleaner.Counter[region.Name] = &Counter{
			VolumesFound: len(volumes),
		}

		// Process volumes
		for _, v := range volumes {
			processVolume(v)
			ebsCleaner.Counter[region.Name].VolumesDeleted = ebsCleaner.Counter[region.Name].VolumesDeleted + 1
		}

		log.WithFields(logrus.Fields{
			"Region": region.Name,
		}).Info("Finished region")

	}
	return nil
}

func main() {

	lambda.Start(handler)

	var deleted, found int

	for k := range ebsCleaner.Counter {
		found = found + ebsCleaner.Counter[k].VolumesFound
		deleted = deleted + ebsCleaner.Counter[k].VolumesDeleted
	}

	log.WithFields(logrus.Fields{
		"VolumesDeleted":    deleted,
		"Found":             found,
		"totalWasted":       totalWasted,
		"FutureSavingsYear": futureSavingsPerYear,
	}).Info("Application finished")

}
