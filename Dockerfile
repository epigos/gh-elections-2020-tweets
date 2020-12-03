##################################
# STEP 1 build executable binary #
##################################
FROM golang:1.14-alpine AS build

# All these steps will be cached
RUN mkdir /app
WORKDIR /app

# Copying the .mod and .sum files before the rest of the code
# supposedly improves the caching behavior of Docker
# See: https://medium.com/@petomalina/using-go-mod-download-to-speed-up-golang-docker-builds-707591336888
# And: https://medium.com/@pierreprinetti/the-go-1-11-dockerfile-a3218319d191
COPY go.mod .
COPY go.sum .

# Get dependancies - will also be cached if we won't change mod/sum
RUN go mod download
# COPY the source code as the last step
COPY . .

# Build the Go app
RUN go build -o ./build/tweets cmd/tweets/main.go

##############################
# STEP 2 build a small image #
##############################
FROM alpine:3.9
WORKDIR /app

COPY --from=build /app/build/tweets /bin/tweets

ENTRYPOINT ["/bin/tweets"]