FROM golang:1.14-alpine

# The latest alpine images don't have some tools like (`git` and `bash`).
# Adding git, bash and openssh to the image
RUN apk update && apk upgrade && \
    apk add --no-cache bash git openssh

RUN apk add --no-cache tzdata
ENV TZ Asia/Bangkok

WORKDIR /app

RUN go get -u github.com/pilu/fresh

CMD go get .; \
    fresh;