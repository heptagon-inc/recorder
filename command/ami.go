package command

import (
	"github.com/codegangsta/cli"
	"github.com/simplebackup/ec2"
	"github.com/visionmedia/go-cli-log"
	"github.com/youyo/go-reverse"
)

func CmdAmi(c *cli.Context) error {
	log.Info("info", "Start Program.")

	// set logger
	newLogging(c.GlobalBool("debug"))

	// get region and instance-id
	region, instanceID := setInstanceMetadata(
		c.Bool("self"),
		c.String("region"),
		c.String("instance-id"),
	)

	log.Info("create", "Start create AMI. [instance-id: %s]", instanceID)
	config := newConfig(region)
	log.Debug("simplebackupec2.NewService", "args: [config: %s]", config)
	s, err := simplebackupec2.NewService(config)
	if err != nil {
		log.Error(err)
		return err
	} else {
		log.Debug("simplebackupec2.NewService", "response: [service: %s]", s)
	}
	log.Debug("reverse.Bool", "args: [reboot: %t]", c.Bool("reboot"))
	noRebootOpt := reverse.Bool(c.Bool("reboot"))
	log.Debug("reverse.Bool", "response: [noRebootOpt: %t]", noRebootOpt)
	log.Debug("RegisterAMI", "args: [instanceID: %s, noRebootOpt: %t]", instanceID, noRebootOpt)
	imageID, err := s.RegisterAMI(instanceID, noRebootOpt)
	if err != nil {
		log.Error(err)
		return err
	} else {
		log.Debug("RegisterAMI", "response: [imageID: %s]", imageID)
		log.Info("create", "Create ami success. [instance-id: %s, image-id: %s]", instanceID, imageID)
	}
	log.Info("info", "End Program.")
	return nil
}
