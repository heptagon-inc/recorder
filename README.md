# recorder

[![wercker status](https://app.wercker.com/status/5b8b5d5b3765a30a21226888642e47cf/s/master "wercker status")](https://app.wercker.com/project/byKey/5b8b5d5b3765a30a21226888642e47cf)

## Description

Create EBS-Snapshot and AMI of the Amazon EC2.

## Usage

Execute `recorder ebs` or `recorder ami` .

```
$ recorder ebs -h
NAME:
   recorder ebs - Snapshotting for specific instance's volume.

USAGE:
   recorder ebs [command options] [arguments...]

OPTIONS:
   --self, -s                     Snapshotting for own volume.
   --instance-id value, -i value  Set InstanceId. If '--self' option is set, this is ignored.
   --region value, -r value       Set Region. If '--self' option is set, this is ignored. (default: "ap-northeast-1")
   --lifecycle value, -l value    Set the number of life cycle for snapshot. (default: 5)
```

```
$ recorder ami -h
NAME:
   recorder ami - Creating Image for specific instance

USAGE:
   recorder ami [command options] [arguments...]

OPTIONS:
   --self, -s                     Creating Image for own.
   --instance-id value, -i value  Set InstanceId. If '--self' option is set, this is ignored. (default: "i-xxxxxxx")
   --region value, -r value       Set Region. If '--self' option is set, this is ignored. (default: "ap-northeast-1")
   --reboot                       Reboot instance when create image. (Default-value: false, NOT-Reboot.)
```

If you wish execute another profile, run `$ AWS_DEFAULT_PROFILE='another_profile_name' recorder...` .

## Install

- Download binary-file.

```
$ wget https://github.com/heptagon-inc/recorder/releases/download/v0.4.1/recorder_linux_amd64.zip
$ unzip recorder_linux_amd64.zip -d /usr/local/bin/
```

- Or to install, use `go get`:

```bash
$ go get -d github.com/heptagon-inc/recorder
$ cd $GOPATH/src/github.com/heptagon-inc/recorder
$ make build-local
```

## Configure

Set of credentials, or instances roll configuration required.

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
