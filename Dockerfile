## docker image prune
## docker rmi $(docker images -a -q)

# FROM alpine
# RUN mkdir /sif-json2xml
# COPY . / /sif-json2xml/
# WORKDIR /sif-json2xml/
# CMD ["./server"]

### ! run this Dockerfile 
### docker build --tag=sif-json2xml . 

### ! run this docker image
### docker run --name sif-json2xml --net host sif-json2xml:latest

### ! push image to docker hub
### docker tag IMAGE_ID dockerhub-user/sif-json2xml:latest
### docker login
### docker push dockerhub-user/sif-json2xml


###########################
# INSTRUCTIONS
############################
# BUILD
#	docker build --rm -t nsip/sif-json2xml:latest -t nsip/sif-json2xml:v0.1.0 .
# TEST: docker run -it -v $PWD/test/data:/data -v $PWD/test/config.json:/config.json nsip/sif-json2xml:develop .
# RUN: docker run -d nsip/sif-json2xml:develop
#
# PUSH
#	Public:
#		docker push nsip/sif-json2xml:v0.1.0
#		docker push nsip/sif-json2xml:latest
#
#	Private:
#		docker tag nsip/sif-json2xml:v0.1.0 the.hub.nsip.edu.au:3500/nsip/sif-json2xml:v0.1.0
#		docker tag nsip/sif-json2xml:latest the.hub.nsip.edu.au:3500/nsip/sif-json2xml:latest
#		docker push the.hub.nsip.edu.au:3500/nsip/sif-json2xml:v0.1.0
#		docker push the.hub.nsip.edu.au:3500/nsip/sif-json2xml:latest
#
###########################
# DOCUMENTATION
############################



# docker build --rm -t nsip/sif-json2xml:latest -t nsip/sif-json2xml:v0.1.0 .

###########################
# STEP 0 Get them certificates
############################
# (note, step 2 is using alpine now) 
# FROM alpine:latest as certs

############################
# STEP 1 build executable binary (go.mod version)
############################
FROM golang:1.15.3-alpine3.12 as builder
RUN apk add --no-cache ca-certificates
RUN apk update && apk add --no-cache git bash
RUN mkdir -p /sif-json2xml
COPY . / /sif-json2xml/
WORKDIR /sif-json2xml/
RUN ["/bin/bash", "-c", "./build_d.sh"]
RUN ["/bin/bash", "-c", "./release_d.sh"]

############################
# STEP 2 build a small image
############################
FROM alpine
COPY --from=builder /sif-json2xml/app/ /
# NOTE - make sure it is the last build that still copies the files
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
WORKDIR /
CMD ["./server"]

# docker run --rm --mount type=bind,source=$(pwd)/config.toml,target=/config.toml -p 0.0.0.0:1325:1325 nsip/sif-json2xml