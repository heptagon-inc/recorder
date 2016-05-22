package command

import (
	"io/ioutil"
	"net/http"
	"regexp"

	"github.com/Sirupsen/logrus"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
)

var logger = logrus.New()

type metadata struct {
	region           string
	availabilityZone string
	instanceId       string
	nameTag          string
}

type svc struct {
	ec2      *ec2.EC2
	instance metadata
	profile  string
}

func (svc svc) auth(p string) *ec2.EC2 {
	logger.WithFields(logrus.Fields{"Action": "auth"}).Debug()
	svc.profile = p
	logger.WithFields(logrus.Fields{
		"Action":  "auth",
		"profile": svc.profile,
	}).Debug()
	c := func(p string) aws.Config {
		if p == "" {
			return aws.Config{
				Region: aws.String(svc.instance.region),
			}
		} else {
			return aws.Config{
				Credentials: credentials.NewSharedCredentials("", p),
				Region:      aws.String(svc.instance.region),
			}
		}
	}(svc.profile)
	svc.ec2 = ec2.New(session.New(), &c)
	return svc.ec2
}

func (m metadata) getMetadata(s string) *string {
	logger.WithFields(logrus.Fields{"Action": "getMetadata"}).Debug()
	url := "http://169.254.169.254/latest/meta-data/" + s
	res, err := http.Get(url)
	if err != nil {
		logger.WithFields(logrus.Fields{
			"Action": "getMetadata.httpGet",
			"res":    res,
			"url":    url,
		}).Fatal(err)
	}
	byteSlice, err := ioutil.ReadAll(res.Body)
	if err != nil {
		logger.WithFields(logrus.Fields{
			"Action":    "getMetadata.ioutil.ReadAll",
			"byteSlice": byteSlice,
			"res.Body":  res.Body,
		}).Fatal(err)
	}
	defer res.Body.Close()
	response := string(byteSlice)
	logger.WithFields(logrus.Fields{
		"Action":   "getMetadata",
		"url":      url,
		"response": response,
	}).Debug()
	return &response
}

func (m metadata) getRegion() *string {
	logger.WithFields(logrus.Fields{"Action": "getRegion"}).Debug()
	m.availabilityZone = *m.getMetadata("placement/availability-zone")
	m.region = string(m.availabilityZone[:len(m.availabilityZone)-1])
	logger.WithFields(logrus.Fields{
		"Action": "getRegion",
		"region": m.region,
	}).Debug()
	return &m.region
}

func (m metadata) getInstanceId() *string {
	logger.WithFields(logrus.Fields{"Action": "getInstanceId"}).Debug()
	m.instanceId = *m.getMetadata("instance-id")
	logger.WithFields(logrus.Fields{
		"Action":      "getInstanceId",
		"instance-id": m.instanceId,
	}).Debug()
	return &m.instanceId
}

func (svc svc) describeInstances(i string) *ec2.DescribeInstancesOutput {
	logger.WithFields(logrus.Fields{"Action": "describeInstances"}).Debug()
	params := &ec2.DescribeInstancesInput{
		InstanceIds: []*string{
			aws.String(i),
		},
	}
	resp, err := svc.ec2.DescribeInstances(params)
	if err != nil {
		logger.WithFields(logrus.Fields{
			"Action":      "describeInstances",
			"instance-id": i,
			"params":      params,
			"resp":        resp,
		}).Fatal(err)
	}
	logger.WithFields(logrus.Fields{
		"Action":      "describeInstances",
		"instance-id": i,
		"resp":        resp,
	}).Debug()
	return resp
}

func (svc svc) hasNameTag(resp *ec2.DescribeInstancesOutput) (bool, string) {
	logger.WithFields(logrus.Fields{"Action": "hasNameTag"}).Debug()
	for _, res := range resp.Reservations {
		for _, res := range res.Instances {
			for _, res := range res.Tags {
				if *res.Key == "Name" && *res.Value != "" {
					logger.WithFields(logrus.Fields{
						"Action": "hasNameTag",
						"bool":   true,
						"value":  *res.Value,
					}).Debug()
					return true, *res.Value
				}
			}
		}
	}
	logger.WithFields(logrus.Fields{
		"Action": "hasNameTag",
		"bool":   false,
		"value":  "",
	}).Debug()
	return false, ""
}

func (svc svc) createNameTag(snapshotId, message string) {
	logger.WithFields(logrus.Fields{"Action": "createNameTag"}).Debug()
	params := &ec2.CreateTagsInput{
		Resources: []*string{
			aws.String(snapshotId),
		},
		Tags: []*ec2.Tag{
			{
				Key:   aws.String("Name"),
				Value: aws.String(message),
			},
		},
	}
	_, err := svc.ec2.CreateTags(params)
	if err != nil {
		logger.WithFields(logrus.Fields{
			"Action":     "createNameTag",
			"snapshotId": snapshotId,
			"message":    message,
		}).Fatal(err)
	}
	logger.WithFields(logrus.Fields{
		"Action":     "createNameTag",
		"snapshotId": snapshotId,
		"message":    message,
	}).Debug()
}

func errorLogging(err error, m string) {
	if err != nil {
		if awsErr, ok := err.(awserr.Error); ok {
			logger.WithFields(logrus.Fields{
				"awsErr.Code":    awsErr.Code(),
				"awsErr.Message": awsErr.Message(),
				"awsErr.OrigErr": awsErr.OrigErr(),
			}).Error(m)
			if reqErr, ok := err.(awserr.RequestFailure); ok {
				logger.WithFields(logrus.Fields{
					"reqErr.Code":       reqErr.Code(),
					"reqErr.Message":    reqErr.Message(),
					"reqErr.StatusCode": reqErr.StatusCode(),
					"reqErr.RequestID":  reqErr.RequestID(),
				}).Error(m)
			}
		} else {
			logger.WithFields(logrus.Fields{
				"err.Error": err.Error(),
			}).Error(m)
		}
		logger.Fatal(m)
	}
}

func isOwn(description string) (b bool) {
	logger.WithFields(logrus.Fields{"Action": "isOwn"}).Debug()
	if m, _ := regexp.MatchString("Created by recorder from.*", description); !m {
		logger.WithFields(logrus.Fields{
			"Action":      "isOwn",
			"description": description,
			"bool":        "false",
		}).Debug()
		return false
	}
	logger.WithFields(logrus.Fields{
		"Action":      "isOwn",
		"description": description,
		"response":    true,
	}).Debug()
	return true
}
