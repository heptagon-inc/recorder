# recorder

[![CircleCI](https://circleci.com/gh/heptagon-inc/recorder.svg?style=svg)](https://circleci.com/gh/heptagon-inc/recorder)

## Description

Create a snapshot of the Amazon EC2.

## Usage

- Set AWS configuration file.

```
cat ~/.aws/credentials
[default]
aws_access_key_id = XXXXXXXXXX
aws_secret_access_key = XXXXXXXXXXXXXXX
```
- Required IAM Policy.

```
{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Effect": "Allow",
            "Action": [
                "ec2:CreateSnapshot",
                "ec2:CreateTags",
                "ec2:DeleteSnapshot",
                "ec2:DescribeSnapshots",
                "ec2:DescribeInstances",
                "ec2:CreateImage",
                "ec2:DeregisterImage",
                "ec2:DescribeImages"
            ],
            "Resource": [
                "*"
            ]
        }
    ]
}
```

- Run

Execute `recorder ebs` or `recorder ami` .  
But, `recorder ami` is buggy. (use not reccomend.)

```
$ recorder -h
NAME:
   recorder - Create a snapshot of the Amazon EC2.

USAGE:
   recorder [global options] command [command options] [arguments...]

VERSION:
   0.3.0

AUTHOR(S):
   youyo

COMMANDS:
   ebs		Snapshotting for specific instance's volume.
   ami		Creating Image for specific instance
   help, h	Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --help, -h		show help
   --version, -v	print the version
```

```
$ recorder ebs -h
NAME:
   ebs - Snapshotting for specific instance's volume.

USAGE:
   command ebs [command options] [arguments...]

OPTIONS:
   --self, -s				Snapshotting for own volume.
   --profile, -p 			Set AWS-Credentials profile name.
   --instance-id, -i "i-xxxxxxx"	Set InstanceId. If '--self' option is set, this is ignored.
   --region, -r "ap-northeast-1"	Set Region. If '--self' option is set, this is ignored.
   --lifecycle, -l "5"			Set the number of life cycle for snapshot.
   --json, -j				Log Format json.
   --debug				Set LogLevel Debug.
```

```
$ recorder ami -h
NAME:
   ami - Creating Image for specific instance

USAGE:
   command ami [command options] [arguments...]

OPTIONS:
   --self, -s				Creating Image for own.
   --profile, -p 			Set AWS-Credentials profile name.
   --instance-id, -i "i-xxxxxxx"	Set InstanceId. If '--self' option is set, this is ignored.
   --region, -r "ap-northeast-1"	Set Region. If '--self' option is set, this is ignored.
   --lifecycle, -l "5"			Set the number of life cycle for AMI
   --reboot				Reboot instance when create image.
   --json, -j				Log Format json.
   --debug				Set LogLevel Debug.
```


## Install

Download binary-file.

```
$ wget https://github.com/heptagon-inc/recorder/releases/download/v0.3.0/recorder_linux_amd64.zip
$ unzip recorder_linux_amd64.zip
$ mv recorder /usr/local/bin/
$ chmod 755 /usr/local/bin/recorder
```

Or to install, use `go get`:

```bash
$ go get -d github.com/heptagon-inc/recorder
$ cd $GOPATH/src/github.com/heptagon-inc/recorder
$ go build
```

## Warning!

Test does not exist yet.

## Contribution

1. Fork ([https://github.com/heptagon-inc/recorder/fork](https://github.com/heptagon-inc/recorder/fork))
1. Create a feature branch
1. Commit your changes
1. Rebase your local changes against the master branch
1. Run test suite with the `go test ./...` command and confirm that it passes
1. Run `gofmt -s`
1. Create a new Pull Request

## Author

[youyo](https://github.com/youyo)
