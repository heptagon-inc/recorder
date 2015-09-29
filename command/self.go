package command

import (
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"sort"
	"strings"
	"regexp"

	"github.com/Sirupsen/logrus"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/codegangsta/cli"
)


type Snapshot struct {
	snapshotId string
	startTime  int64
}

type Snapshots []Snapshot

type Http struct {
	Url string
	Byte []byte
}

type Region struct {
	Http
	Region string
}

type InstanceId struct {
	Http
	InstanceId string
}

func (p Snapshots) Len() int {
	return len(p)
}

func (p Snapshots) Swap(i, j int) {
	p[i], p[j] = p[j], p[i]
}

func (p Snapshots) Less(i, j int) bool {
	return p[i].startTime < p[j].startTime
}

func getRegion() *string {
	r := Region{}
	r.Url = "http://169.254.169.254/latest/meta-data/local-hostname"
	res, err := http.Get(r.Url)
	if err != nil {
		logger.Fatal(err)
	}
	r.Byte, err = ioutil.ReadAll(res.Body)
	if err != nil {
		logger.Fatal(err)
	}
	r.Region = strings.Split(string(r.Byte), ".")[1]
	defer res.Body.Close()
	logger.WithFields(logrus.Fields{
		"region": r.Region,
	}).Info("get region")
	return &r.Region
}

func getInstanceId() *string {
	i := InstanceId{}
	i.Url = "http://169.254.169.254/latest/meta-data/instance-id"
	res, err := http.Get(i.Url)
	if err != nil {
		logger.Fatal(err)
	}
	i.Byte, err = ioutil.ReadAll(res.Body)
	if err != nil {
		logger.Fatal(err)
	}
	i.InstanceId = string(i.Byte)
	defer res.Body.Close()
	logger.WithFields(logrus.Fields{
		"instance-id": i.InstanceId,
	}).Info("get instance-id")
	return &i.InstanceId
}

func isOwnSnapshot(description string) (b bool) {
	if m, _ := regexp.MatchString("Created by recorder from.*", description); !m {
		return false
	}
	return true
}

var logger = logrus.New()

func CmdSelf(c *cli.Context) {
	// logging
	log.SetOutput(os.Stderr)
	if c.Bool("json") {
		logger.Formatter = new(logrus.JSONFormatter)
	}

	// get region
	region := getRegion()

	// get instance-id
	instanceId := getInstanceId()

	// AWS Auth
	svc := ec2.New(&aws.Config{Region: aws.String(*region)})
	logger.Info("auth credential")

	// get instance-info
	params := &ec2.DescribeInstancesInput{
		InstanceIds: []*string{
			aws.String(*instanceId),
		},
	}
	resp, err := svc.DescribeInstances(params)
	if err != nil {
		logger.Fatal(err)
	}
	logger.Info("get instance-info")

	// get all volume-id
	var volumeIds []string = make([]string, 1)
	for _, res := range resp.Reservations {
		for _, res := range res.Instances {
			for index, res := range res.BlockDeviceMappings {
				if index == 0 {
					volumeIds[0] = *res.Ebs.VolumeId
				} else {
					volumeIds = append(volumeIds, *res.Ebs.VolumeId)
				}
			}
		}
	}
	logger.WithFields(logrus.Fields{
		"volume-ids": volumeIds,
	}).Info("get volume-ids")

	for _, volumeId := range volumeIds {
		// Description
		snapshotDescription := "Created by recorder from " + volumeId + " of " + *instanceId

		// create-snapshot config
		snapshotParams := &ec2.CreateSnapshotInput{
			VolumeId:    aws.String(volumeId),
			Description: aws.String(snapshotDescription),
		}

		// create snapshot
		snapshotResp, err := svc.CreateSnapshot(snapshotParams)
		if err != nil {
			if awsErr, ok := err.(awserr.Error); ok {
				logger.WithFields(logrus.Fields{
					"awsErr.Code":    awsErr.Code(),
					"awsErr.Message": awsErr.Message(),
					"awsErr.OrigErr": awsErr.OrigErr(),
				}).Error("create snapshot")
				if reqErr, ok := err.(awserr.RequestFailure); ok {
					logger.WithFields(logrus.Fields{
						"reqErr.Code":       reqErr.Code(),
						"reqErr.Message":    reqErr.Message(),
						"reqErr.StatusCode": reqErr.StatusCode(),
						"reqErr.RequestID":  reqErr.RequestID(),
					}).Error("create snapshot")
				}
			} else {
				logger.WithFields(logrus.Fields{
					"err.Error": err.Error(),
				}).Error("create snapshot")
			}
			logger.Fatal("create snapshot")
		}
		logger.WithFields(logrus.Fields{
			"InstanceID":  *instanceId,
			"VolumeId":    *snapshotResp.VolumeId,
			"SnapshotId":  *snapshotResp.SnapshotId,
			"Description": *snapshotResp.Description,
		}).Info("create snapshot")

		// snapshot lifecycle
		// describe-snapshot config
		DescribeSnapshotParams := &ec2.DescribeSnapshotsInput{
			Filters: []*ec2.Filter{
				{
					Name: aws.String("volume-id"),
					Values: []*string{
						aws.String(volumeId),
					},
				},
			},
		}
		// get snapshots
		describeResp, err := svc.DescribeSnapshots(DescribeSnapshotParams)
		if err != nil {
			if awsErr, ok := err.(awserr.Error); ok {
				logger.WithFields(logrus.Fields{
					"awsErr.Code":    awsErr.Code(),
					"awsErr.Message": awsErr.Message(),
					"awsErr.OrigErr": awsErr.OrigErr(),
				}).Error("describe snapshot")
				if reqErr, ok := err.(awserr.RequestFailure); ok {
					logger.WithFields(logrus.Fields{
						"reqErr.Code":       reqErr.Code(),
						"reqErr.Message":    reqErr.Message(),
						"reqErr.StatusCode": reqErr.StatusCode(),
						"reqErr.RequestID":  reqErr.RequestID(),
					}).Error("describe snapshot")
				}
			} else {
				logger.WithFields(logrus.Fields{
					"err.Error": err.Error(),
				}).Error("describe snapshot")
			}
			logger.Fatal("describe snapshot")
		}
		logger.Info("describe snapshot")

		// linkage of snapshot-id and the start-time
		var snapshotIds Snapshots = make([]Snapshot, 1)
		for _, res := range describeResp.Snapshots {
			if snapshotIds[0].startTime == 0 {
				if isOwnSnapshot(*res.Description) {
					snapshotIds[0].snapshotId = *res.SnapshotId
					snapshotIds[0].startTime = res.StartTime.Unix()
				}
			} else {
				if isOwnSnapshot(*res.Description) {
					snapshotIds = append(snapshotIds, Snapshot{*res.SnapshotId, res.StartTime.Unix()})
				}
			}
		}
		// sort asc
		sort.Sort(snapshotIds)
		// If the number of snapshot is life-cycle or more, Delete snapshot.
		logger.WithFields(logrus.Fields{
			"lifecycle": c.Int("lifecycle"),
			"snapshots": len(snapshotIds),
		}).Info("life cycle settings")
		for len(snapshotIds) > c.Int("lifecycle") {
			for index, snapshots := range snapshotIds {
				if index == 0 {
					// delete snapshot
					deleteParam := &ec2.DeleteSnapshotInput{
						SnapshotId: aws.String(snapshots.snapshotId),
					}
					_, err := svc.DeleteSnapshot(deleteParam)
					if err != nil {
						if awsErr, ok := err.(awserr.Error); ok {
							logger.WithFields(logrus.Fields{
								"SnapshotID":     snapshots.snapshotId,
								"awsErr.Code":    awsErr.Code(),
								"awsErr.Message": awsErr.Message(),
								"awsErr.OrigErr": awsErr.OrigErr(),
							}).Error("delete snapshot")
							if reqErr, ok := err.(awserr.RequestFailure); ok {
								logger.WithFields(logrus.Fields{
									"SnapshotID":        snapshots.snapshotId,
									"reqErr.Code":       reqErr.Code(),
									"reqErr.Message":    reqErr.Message(),
									"reqErr.StatusCode": reqErr.StatusCode(),
									"reqErr.RequestID":  reqErr.RequestID(),
								}).Error("delete snapshot")
							}
						} else {
							logger.WithFields(logrus.Fields{
								"SnapshotID": snapshots.snapshotId,
								"err.Error":  err.Error(),
							}).Error("delete snapshot")
						}
						logger.Fatal("delete snapshot")
					}
					logger.WithFields(logrus.Fields{
						"SnapshotID": snapshots.snapshotId,
					}).Info("delete snapshot")
				}
			}
			// remove snapshotID
			snapshotIds = append(snapshotIds[1:])
		}
	}
}
