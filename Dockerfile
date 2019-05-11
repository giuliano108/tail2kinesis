FROM ubuntu:xenial
RUN apt-get update && apt-get install -y curl software-properties-common build-essential git-core entr

RUN curl -sL add https://deb.nodesource.com/gpgkey/nodesource.gpg.key | apt-key add -
RUN echo "deb http://deb.nodesource.com/node_11.x xenial main" > /etc/apt/sources.list.d/nodesource.list
RUN echo "deb-src http://deb.nodesource.com/node_11.x xenial main" >> /etc/apt/sources.list.d/nodesource.list
RUN apt-get update
RUN apt-get install -y nodejs
RUN npm install -g kinesalite --unsafe

RUN apt-get install -y awscli
RUN mkdir /root/.aws
RUN echo '[default]\n\
region = x' > /root/.aws/config
RUN echo '[default]\n\
aws_secret_access_key = x\n\
aws_access_key_id = x' > /root/.aws/credentials

RUN add-apt-repository ppa:gophers/archive
RUN apt-get update
RUN apt-get install -y golang-1.11-go

RUN apt-get install -y jq

ENV GOPATH="/gopath"
RUN mkdir /gopath
ENV PATH="/usr/lib/go-1.11/bin:${PATH}"
