FROM gliderlabs/alpine:latest

RUN apk-install ca-certificates

# assumes gox has already installed the files here
COPY .docker_build/shh /bin/shh
COPY .docker_build/shh-value /bin/shh-value
ENTRYPOINT ["/bin/shh"]
