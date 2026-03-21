
FROM golang:alpine AS builder

WORKDIR /app
# copy dependencies files 
COPY go.mod go.sum ./
# install dependencies
RUN go mod download
# copy all the code
COPY . .
# build the binary # I WILL BUILD ./MAIN AND OUTPUT WILL BE IN /go/bin
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /go/bin/app ./main
# CGO_ENABLED=0 to disable CGO => we don't need to have go compiler to run the binary , so make binary static run everywhere
# GOOS=linux GOARCH=amd64 to build the binary for linux amd64
# can run with out those command => go build -o /go/bin/app ./main
# RUN apk --no-cache add ca-certificates  # doesn't need it right now 


FROM scratch  
# here for ssl certificates to when this service sends request to other service will be use https , this not realted to ingess tls certificates
# make app directory
WORKDIR /app
#copy cert i don't know if it will use it 
# COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# copy the binaries only => the go generate binaries only so just make it executable and run it no need for compileres or specific image look its normal image alpine not something have go image to have compiler
COPY --from=builder /go/bin/app .
# make binary executable
# RUN chmod +x ./app # no need to make it executable cause its already executable from builder 

# this also describtion but the real port inside code will be on 8080 also , this not open any thing it just describtion
EXPOSE 8080

ENTRYPOINT ["./app"] 