# here we have service build with go need to dockarize it 
# lets see how we can dockarize multi-stage a go application
FROM golang:alpine AS builder

# here have 2 folder authdb is like controller and main is import this controller function and use them 
WORKDIR /go/src/app
#here copy all of code it should copy mod,sum first and install dependencies then copy the code and build the binary ,but this old way to build docker go file
COPY . . 
# after copy all the code got to main directory
WORKDIR /go/src/app/main 
#go get all the dependency -v verbose for ./... this dir ./or subdirs ...
RUN go get -v ./...
# build the binary  and put the output -o to /go/bin/app  also -v verbose for ./... this dir ./or subdirs ...
# can be just  go build -o /go/bin/app .
RUN go build -o /go/bin/app -v ./...


# this not optimal way this old way to build docker go file 
# this is better way 
# COPY go.mod go.sum ./ # first copy mod and sum
# RUN go mod download # install dependencies

# COPY . . # copy all the code
# RUN go build -o /go/bin/app ./main # build the binary
# this better for layer cacheing cause if you have change in code it will not consider there change in dependencies and install it again 

# Final stage
# use any iamge you will run machine code in it , no need image have go compiler
FROM alpine:latest  
# make sure to add ca-certificates on the continer cause we use https , you in mid project make ca with ansible on ec2 machine 
# and mount them with docker compsose inside container nginx , you can do make it like this , install inside the container of nginx 
RUN apk --no-cache add ca-certificates 
# make app directory
WORKDIR /app
# copy the binaries only => the go generate binaries only so just make it executable and run it no need for compileres or specific image look its normal image alpine not something have go image to have compiler
COPY --from=builder /go/bin/app .
# make binary executable
RUN chmod +x ./app
ENTRYPOINT ["./app"] 
#just describe the image , metdata 
LABEL Name=auth Version=2.0.0
# this also describtion but the real port inside code will be on 8080 also , this not open any thing it just describtion
EXPOSE 8080
