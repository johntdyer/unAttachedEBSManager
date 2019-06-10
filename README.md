# unAttachedEBSManager
Go Lambda function to managed unattched volumes to control waste

## Purpose

We had an issue were some volumes were orphaned in our accounts in various regions.  Unfortunitly these volumes added up over time and this tool was written to prevent such an occurance from happening again in the future.

## Features

* Adds up waste since volume was created based on us-east-1 pricing at the time of release
* Ability to run in no-op mode
* written in go and deployed in serverless so its 100% buzzword compliant !!

### Assumptions:
  * Since the pricing API  was difficult I assumed us-east-1 pricing for provisioned IOPS and EBS.

## TODO:

  * Actually pull pricing from the horrible pricing API, my gawd its horrible
  * Cloudwatch metrics

## Dependencies

* serverless 1.44.1
* serverless-pseudo-parameters plugin

### Installing serverless plugin

```
 npm install --save-dev serverless-pseudo-parameters
```

## Variables

There are two variables you may want to tune to set this up.

* The first one is the tag we will use to determine if a volume should not be culled. If this tag is present we will skip that volume.  The value of that tag can be anything, the important part is that it is set.

      `SAVE_VOLUME_TAG: "SomeRandomTag"`

* Run application in specified log level.  Valid options are `info, warn, critical, and debug`.  The default is info level.  This config is case insensitive.

      `APPLICATION_LOG_LEVEL: info`


* Run application in no-op mode.  This will log as if it were performing actions but in real life it will skip the actual delete phase of the run

      `NO_OP_MODE_TRUE: true`

## Deploying

Simply run make file

```bash
make deploy
```

Then you can test via

```bash
serverless invoke -f monitor
```

Then you can view logs by running

```bash
serverless logs -f monitor
time="2019-06-10T21:21:07Z" level=info msg="Application initalizing" function=0
time="2019-06-10T21:21:07Z" level=info msg="Switching to log level info" function=0
time="2019-06-10T21:21:07Z" level=info msg="No OP Mode set to false via config" function=0
time="2019-06-10T21:21:07Z" level=info msg="Using tag from environment variable" SAVE_VOLUME_TAG=CCE_Meta_dont_delete_when_unmounted function=0
time="2019-06-10T21:21:07Z" level=info msg="Application initalization completed" function=0
START RequestId: 9704baf0-32c7-48a8-b4b5-f183aa4b6f9c Version: $LATEST
time="2019-06-10T21:21:08Z" level=info msg="Starting region" Region=eu-north-1 function=handler
time="2019-06-10T21:21:08Z" level=info msg="Finished region" Region=eu-north-1 function=handler
time="2019-06-10T21:21:08Z" level=info msg="Starting region" Region=ap-south-1 function=handler
time="2019-06-10T21:21:09Z" level=info msg="Finished region" Region=ap-south-1 function=handler
time="2019-06-10T21:21:09Z" level=info msg="Starting region" Region=eu-west-3 function=handler
time="2019-06-10T21:21:09Z" level=info msg="Finished region" Region=eu-west-3 function=handler
time="2019-06-10T21:21:09Z" level=info msg="Starting region" Region=eu-west-2 function=handler
time="2019-06-10T21:21:10Z" level=info msg="Finished region" Region=eu-west-2 function=handler
time="2019-06-10T21:21:10Z" level=info msg="Starting region" Region=eu-west-1 function=handler
time="2019-06-10T21:21:10Z" level=info msg="Finished region" Region=eu-west-1 function=handler
time="2019-06-10T21:21:10Z" level=info msg="Starting region" Region=ap-northeast-2 function=handler
time="2019-06-10T21:21:11Z" level=info msg="Finished region" Region=ap-northeast-2 function=handler
time="2019-06-10T21:21:11Z" level=info msg="Starting region" Region=ap-northeast-1 function=handler
time="2019-06-10T21:21:12Z" level=info msg="Finished region" Region=ap-northeast-1 function=handler
time="2019-06-10T21:21:12Z" level=info msg="Starting region" Region=sa-east-1 function=handler
time="2019-06-10T21:21:12Z" level=info msg="Finished region" Region=sa-east-1 function=handler
time="2019-06-10T21:21:12Z" level=info msg="Starting region" Region=ca-central-1 function=handler
time="2019-06-10T21:21:12Z" level=info msg="Finished region" Region=ca-central-1 function=handler
time="2019-06-10T21:21:12Z" level=info msg="Starting region" Region=ap-southeast-1 function=handler
time="2019-06-10T21:21:13Z" level=info msg="Finished region" Region=ap-southeast-1 function=handler
time="2019-06-10T21:21:13Z" level=info msg="Starting region" Region=ap-southeast-2 function=handler
time="2019-06-10T21:21:14Z" level=info msg="Finished region" Region=ap-southeast-2 function=handler
time="2019-06-10T21:21:14Z" level=info msg="Starting region" Region=eu-central-1 function=handler
time="2019-06-10T21:21:14Z" level=info msg="Finished region" Region=eu-central-1 function=handler
time="2019-06-10T21:21:14Z" level=info msg="Starting region" Region=us-east-1 function=handler
time="2019-06-10T21:21:15Z" level=info msg="Skipping retained" CreateTime="2019-06-10 16:12:49.978 +0000 UTC" CreateTimeHuman="5 hours ago" DaysOld=0 VolumeID=vol-0ba098a83a519229f VolumeIops=300 VolumeSize=100 VolumeType=gp2 function=processVolume
time="2019-06-10T21:21:15Z" level=warning msg="Deleting Volume" CreateTime="2019-06-10 21:18:29.401 +0000 UTC" CreateTimeHuman="3 minutes ago" DaysOld=0 VolumeID=vol-0163026da42e5a4c0 VolumeIops=300 VolumeSize=100 VolumeType=gp2 function=deleteVolume futureSavingsPerYear=120 moneyWasted=0
time="2019-06-10T21:21:15Z" level=info msg="Finished region" Region=us-east-1 function=handler
time="2019-06-10T21:21:15Z" level=info msg="Starting region" Region=us-east-2 function=handler
time="2019-06-10T21:21:15Z" level=info msg="Finished region" Region=us-east-2 function=handler
time="2019-06-10T21:21:15Z" level=info msg="Starting region" Region=us-west-1 function=handler
time="2019-06-10T21:21:15Z" level=info msg="Finished region" Region=us-west-1 function=handler
time="2019-06-10T21:21:15Z" level=info msg="Starting region" Region=us-west-2 function=handler
time="2019-06-10T21:21:15Z" level=info msg="Finished region" Region=us-west-2 function=handler
END RequestId: 9704baf0-32c7-48a8-b4b5-f183aa4b6f9c
REPORT RequestId: 9704baf0-32c7-48a8-b4b5-f183aa4b6f9c	Duration: 7885.41 ms	Billed Duration: 7900 ms 	Memory Size: 1024 MB	Max Memory Used: 54 MB
```

