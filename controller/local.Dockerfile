# ============= Build Stage ======================
FROM golang:1.17-alpine AS builder

RUN mkdir -p /go/src/github.com/chain4travel

# Copy the code into the container
WORKDIR $GOPATH/src/github.com/chain4travel
COPY caminogo caminogo
COPY camino-testing camino-testing

WORKDIR $GOPATH/src/github.com/chain4travel/camino-testing
RUN go mod edit -replace github.com/chain4travel/caminogo=../caminogo
RUN go mod download

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o /build/test-controller controller/main.go

# ============= Execution Stage ================
FROM docker:stable AS execution
WORKDIR /run

# Copy the binary into the execution container
COPY --from=builder /build/test-controller .

# Note that this CANNOT be an execution list else the variables won't be expanded
# See: https://stackoverflow.com/questions/40454470/how-can-i-use-a-variable-inside-a-dockerfile-cmd
CMD set -euo pipefail && ./test-controller \
    --test-volume=${TEST_VOLUME} \
    --test-volume-mountpoint=${TEST_VOLUME_MOUNTPOINT} \
    --test=${TEST_NAME} \
    --camino-image-name=${CAMINO_IMAGE_NAME} \
    --byzantine-image-name=${BYZANTINE_IMAGE_NAME} \
    --docker-network=${NETWORK_ID} \
    --subnet-mask=${SUBNET_MASK} \
    --test-controller-ip=${TEST_CONTROLLER_IP} \
    --gateway-ip=${GATEWAY_IP} \
    --log-level=${LOG_LEVEL} 2>&1 | tee ${LOG_FILEPATH}