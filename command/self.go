package command

import (
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"sort"
	"strings"

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

func (p Snapshots) Len() int {
	return len(p)
}

func (p Snapshots) Swap(i, j int) {
	p[i], p[j] = p[j], p[i]
}

func (p Snapshots) Less(i, j int) bool {
	return p[i].startTime < p[j].startTime
}

func CmdSelf(c *cli.Context) {
	// logging
	log.SetOutput(os.Stderr)
	logger := logrus.New()
	if c.Bool("json") {
		logger.Formatter = new(logrus.JSONFormatter)
	}

	// get region
	r_url := "http://169.254.169.254/latest/meta-data/local-hostname"
	r_res, err := http.Get(r_url)
	if err != nil {
		logger.Fatal(err)
	}
	r_b, err := ioutil.ReadAll(r_res.Body)
	if err != nil {
		logger.Fatal(err)
	}
	region := strings.Split(string(r_b), ".")[1]
	defer r_res.Body.Close()

	logger.WithFields(logrus.Fields{
		"region": region,
	}).Info("get region")

	// get instance-id
	i_url := "http://169.254.169.254/latest/meta-data/instance-id"
	i_res, err := http.Get(i_url)
	if err != nil {
		logger.Fatal(err)
	}
	i_b, err := ioutil.ReadAll(i_res.Body)
	if err != nil {
		logger.Fatal(err)
	}
	instance_id := string(i_b)
	defer i_res.Body.Close()

	logger.WithFields(logrus.Fields{
		"instance-id": instance_id,
	}).Info("get instance-id")

	// Auth
	svc := ec2.New(&aws.Config{Region: aws.String(region)})
	logger.Info("auth credential")

	// create instance-id-config
	params := &ec2.DescribeInstancesInput{
		InstanceIds: []*string{
			aws.String(instance_id),
		},
	}

	// get instance-info
	resp, err := svc.DescribeInstances(params)
	if err != nil {
		logger.Fatal(err)
	}
	logger.Info("get instance-info")

	// get all volume-id
	var volume_ids []string = make([]string, 1)
	for _, res := range resp.Reservations {
		for _, res := range res.Instances {
			for index, res := range res.BlockDeviceMappings {
				if index == 0 {
					volume_ids[0] = *res.Ebs.VolumeId
				} else {
					volume_ids = append(volume_ids, *res.Ebs.VolumeId)
				}
			}
		}
	}

	logger.WithFields(logrus.Fields{
		"volume-ids": volume_ids,
	}).Info("get volume-ids")

	for _, volume_id := range volume_ids {
		// Description
		snapshotDescription := "Created by recorder from " + volume_id + " of " + instance_id

		// create-snapshot config
		snapshotParams := &ec2.CreateSnapshotInput{
			VolumeId:    aws.String(volume_id),
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
			"InstanceID":  instance_id,
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
						aws.String(volume_id),
					},
				},
				{
					Name: aws.String("description"),
					Values: []*string{
						aws.String("Created by recorder from"),
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
		var snapshotId Snapshots = make([]Snapshot, 1)
		for index, res := range describeResp.Snapshots {
			if index == 0 {
				snapshotId[0].snapshotId = *res.SnapshotId
				snapshotId[0].startTime = res.StartTime.Unix()
			} else {
				snapshotId = append(snapshotId, Snapshot{*res.SnapshotId, res.StartTime.Unix()})
			}
		}
		// sort asc
		sort.Sort(snapshotId)
		// If the number of snapshot is life-cycle or more, Delete snapshot.
		logger.WithFields(logrus.Fields{
			"lifecycle": c.Int("lifecycle"),
			"snapshots": len(snapshotId),
		}).Info("life cycle settings")
		for len(snapshotId) > c.Int("lifecycle") {
			for index, snapshots := range snapshotId {
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
			snapshotId = append(snapshotId[1:])
		}
	}
}
