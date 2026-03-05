# syntax=docker/dockerfile:1

ARG BASE_IMG=golang:1.24
ARG RUN_IMG=alpine:3.14

#############
# Base stage
#############
FROM ${BASE_IMG} AS base

RUN DEBIAN_FRONTEND=noninteractive \
  apt-get update && apt-get install -y tzdata && \
  update-ca-certificates

# Add an unprivileged user
ENV USER=appuser
ENV UID=10001
RUN adduser \    
    --disabled-password \    
    --gecos "" \    
    --home "/nonexistent" \    
    --no-create-home \ 
    --shell "/sbin/nologin" \        
    --uid "${UID}" \    
    "${USER}"


#############
# Compile stage
#############
FROM ${BASE_IMG} AS compile

WORKDIR /go/src/github.com/massdriver-cloud/mcp-server
COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o /usr/bin/mcp-server .

ENTRYPOINT ["/usr/bin/mcp-server"]


#############
# Run stage
#############
FROM ${RUN_IMG}

# Get tzdata
COPY --from=base /usr/share/zoneinfo /usr/share/zoneinfo

# Get updated certs
COPY --from=base /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# Use unprivileged user
COPY --from=base /etc/passwd /etc/passwd
COPY --from=base /etc/group /etc/group
USER appuser:appuser

COPY --from=compile /usr/bin/mcp-server /usr/bin/mcp-server

ENTRYPOINT ["/usr/bin/mcp-server"]
