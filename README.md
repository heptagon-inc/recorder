# recorder

## Description

Create a snapshot of the Amazon EC2.

## Usage

Set AWS configuration file.

```
cat ~/.aws/credentials
[default]
aws_access_key_id = XXXXXXXXXX
aws_secret_access_key = XXXXXXXXXXXXXXX
```

Run the `recorder self`

```
$ recorder self -h
NAME:
   self - Snapshotting for own volume.

USAGE:
   command self [command options] [arguments...]

OPTIONS:
   --lifecycle, -l "5"	Set the number of life cycle for snapshot.
```

```
$ recorder -h
NAME:
   recorder - Create a snapshot of the Amazon EC2.

USAGE:
   recorder [global options] command [command options] [arguments...]

VERSION:
   0.1.0

AUTHOR(S):
   youyo

COMMANDS:
   self		Snapshotting for own volume.
   help, h	Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --help, -h		show help
   --version, -v	print the version
```


## Install

To install, use `go get`:

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
