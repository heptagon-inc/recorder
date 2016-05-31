package command

import (
	"log"
	"os"
	"sort"
	"strconv"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/codegangsta/cli"
)

type amiOptions struct {
	self       bool
	profile    string
	instanceId string
	region     string
	lifecycle  int
	reboot     bool
	json       bool
	logLevel   bool
}

type image struct {
	imageId      string
	CreationDate int64
}

type images []image

func CmdAmi(c *cli.Context) error {
	// logging
	log.SetOutput(os.Stderr)

	// options
	o := amiOptions{
		self:       c.Bool("self"),
		profile:    c.String("profile"),
		instanceId: c.String("instance-id"),
		region:     c.String("region"),
		lifecycle:  c.Int("lifecycle"),
		reboot:     c.Bool("reboot"),
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

	// AWS Auth
	svc.profile = o.profile
	svc.ec2 = svc.auth(svc.profile)

	// Describe Instance
	resp := svc.describeInstances(svc.instance.instanceId)

	// Has Name Key?
	var hasNameTag bool
	hasNameTag, svc.instance.nameTag = svc.hasNameTag(resp)

	// create image
	noReboot := reverseRebootOption(o.reboot)
	imageId := svc.createImage(svc.instance.instanceId, noReboot)
	if hasNameTag {
		m := svc.instance.nameTag + "'s Image"
		svc.createNameTag(imageId, m)
	}

	// list ami
	images := svc.describeImages(svc.instance.instanceId)
	// sort ami
	imageIds := sortImages(images)
	// delete ami
	svc.deleteImages(o.lifecycle, imageIds)

	return nil
}

func (p images) Len() int {
	return len(p)
}

func (p images) Swap(i, j int) {
	p[i], p[j] = p[j], p[i]
}

func (p images) Less(i, j int) bool {
	return p[i].CreationDate < p[j].CreationDate
}

func (svc svc) createImage(i string, reboot bool) string {
	logger.WithFields(logrus.Fields{"Action": "createImage"}).Debug()
	d := "Created by recorder from " + i
	n := d + " at " + strconv.FormatInt(time.Now().Unix(), 10)
	params := &ec2.CreateImageInput{
		InstanceId:  aws.String(i),
		Name:        aws.String(n),
		Description: aws.String(d),
		NoReboot:    aws.Bool(true),
	}
	resp, err := svc.ec2.CreateImage(params)
	errorLogging(err, "createImage")
	logger.WithFields(logrus.Fields{
		"Action":      "createImage",
		"InstanceId":  i,
		"Description": d,
		"ImageId":     *resp.ImageId,
	}).Debug()
	return *resp.ImageId
}

func reverseRebootOption(b bool) bool {
	logger.WithFields(logrus.Fields{"Action": "reverseRebootOption"}).Debug()
	switch b {
	case true:
		logger.WithFields(logrus.Fields{
			"Action":   "reverseRebootOption",
			"args":     b,
			"response": false,
		}).Debug()
		return false
	case false:
		logger.WithFields(logrus.Fields{
			"Action":   "reverseRebootOption",
			"args":     b,
			"response": true,
		}).Debug()
		return true
	default:
		logger.WithFields(logrus.Fields{
			"Action":   "reverseRebootOption",
			"args":     b,
			"response": true,
		}).Debug()
		return true
	}
}

func (svc svc) describeImages(i string) *ec2.DescribeImagesOutput {
	logger.WithFields(logrus.Fields{"Action": "describeImages"}).Debug()
	params := &ec2.DescribeImagesInput{
		Filters: []*ec2.Filter{
			{
				Name: aws.String("name"),
				Values: []*string{
					aws.String("Created by recorder from " + i),
				},
			},
		},
	}
	resp, err := svc.ec2.DescribeImages(params)
	errorLogging(err, "describeImages")
	logger.WithFields(logrus.Fields{
		"Action": "describeImages",
		"resp":   resp,
	}).Debug()
	return resp
}

func sortImages(i *ec2.DescribeImagesOutput) images {
	logger.WithFields(logrus.Fields{"Action": "sortImages"}).Debug()
	var imageIds images = make([]image, 1)
	for _, res := range i.Images {
		if imageIds[0].CreationDate == 0 {
			if isOwn(*res.Description) {
				t, _ := time.Parse("Jan 2, 2006 at 15:04:05 PM UTC+9", *res.CreationDate)
				imageIds[0].imageId = *res.ImageId
				imageIds[0].CreationDate = t.Unix()
			}
		} else {
			if isOwn(*res.Description) {
				t, _ := time.Parse("Jan 2, 2006 at 15:04:05 PM UTC+9", *res.CreationDate)
				imageIds = append(imageIds, image{*res.ImageId, t.Unix()})
			}
		}
	}
	sort.Sort(imageIds)
	logger.WithFields(logrus.Fields{
		"Action":   "sortImages",
		"imageids": imageIds,
	}).Debug()
	return imageIds
}

func (svc svc) deleteImages(l int, i images) {
	logger.WithFields(logrus.Fields{"Action": "deleteImages"}).Debug()
	for len(i) > l {
		for index, images := range i {
			if index == 0 {
				params := &ec2.DeregisterImageInput{
					ImageId: aws.String(images.imageId),
				}
				resp, err := svc.ec2.DeregisterImage(params)
				errorLogging(err, "deleteImages")
				logger.WithFields(logrus.Fields{
					"Action": "deleteImages",
					"params": params,
					"resp":   resp,
				}).Debug()
			}
		}
		i = append(i[1:])
	}
}
