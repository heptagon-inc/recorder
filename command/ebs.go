package command

import (
	"log"
	"os"
	"sort"

	"github.com/Sirupsen/logrus"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/codegangsta/cli"
)

type ebsOptions struct {
	self       bool
	profile    string
	instanceId string
	region     string
	lifeCycle  int
	json       bool
	logLevel   bool
}

type Snapshot struct {
	snapshotId string
	startTime  int64
}

type Snapshots []Snapshot

func CmdEbs(c *cli.Context) {
	// logging
	log.SetOutput(os.Stderr)

	// options
	o := ebsOptions{
		self:       c.Bool("self"),
		profile:    c.String("profile"),
		instanceId: c.String("instance-id"),
		region:     c.String("region"),
		lifeCycle:  c.Int("lifecycle"),
		json:       c.Bool("json"),
		logLevel:   c.Bool("debug"),
	}

	// set log format
	if o.json {
		logger.Formatter = new(logrus.JSONFormatter)
	}

	// set log level debug
	if o.logLevel {
		logger.Level = logrus.DebugLevel
	}

	// initialize
	logger.WithFields(logrus.Fields{"Action": "CmdEbsInitialize", "ebsOptions": o}).Debug()
	svc := svc{}

	// get region
	if o.self {
		svc.instance.region = *svc.instance.getRegion()
	} else {
		svc.instance.region = o.region
	}

	// get instance-id
	if o.self {
		svc.instance.instanceId = *svc.instance.getInstanceId()
	} else {
		svc.instance.instanceId = o.instanceId
	}
	logger.WithFields(logrus.Fields{
		"InstanceID": svc.instance.instanceId,
	}).Info("Execute recorder ebs")

	// AWS Auth
	svc.profile = o.profile
	svc.ec2 = svc.auth(svc.profile)

	// Describe Instance
	resp := svc.describeInstances(svc.instance.instanceId)

	// Has Name Key?
	var hasNameTag bool
	hasNameTag, svc.instance.nameTag = svc.hasNameTag(resp)

	// get all volume-id
	volumeIds := describeAllVolumeIds(resp)

	for _, volumeId := range volumeIds {

		// create snapshot
		snapshotId, err := svc.createSnapshot(svc.instance.instanceId, volumeId)
		if err == nil {
			logger.WithFields(logrus.Fields{
				"InstanceID": svc.instance.instanceId,
				"SnapshotID": snapshotId,
			}).Info("Created snapshot")

			// create name-tag if has it.
			if hasNameTag {
				m := svc.instance.nameTag + "'s Snapshot"
				svc.createNameTag(snapshotId, m)
			}
		}

		// snapshot lifecycle
		// get snapshots
		snapshots := svc.describeSnapshots(volumeId)

		// sort snapshots
		snapshotIds := sortSnapshots(snapshots)

		// If the number of snapshot is life-cycle or more, Delete snapshot.
		svc.deleteSnapshots(o.lifeCycle, snapshotIds)
	}
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

func describeAllVolumeIds(resp *ec2.DescribeInstancesOutput) []string {
	logger.WithFields(logrus.Fields{"Action": "describeAllVolumeIds"}).Debug()
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
		"Action":     "describeAllVolumeIds",
		"volume-ids": volumeIds,
	}).Debug()
	return volumeIds
}

func (svc svc) createSnapshot(i, v string) (string, error) {
	logger.WithFields(logrus.Fields{"Action": "createSnapshot"}).Debug()
	d := "Created by recorder from " + v + " of " + i
	params := &ec2.CreateSnapshotInput{
		VolumeId:    aws.String(v),
		Description: aws.String(d),
	}
	resp, err := svc.ec2.CreateSnapshot(params)
	errorLogging(err, "createSnapshot")
	logger.WithFields(logrus.Fields{
		"Action":      "createSnapshot",
		"InstanceID":  i,
		"VolumeId":    *resp.VolumeId,
		"SnapshotId":  *resp.SnapshotId,
		"Description": *resp.Description,
	}).Debug()
	return *resp.SnapshotId, err
}

func (svc svc) describeSnapshots(v string) *ec2.DescribeSnapshotsOutput {
	logger.WithFields(logrus.Fields{"Action": "describeSnapshots"}).Debug()
	params := &ec2.DescribeSnapshotsInput{
		Filters: []*ec2.Filter{
			{
				Name: aws.String("volume-id"),
				Values: []*string{
					aws.String(v),
				},
			},
		},
	}
	resp, err := svc.ec2.DescribeSnapshots(params)
	errorLogging(err, "describeSnapshots")
	return resp
}

func sortSnapshots(s *ec2.DescribeSnapshotsOutput) Snapshots {
	logger.WithFields(logrus.Fields{"Action": "sortSnapshots"}).Debug()
	var snapshotIds Snapshots = make([]Snapshot, 1)
	for _, res := range s.Snapshots {
		if snapshotIds[0].startTime == 0 {
			if isOwn(*res.Description) {
				snapshotIds[0].snapshotId = *res.SnapshotId
				snapshotIds[0].startTime = res.StartTime.Unix()
			}
		} else {
			if isOwn(*res.Description) {
				snapshotIds = append(snapshotIds, Snapshot{*res.SnapshotId, res.StartTime.Unix()})
			}
		}
	}
	sort.Sort(snapshotIds)
	logger.WithFields(logrus.Fields{
		"Action":      "sortSnapshots",
		"snapshotIds": snapshotIds,
	}).Debug()
	return snapshotIds
}

func (svc svc) deleteSnapshots(i int, s Snapshots) {
	logger.WithFields(logrus.Fields{"Action": "deleteSnapshots"}).Debug()
	for len(s) > i {
		for index, snapshots := range s {
			if index == 0 {
				// delete snapshot
				params := &ec2.DeleteSnapshotInput{
					SnapshotId: aws.String(snapshots.snapshotId),
				}
				_, err := svc.ec2.DeleteSnapshot(params)
				errorLogging(err, "deleteSnapshots")
				if err == nil {
					logger.WithFields(logrus.Fields{
						"InstanceID": svc.instance.instanceId,
						"SnapshotID": snapshots.snapshotId,
					}).Info("Deleted snapshot")
				}
			}
		}
		s = append(s[1:])
	}
}
