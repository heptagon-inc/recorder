package command

import (
	"io/ioutil"
	"net/http"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/simplebackup/ec2"
	"github.com/visionmedia/go-cli-log"
)

func newLogging(logLevelDebug bool) {
	if logLevelDebug {
		log.Verbose = true
	}
}

func setInstanceMetadata(self bool, region, instanceID string) (string, string) {
	log.Debug("setInstanceMetadata", "args: [self: %t, region: %s, instanceID: %s]", self, region, instanceID)
	return func(self bool) (string, string) {
		if self {
			r := getRegion()
			i := getInstanceID()
			log.Debug("setInstanceMetadata", "response: [region: %s, instance-id: %s]", r, i)
			return r, i
		} else {
			log.Debug("setInstanceMetadata", "response: [region: %s, instance-id: %s]", region, instanceID)
			return region, instanceID
		}
	}(self)
}

func getMetadata(resource string) string {
	log.Debug("getMetadata", "args: [resource: %s]", resource)
	url := "http://169.254.169.254/latest/meta-data/" + resource
	res, err := http.Get(url)
	defer res.Body.Close()
	if err != nil {
		log.Error(err)
		panic(err)
	}
	byteSlice, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.Error(err)
		panic(err)
	}
	response := string(byteSlice)
	log.Debug("getMetadata", "response: [response: %s]", response)
	return response
}

func getRegion() string {
	log.Debug("getRegion", "args: []")
	availabilityZone := getMetadata("placement/availability-zone")
	region := azToRegion(availabilityZone)
	log.Debug("getRegion", "response: [region: %s]", region)
	return region
}

func azToRegion(az string) string {
	log.Debug("azToRegion", "args: [AZ: %s]", az)
	region := string(az[:len(az)-1])
	log.Debug("azToRegion", "response: [region: %s]", region)
	return region
}

func getInstanceID() string {
	log.Debug("getInstanceID", "args: []")
	instanceID := getMetadata("instance-id")
	log.Debug("getInstanceID", "response: [instanceID: %s]", instanceID)
	return instanceID
}

func newConfig(region string) *aws.Config {
	log.Debug("newConfig", "args: [region: %s]", region)
	c := simplebackupec2.NewConfig().WithRegion(region)
	log.Debug("newConfig", "response: [aws.Config: %s]", c)
	return c
}
