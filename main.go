package main

import (
	"log"

	"os/user"
	"path/filepath"

	"bufio"
	"flag"
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
)

var buf = bufio.NewWriter(os.Stdout)
var is_public, is_all bool

func init() {
	f := flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	f.BoolVar(&is_public, "public", false, "output public ip")
	//f.BoolVar(&is_all, "all", false, "output all state ec2 instances(default: ouly running)")
	f.Parse(os.Args[1:])
	f.Parse(f.Args()[1:])
}

func main() {
	profile := os.Args[1]

	usr, err := user.Current()
	if err != nil {
		log.Fatal(err)
	}

	s, err := session.NewSession(
		&aws.Config{
			Region: aws.String("ap-northeast-1"),
			Credentials: credentials.NewSharedCredentials(
				filepath.Join(usr.HomeDir, ".aws", "credentials"), profile),
		},
	)
	if err != nil {
		log.Fatal("Could not get session: %v", err)
	}

	svc := ec2.New(s)

	input := ec2.DescribeInstancesInput{
		Filters: []*ec2.Filter{
			&ec2.Filter{
				Name:   aws.String("instance-state-name"),
				Values: []*string{aws.String("running")},
			},
		},
	}
	out, err := svc.DescribeInstances(&input)
	if err != nil {
		log.Fatal(err)
	}

	var nameTag string
	var ipAddress string
	for _, r := range out.Reservations {
		for _, i := range r.Instances {
			for _, t := range i.Tags {
				nameTag = ""
				if *t.Key == "Name" {
					nameTag = *t.Value
					break
				}
			}

			ipAddress = ""
			if is_public {
				if i.PublicIpAddress != nil {
					ipAddress = *i.PublicIpAddress
				}
			} else {
				if i.PrivateIpAddress != nil {
					ipAddress = *i.PrivateIpAddress
				}
			}
			fmt.Fprintf(buf, fmt.Sprintf(
				"%v, %v\n",
				nameTag,
				ipAddress,
			))
		}
	}
	buf.Flush()
}
