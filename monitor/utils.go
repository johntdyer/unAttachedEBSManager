package main

import (
	"context"
	"math"
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/nleeper/goment"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

var data = map[string]map[string]float64{
	"standard": map[string]float64{},
	"gp2":      map[string]float64{},
	"io1":      map[string]float64{},
	"sc1":      map[string]float64{},
	"st1":      map[string]float64{},
}

func (aws Application) getRegions(cfg aws.Config) ([]Region, error) {
	svc := ec2.New(cfg)
	req := svc.DescribeRegionsRequest(&ec2.DescribeRegionsInput{})
	regions, err := req.Send(context.Background())
	if err != nil {
		return []Region{}, err
	}
	listOfRegions := make([]Region, 0, len(regions.Regions))
	for _, region := range regions.Regions {
		listOfRegions = append(listOfRegions, Region{
			Name: *region.RegionName,
		})
	}
	return listOfRegions, nil
}

func price(v ec2.Volume) (float64, float64) {
	data["standard"]["price_per_gb"] = 0.10
	data["gp2"]["price_per_gb"] = 0.10
	data["io1"]["price_per_gb"] = 0.125
	data["io1"]["price_per_iop"] = 0.065
	data["sc1"]["price_per_gb"] = 0.025
	data["st1"]["price_per_gb"] = 0.045

	_, daysOld, _ := doDateMath(v)

	sizeValue := aws.Int64Value(v.Size)
	futureCostPerMonth := data[string(v.VolumeType)]["price_per_gb"] * float64(sizeValue)
	months := (float64(daysOld) / 30.0)

	moneyWasted := futureCostPerMonth * months
	if v.VolumeType == "io1" {
		iopsCost := (float64(aws.Int64Value(v.Iops)) * data["io1"]["price_per_iop"])
		futureCostPerMonth = futureCostPerMonth + iopsCost
		moneyWasted = moneyWasted + (iopsCost * months)
	}

	return futureCostPerMonth, moneyWasted

}

func listAvailableVolumes(cfg aws.Config) ([]ec2.Volume, error) {

	svc := ec2.New(cfg)
	req := svc.DescribeVolumesRequest(&ec2.DescribeVolumesInput{
		Filters: []ec2.Filter{
			ec2.Filter{
				Name:   aws.String("status"),
				Values: []string{"available"},
			},
		},
	})
	res, err := req.Send(context.Background())
	if err != nil {
		s := make([]ec2.Volume, 0)
		return s, err
	}

	return res.Volumes, nil
}

func processVolume(volume ec2.Volume) {
	var iops int64

	if volume.VolumeType == "standard" {
		iops = 0
	} else {
		iops = *volume.Iops
	}
	DaysHuman, daysOld, _ := doDateMath(volume)
	futureCostPerMonth, moneyWasted := price(volume)
	totalWasted = totalWasted + math.Round(moneyWasted)
	futureSavingsPerYear = futureSavingsPerYear + math.Round(futureCostPerMonth*12)

	// Check if tag is present to indicate we want to skip this volume
	if checkIfSkipBasedOnTag(volume.Tags, defaulSaveVolumeTag) {

		log.WithFields(logrus.Fields{
			"VolumeID":        *volume.VolumeId,
			"VolumeType":      volume.VolumeType,
			"VolumeSize":      strconv.FormatInt(*volume.Size, 10),
			"VolumeIops":      iops,
			"CreateTime":      *volume.CreateTime,
			"CreateTimeHuman": DaysHuman,
			"DaysOld":         daysOld,
		}).Info("Skipping retained")
	} else {

		logEventTags := logrus.Fields{
			"VolumeID":             *volume.VolumeId,
			"VolumeType":           volume.VolumeType,
			"VolumeSize":           strconv.FormatInt(*volume.Size, 10),
			"VolumeIops":           iops,
			"CreateTime":           *volume.CreateTime,
			"CreateTimeHuman":      DaysHuman,
			"DaysOld":              daysOld,
			"moneyWasted":          math.Round(moneyWasted),
			"futureSavingsPerYear": math.Round(futureCostPerMonth * 12),
		}

		ebsCleaner.deleteVolume(*volume.VolumeId, logEventTags)

	}
}

// search tags to see if our skipTag exists
func checkIfSkipBasedOnTag(volume []ec2.Tag, tagName string) bool {
	//
	for _, t := range volume {
		if *t.Key == tagName {
			if *t.Value == "true" {
				return true
			}
		}
	}
	return false

}

// Do date math to get age, human readable realative age, and days old
func doDateMath(v ec2.Volume) (string, int, error) {
	g, err := goment.New(*v.CreateTime)
	errorCheck(err)

	date := time.Now()
	diff := date.Sub(g.ToTime())

	daysOld := int(diff.Hours() / 24)

	return g.FromNow(), daysOld, nil
}

// Generic error handler
func errorCheck(err error) {
	if err != nil {
		log.Errorf("%+v\n", errors.Wrap(err, ""))
	}
}

func (aws Application) deleteVolume(volumeID string, fieldTags logrus.Fields) {
	// func deleteVolume(volumeID string, fieldTags logrus.Fields) {
	log.WithFields(fieldTags).Warn("Deleting Volume")
	// svc := ec2.New(cfg)
	// req := svc.DeleteVolumeRequest(&ec2.DeleteVolumeInput{
	// 	VolumeId: &volumeID,
	// })
	// _, err := req.Send(context.Background())

	// return err
}

// func (aws Application) getPrice(volume ec2.Volume) float64 {
// 	ebsPriceList := &pricing.EbsPriceList{}
// 	pricing.GetEbsPriceList(aws.Client.Region, ebsPriceList)
// 	// Get cache
// 	if aws.Regions2[aws.Client.Region] == nil {
// 		aws.Regions2[aws.Client.Region].Prices = *ebsPriceList
// 		log.Errorf("%s - Not Found", aws.Client.Region)

// 		pricing.GetEbsPriceList(aws.Client.Region, ebsPriceList)
// 		// println(ebsPriceList.Prices[0].Price.USD)
// 	} else {
// 		log.Errorf("%s - Found", aws.Client.Region)
// 		return 0.0
// 	}
// 	return 0.0
// }

// func (aws Application) setCache() error {
// 	ebsPriceList := &pricing.EbsPriceList{}
// 	pricing.GetEbsPriceList(aws.Client.Region, ebsPriceList)

// 	if aws.Regions2[aws.Client.Region] == nil {
// 		aws.Regions2[aws.Client.Region].Prices = *ebsPriceList
// 		log.Errorf("%s - Not Found", aws.Client.Region)

// 		pricing.GetEbsPriceList(aws.Client.Region, ebsPriceList)
// 		// println(ebsPriceList.Prices[0].Price.USD)
// 	} else {
// 		log.Errorf("%s - Found", aws.Client.Region)
// 	}

// if aws.PriceList
// for i := range aws.PriceList {
// 	if aws.PriceList[i].Name == region {
// 		fmt.Printf("%d exists", i)
// 		return nil
// 	}
// }

// 	return nil
// }
