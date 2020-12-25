# Pull base image
FROM golang:alpine

# Install git
RUN apk update && apk add --no-cache git

# Where our files will be in the docker container 
WORKDIR /opt/app

# environment variables
ENV MONGO_DATABASE=legislation \
    MONGO_COLLECTION=legis_anp


# Copy the source from the current directory to the working Directory inside the container 
# Source also contains go.mod and go.sum which are dependency files
COPY . .
# Get Dependency
RUN go mod download