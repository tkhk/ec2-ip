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
var is_public, is_all, simple bool
var region string

func init() {
	f := flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	f.BoolVar(&is_public, "public", false, "public ip")
	f.BoolVar(&simple, "simple", false, "only ip")
	f.BoolVar(&is_all, "a", false, "all state ec2 instances(default: ouly running)")
	f.StringVar(&region, "r", "ap-northeast-1", "specify region(default: ap-northeast-1)")
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
			Region: aws.String(region),
			Credentials: credentials.NewSharedCredentials(
				filepath.Join(usr.HomeDir, ".aws", "credentials"), profile),
		},
	)
	if err != nil {
		log.Fatal("Could not get session: %v", err)
	}

	svc := ec2.New(s)

	filters := make([]*ec2.Filter, 1)
	if is_all {
		filters = nil
	} else {
		filters[0] = &ec2.Filter{
			Name:   aws.String("instance-state-name"),
			Values: []*string{aws.String("running")},
		}
	}

	input := ec2.DescribeInstancesInput{
		Filters: filters,
	}

	out, err := svc.DescribeInstances(&input)
	if err != nil {
		log.Fatal(err)
	}

	var nameTag string
	var ipAddress string
	var output string
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

			output = fmt.Sprintf("%v, %v\n", nameTag, ipAddress)
			if simple {
				output = fmt.Sprintf("%v\n", ipAddress)
			} else {
				output = fmt.Sprintf("%v %v\n", nameTag, ipAddress)
			}
			fmt.Fprintf(buf, output)
		}
	}
	buf.Flush()
}
