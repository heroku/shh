FROM gliderlabs/alpine:3.1

RUN apk-install ca-certificates

# assumes gox has already installed the files here
COPY .docker_build/shh_linux_amd64 /bin/shh
COPY .docker_build/shh-value_linux_amd64 /bin/shh-value
ENTRYPOINT ["/bin/shh"]
