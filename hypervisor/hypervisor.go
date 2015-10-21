/*
**初始化worker，启动worker
 */
package hypervisor

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/golang/glog"
	"github.com/spider-docker/workers"
)

const (
	USER_QUEUE_NAME      = "HypervisorUser"
	CONTAINER_QUEUE_NAME = "HypervisorContainer"
)

var ()

func Main() {
	var (
		SQS                        *sqs.SQS
		getUserQueueUrlOutput      *sqs.GetQueueUrlOutput
		getContainerQueueUrlOutput *sqs.GetQueueUrlOutput
		UserQueueUrl               *string
		ContainerQueueUrl          *string

		Dynamo *dynamodb.DynamoDB

		socialWorker *workers.SocialWorker
	)

	SQS = sqs.New(&aws.Config{Region: aws.String("cn-north-1")})
	getUserQueueUrlOutput, err := SQS.GetQueueUrl(&sqs.GetQueueUrlInput{QueueName: aws.String(USER_QUEUE_NAME)})
	if err != nil {
		glog.Errorln("Error on connect user queue url:", err.Error())
		return
	}
	UserQueueUrl = getUserQueueUrlOutput.QueueUrl
	getContainerQueueUrlOutput, err = SQS.GetQueueUrl(&sqs.GetQueueUrlInput{QueueName: aws.String(CONTAINER_QUEUE_NAME)})
	if err != nil {
		glog.Errorln("Error on connect container queue url:", err.Error())
		return
	}
	ContainerQueueUrl = getContainerQueueUrlOutput.QueueUrl

	Dynamo = dynamodb.New(&aws.Config{Region: aws.String("cn-north-1")})

	socialWorker = workers.NewSocialWorker(SQS, UserQueueUrl, ContainerQueueUrl, Dynamo)
	socialWorker.Start()
}
