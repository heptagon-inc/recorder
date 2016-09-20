package command

import (
	"github.com/codegangsta/cli"
	"github.com/simplebackup/ec2"
	"github.com/visionmedia/go-cli-log"
)

func CmdEbs(c *cli.Context) error {
	log.Info("info", "Start Program.")

	// set logger
	newLogging(c.GlobalBool("debug"))

	// get region and instance-id
	region, instanceID := setInstanceMetadata(
		c.Bool("self"),
		c.String("region"),
		c.String("instance-id"),
	)

	log.Info("create", "Start create snapshot. [region: %s, instance-id: %s]", region, instanceID)
	config := newConfig(region)
	log.Debug("simplebackupec2.NewService", "args: [config: %s]", config)
	s, err := simplebackupec2.NewService(config)
	if err != nil {
		log.Error(err)
		return err
	} else {
		log.Debug("simplebackupec2.NewService", "response: [service: %s]", s)
	}
	log.Debug("CreateSnapshots", "args: [instance-id: %s]", instanceID)
	if err := s.CreateSnapshots(instanceID); err != nil {
		log.Error(err)
		return err
	} else {
		log.Debug("CreateSnapshots", "response: []")
		log.Info("create", "Create snapshot success. [instance-id: %s]", instanceID)
	}
	log.Info("delete", "Delete snapshots. [instance-id: %s, lifecycle: %d]", instanceID, c.Int("lifecycle"))
	log.Debug("RotateSnapshots", "args: [instance-id: %s, lifecycle: %d]", instanceID, c.Int("lifecycle"))
	if err := s.RotateSnapshots(instanceID, c.Int("lifecycle")); err != nil {
		log.Error(err)
		return err
	} else {
		log.Debug("RotateSnapshots", "response: []")
		log.Info("delete", "Delete snapshots success. [instance-id: %s, lifecycle: %d]", instanceID, c.Int("lifecycle"))
	}
	log.Info("info", "End Program.")
	return nil
}
