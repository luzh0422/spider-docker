##################################################
#Dockerfile to build golang container images
#Based on centos
##################################################

# Using a compact OS
FROM centos:golang

MAINTAINER Luzh 352233283@qq.com

USER root

RUN yum update -y
# Install git
#RUN yum install -y git
# Install go
#RUN yum install -y go 
# Install vim
RUN yum install -y vim
	
# Set gopath
RUN echo "export GOPATH=/usr/local/go/" >> /etc/profile.d/go.sh
RUN source /etc/profile.d/go.sh
RUN export GOPATH=/usr/local/go/


# Set aws_key
RUN mkdir ~/.aws
RUN echo "[default]" >> ~/.aws/credentials
RUN echo "aws_access_key_id = AKIAPNYXU5GYGP2JHCPQ" >> ~/.aws/credentials
RUN echo "aws_secret_access_key = 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" >> ~/.aws/credentials

# Install golang package
#RUN go get github.com/golang/glog
#RUN go get github.com/golang/text
#RUN go get github.com/golang/net
#RUN go get github.com/PuerkitoBio/goquery
#RUN go get github.com/vaughan0/go-ini
#RUN go get github.com/hu17889/go_spider
#RUN go get github.com/aws/aws-sdk-go
#RUN go get github.com/bitly/go-simplejson

# Install golang.org package
RUN mkdir -p /usr/local/go/src/golang.org/x
RUN cp -rf /usr/local/go/src/github.com/golang/net /usr/local/go/src/golang.org/x/
RUN cp -rf /usr/local/go/src/github.com/golang/text /usr/local/go/src/golang.org/x/ 

RUN mkdir /usr/local/go/src/github.com/spider-docker
RUN mkdir /usr/local/go/src/github.com/spider-docker/models
RUN mkdir /usr/local/go/src/github.com/spider-docker/workers
RUN mkdir /usr/local/go/src/github.com/spider-docker/hypervisor
RUN mkdir /usr/local/go/src/github.com/spider-docker/events
# ADD file
ADD main.go /usr/local/go/src/github.com/spider-docker/
ADD	header.json	/usr/local/go/src/github.com/spider-docker/
ADD ./models /usr/local/go/src/github.com/spider-docker/models
ADD ./workers /usr/local/go/src/github.com/spider-docker/workers
ADD ./hypervisor /usr/local/go/src/github.com/spider-docker/hypervisor
ADD ./events /usr/local/go/src/github.com/spider-docker/events


# Set WORKDIR
# WORKDIR /usr/local/go/src/github.com/

EXPOSE 22

# run getFriends
CMD source /etc/profile.d/go.sh && go run usr/local/go/src/github.com/spider-docker/main.go 