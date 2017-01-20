FROM alpine:latest

RUN apk add --no-cache rsync

COPY ./webroot/ /mnt

# Command to run for init container:
CMD ["rsync", "-a", "/mnt/", "/srv/"]