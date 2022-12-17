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
    curl

WORKDIR /app

# Setup Go programming language

RUN wget https://go.dev/dl/go1.19.3.linux-amd64.tar.gz
RUN tar -C /usr/local -xzf go1.19.3.linux-amd64.tar.gz
RUN chmod +x /usr/local/go/bin/go
RUN chmod +x /usr/local/go/bin/gofmt
RUN cp /usr/local/go/bin/go /usr/bin/go
RUN cp /usr/local/go/bin/gofmt /usr/bin/gofmt

ENV GOLANG_VERSION 1.19.3

# Install kubectl
ENV KUBECTL_VERSION v1.23.13
RUN curl -LO https://dl.k8s.io/release/${KUBECTL_VERSION}/bin/linux/amd64/kubectl

ENV KUBECTL_CHECKSUM fae6957e6a7047ad49cdd20976cd2ce9188b502c831fbf61f36618ea1188ba38
RUN echo "${KUBECTL_CHECKSUM}  kubectl" | sha256sum --check
RUN chmod +x kubectl
RUN mv kubectl /usr/local/bin/kubectl

# copy over source code and build
COPY . .
RUN go build -o kubebackup .
RUN mv kubebackup /usr/bin/kubebackup

# delete source code
RUN rm -rf /app

ENTRYPOINT [ "/usr/bin/kubebackup" ]