package controller

import (
	"encoding/base64"
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
)

const userdata = `#!/bin/sh
curl -LO "https://dl.k8s.io/release/$(curl -L -s https://dl.k8s.io/release/stable.txt)/bin/linux/amd64/kubectl"
chmod +x kubectl
sudo apt update
sudo apt install -y awscli
export AWS_ACCESS_KEY_ID=%s
export AWS_SECRET_ACCESS_KEY=%s
aws eks --region %s update-kubeconfig --name %s --kubeconfig config.yaml
./kubectl --kubeconfig config.yaml apply -f %s
`

// Ubuntu 20.04 LTS hvm:ebs-ssd (amd64)
// See https://cloud-images.ubuntu.com/locator/ec2/
var regionDefaultImages map[string]string = map[string]string{
	"af-south-1":      "ami-0e5e0920562e66b27",
	"ap-northeast-1":  "ami-0f08dd35fe25614c9",
	"ap-northeast-2":  "ami-0688e5f8e4d3823dd",
	"ap-northeast-3":  "ami-00fbd944ec4f0a3a5",
	"ap-southeast-1":  "ami-0b17218a3cbd5ec4a",
	"ap-southeast-2":  "ami-0eb2b6effadbcc070",
	"ap-south-1":      "ami-03caa8eb4ce0612a2",
	"ap-east-1":       "ami-0db28d8badd461a9b",
	"ca-central-1":    "ami-0f71e3f96dbd2a805",
	"cn-north-1":      "ami-0592ccadb56e65f8d", // Note: this is 20180126
	"cn-northwest-1":  "ami-007d0f254ea0f8588", // Note: this is 20180126
	"eu-central-1":    "ami-0dca0d6d4f591c2f4",
	"eu-north-1":      "ami-09f4f0c83b3a83bfb",
	"eu-south-1":      "ami-0d7681eb054d3e470",
	"eu-west-1":       "ami-0dd0f5f97a21a8fe9",
	"eu-west-2":       "ami-0925a09a52d18d09a",
	"eu-west-3":       "ami-03ba9eb13b5975d61",
	"me-south-1":      "ami-0b1119d3877cc4269",
	"sa-east-1":       "ami-0b0cbbe7f83e5bbfd",
	"us-east-1":       "ami-0fa37863afb290840",
	"us-east-2":       "ami-0b5add21e87587ae1",
	"us-west-1":       "ami-01d3aaa89c3c158e5",
	"us-west-2":       "ami-07e573cdaa16d4e61",
	"us-gov-west-1":   "ami-a7edd7c6",
	"us-gov-east-1":   "ami-c39973b2",
	"custom-endpoint": "",
}

func deployAgentWithBastion(ec2Service *ec2.EC2, securityGroups []string, subnet string, region, url, clusterName, secretID, secretKey string) (*string, error) {
	runInstanceOutput, err := ec2Service.RunInstances(&ec2.RunInstancesInput{
		ImageId:          aws.String(regionDefaultImages[region]),
		InstanceType:     aws.String(ec2.InstanceTypeT2Micro),
		MaxCount:         aws.Int64(1),
		MinCount:         aws.Int64(1),
		SecurityGroupIds: aws.StringSlice(securityGroups),
		SubnetId:         aws.String(subnet),
		UserData:         aws.String(base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf(userdata, secretID, secretKey, region, clusterName, url)))),
	})
	if err != nil {
		return nil, err
	}
	if len(runInstanceOutput.Instances) == 0 {
		return nil, fmt.Errorf("no instances returned in run request")
	}

	return runInstanceOutput.Instances[0].InstanceId, nil
}

func deleteBastion(ec2Service *ec2.EC2, instanceID *string) error {
	if _, err := ec2Service.TerminateInstances(&ec2.TerminateInstancesInput{
		InstanceIds: []*string{instanceID},
	}); err != nil {
		return err
	}
	return nil
}
