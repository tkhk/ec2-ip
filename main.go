package main

import (
	"log"

	"bufio"
	"flag"
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
)

var buf = bufio.NewWriter(os.Stdout)
var isPublic, isAll, simple bool
var region string

func usage() {
	fmt.Fprintf(
		os.Stderr,
		`
Usage of %s:
	%s PROFILE [OPTIONS]
	Options\n`,
		os.Args[0],
		os.Args[0])
	flag.PrintDefaults()
}

func init() {
	flag.Usage = usage
	f := flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	f.BoolVar(&isPublic, "public", false, "public ip")
	f.BoolVar(&simple, "simple", false, "only ip")
	f.BoolVar(&isAll, "a", false, "all state ec2 instances(default: ouly running)")
	f.StringVar(&region, "r", "ap-northeast-1", "specify region(default: ap-northeast-1)")
	f.Parse(os.Args[1:])
	if len(f.Args()) < 1 {
		fmt.Println("Error")
		usage()
		return
	}
	f.Parse(f.Args()[1:])
}

func main() {

	profile := os.Args[1]

	svc := ec2.New(session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
		Profile:           profile,
		Config: aws.Config{
			Region: &region,
		},
	})))

	filters := make([]*ec2.Filter, 1)
	if isAll {
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
			if isPublic {
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
				output = fmt.Sprintf("%v %v %v\n", nameTag, ipAddress, *i.LaunchTime)
			}
			fmt.Fprintf(buf, output)
		}
	}
	buf.Flush()
}
