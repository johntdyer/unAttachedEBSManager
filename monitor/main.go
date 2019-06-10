package main

import (
	"os"

	runtime "github.com/banzaicloud/logrus-runtime-formatter"

	"github.com/sirupsen/logrus"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/external"
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
}

var (
	defaultSaveVolumeTag = "CCE_Meta_dont_delete_when_unmounted"
	noOpModeEnabled      bool
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

	// Set default region to start application
	ebsCleaner.Client.Region = "us-east-1"

	// Default log configs
	childFormatter := logrus.TextFormatter{
		DisableColors: false,
		FullTimestamp: true,
	}

	// Add runtime log formatter
	runtimeFormatter := &runtime.Formatter{ChildFormatter: &childFormatter}
	log.Formatter = runtimeFormatter

	// environment derived configuration
	log.Info("Application initalizing")
	// Set up logger
	if os.Getenv("APPLICATION_LOG_LEVEL") == "" {
		log.Info("Using default log level of INFO")
		log.Level = logrus.InfoLevel
	} else {
		log.Infof("Switching to log level %s", os.Getenv("APPLICATION_LOG_LEVEL"))
		log.Level, err = logrus.ParseLevel(os.Getenv("APPLICATION_LOG_LEVEL"))
		errorCheck(err)
	}

	// Set up no op mode
	if os.Getenv("NO_OP_MODE_TRUE") == "" {
		log.Info("No Op Mode not set, using default of false")
		noOpModeEnabled = false
	} else {
		if os.Getenv("NO_OP_MODE_TRUE") == "true" {
			log.Warn("No OP Mode enabled")
			noOpModeEnabled = true
		} else {
			log.Info("No OP Mode set to false via config")
			noOpModeEnabled = false
		}
	}

	// Setup save volume tag
	if os.Getenv("SAVE_VOLUME_TAG") != "" {
		defaultSaveVolumeTag = os.Getenv("SAVE_VOLUME_TAG")

		log.WithFields(logrus.Fields{
			"SAVE_VOLUME_TAG": os.Getenv("SAVE_VOLUME_TAG"),
		}).Info("Using tag from environment variable")

	} else {
		log.WithFields(logrus.Fields{
			"SAVE_VOLUME_TAG": os.Getenv("SAVE_VOLUME_TAG"),
		}).Info("Using default save environment variable")
	}

	log.Info("Application initalization completed")
}

// entry point into lambda function
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
		volumes, err := ebsCleaner.getAvailableVolumes()
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
