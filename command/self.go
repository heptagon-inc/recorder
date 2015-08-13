package command

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"sort"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/awsutil"
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
	// get region
	r_url := "http://169.254.169.254/latest/meta-data/local-hostname"
	r_res, _ := http.Get(r_url)
	r_b, _ := ioutil.ReadAll(r_res.Body)
	defer r_res.Body.Close()
	region := strings.Split(string(r_b), ".")[1]

	// get instance-id
	url := "http://169.254.169.254/latest/meta-data/instance-id"
	res, _ := http.Get(url)
	b, _ := ioutil.ReadAll(res.Body)
	defer res.Body.Close()
	instance_id := string(b)

	// Auth
	svc := ec2.New(&aws.Config{Region: aws.String(region)})

	// create instance-id-config
	params := &ec2.DescribeInstancesInput{
		InstanceIDs: []*string{
			aws.String(instance_id),
		},
	}

	// get instance-info
	resp, err := svc.DescribeInstances(params)
	if err != nil {
		panic(err)
	}

	// get all volume-id
	var volume_ids []string = make([]string, 1)
	for _, res := range resp.Reservations {
		for _, res := range res.Instances {
			for index, res := range res.BlockDeviceMappings {
				if index == 0 {
					volume_ids[0] = *res.EBS.VolumeID
				} else {
					volume_ids = append(volume_ids, *res.EBS.VolumeID)
				}
			}
		}
	}

	for _, volume_id := range volume_ids {
		// Description
		snapshotDescription := "Created by recorder from " + volume_id + " of " + instance_id

		// create-snapshot config
		snapshotParams := &ec2.CreateSnapshotInput{
			VolumeID:    aws.String(volume_id),
			Description: aws.String(snapshotDescription),
		}

		// create snapshot
		snapshotResp, err := svc.CreateSnapshot(snapshotParams)
		if err != nil {
			if awsErr, ok := err.(awserr.Error); ok {
				fmt.Println(awsErr.Code(), awsErr.Message(), awsErr.OrigErr())
				if reqErr, ok := err.(awserr.RequestFailure); ok {
					fmt.Println(reqErr.Code(), reqErr.Message(), reqErr.StatusCode(), reqErr.RequestID())
				}
			} else {
				fmt.Println(err.Error())
			}
		}
		fmt.Println(awsutil.Prettify(snapshotResp))

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
			},
		}
		// get snapshots
		describeResp, err := svc.DescribeSnapshots(DescribeSnapshotParams)
		if err != nil {
			if awsErr, ok := err.(awserr.Error); ok {
				fmt.Println(awsErr.Code(), awsErr.Message(), awsErr.OrigErr())
				if reqErr, ok := err.(awserr.RequestFailure); ok {
					fmt.Println(reqErr.Code(), reqErr.Message(), reqErr.StatusCode(), reqErr.RequestID())
				}
			} else {
				fmt.Println(err.Error())
			}
		}

		// linkage of snapshot-id and the start-time
		var snapshotId Snapshots = make([]Snapshot, 1)
		for index, res := range describeResp.Snapshots {
			if index == 0 {
				snapshotId[0].snapshotId = *res.SnapshotID
				snapshotId[0].startTime = res.StartTime.Unix()
			} else {
				snapshotId = append(snapshotId, Snapshot{*res.SnapshotID, res.StartTime.Unix()})
			}
		}
		// sort asc
		sort.Sort(snapshotId)
		// If the number of snapshot is life-cycle or more, Delete snapshot.
		for len(snapshotId) > c.Int("lifecycle") {
			for index, snapshots := range snapshotId {
				if index == 0 {
					// delete snapshot
					deleteParam := &ec2.DeleteSnapshotInput{
						SnapshotID: aws.String(snapshots.snapshotId),
					}
					_, err := svc.DeleteSnapshot(deleteParam)
					if err != nil {
						if awsErr, ok := err.(awserr.Error); ok {
							fmt.Println(awsErr.Code(), awsErr.Message(), awsErr.OrigErr())
							if reqErr, ok := err.(awserr.RequestFailure); ok {
								fmt.Println(reqErr.Code(), reqErr.Message(), reqErr.StatusCode(), reqErr.RequestID())
							}
						} else {
							fmt.Println(err.Error())
						}
					}
					deleteMessage := "deleted: " + snapshots.snapshotId
					fmt.Println(deleteMessage)
				}
			}
			// delete snapshotID
			snapshotId = append(snapshotId[1:])
		}
	}
}
