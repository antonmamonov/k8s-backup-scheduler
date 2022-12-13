FROM ubuntu:20.04@sha256:450e066588f42ebe1551f3b1a535034b6aa46cd936fe7f2c6b0d72997ec61dbd

# fix timezone stalling during build
ENV TZ=America/Toronto
RUN ln -snf /usr/share/zoneinfo/$TZ /etc/localtime && echo $TZ > /etc/timezone

RUN apt-get update
RUN apt-get install -y \
    build-essential \
    cmake \
    git \
    make \
    wget \
    tmux

WORKDIR /app

# Setup Go programming language

RUN wget https://go.dev/dl/go1.19.3.linux-amd64.tar.gz
RUN tar -C /usr/local -xzf go1.19.3.linux-amd64.tar.gz
RUN chmod +x /usr/local/go/bin/go
RUN chmod +x /usr/local/go/bin/gofmt
RUN cp /usr/local/go/bin/go /usr/bin/go
RUN cp /usr/local/go/bin/gofmt /usr/bin/gofmt

ENV GOLANG_VERSION 1.19.3

# copy over source code and build
COPY . .
RUN go build -o main .
RUN mv main /usr/bin/main

# delete source code
RUN rm -rf /app

ENTRYPOINT [ "/usr/bin/main" ]