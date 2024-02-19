FROM ubuntu:jammy

ARG NPM_REGISTRY

RUN apt-get update && DEBIAN_FRONTEND=noninteractive apt-get install -y curl software-properties-common build-essential git-core entr npm

RUN echo registry=${NPM_REGISTRY} > /root/.npmrc
RUN npm install -g kinesalite

RUN DEBIAN_FRONTEND=noninteractive apt-get install -y awscli
RUN mkdir /root/.aws
RUN echo '[default]\n\
region = x' > /root/.aws/config
RUN echo '[default]\n\
aws_secret_access_key = x\n\
aws_access_key_id = x' > /root/.aws/credentials

RUN add-apt-repository ppa:longsleep/golang-backports
RUN apt-get update
RUN DEBIAN_FRONTEND=noninteractive apt-get install -y golang-go
RUN DEBIAN_FRONTEND=noninteractive apt-get install -y jq

ENV GOPATH="/gopath"
RUN mkdir /gopath
