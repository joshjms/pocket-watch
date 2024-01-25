# Stage 1: Build the application
FROM ubuntu:22.04 AS build
RUN apt-get update && apt-get install -y ca-certificates gcc g++ golang-go git
RUN update-ca-certificates
ENV PATH=$PATH:/usr/local/go/bin
WORKDIR /app
COPY . .
RUN CGO_ENABLED=1 GOOS=linux go build -a -installsuffix cgo -o bin/pocket-watch .

# Stage 2: Create a minimal runtime image
FROM ubuntu:22.04
RUN apt-get update && apt-get install -y gcc g++ git make libcap-dev
RUN apt-get --no-install-recommends install asciidoc xmlto -y
RUN git clone https://github.com/ioi/isolate.git /tmp/isolate
RUN cd /tmp/isolate && make && make install
RUN cp /tmp/isolate/isolate /usr/local/bin/isolate
RUN rm -rf /tmp/isolate
RUN apt-get remove --purge -y git make libcap-dev
RUN apt-get autoremove -y
RUN apt-get clean
RUN rm -rf /var/lib/apt/lists/*
RUN mkdir /app
ENV PW_META_FILES_DIR=/app/meta
COPY --from=build /app/bin/pocket-watch /app/pocket-watch
RUN chmod +x /app/pocket-watch
EXPOSE 8080
ENTRYPOINT [ "app/./pocket-watch" ]

