FROM golang:1.23.4-bookworm
USER root


################
## Install Go ##
################
# RUN cd /usr/local
# RUN curl https://go.dev/dl/go1.23.4.linux-amd64.tar.gz
# RUN tar -C /usr/local -xzf go1.23.4.linux-amd64.tar.gz
# RUN export PATH=$PATH:/usr/local/go/bin
# RUN cd /

##################
## Install GoSu ##
##################
WORKDIR /app/src
RUN set -eux; \
	apt-get update; \
	apt-get install -y gosu; \
	rm -rf /var/lib/apt/lists/*; \
# verify that the binary works
	gosu nobody true


###########################################
## Add Src Code and Install Dependencies ##
###########################################
RUN go install github.com/air-verse/air@latest
COPY src ./
RUN go mod download



#########################################
## Making Users and Setting Permisions ##
#########################################
RUN adduser threemix
RUN chown threemix:threemix -R ../
RUN find ../ -type f -exec chmod u=r,g=r,o= {} +
RUN find ../ -type d -exec chmod u=rx,g=rx,o=rx {} +
RUN chmod u=rx,g=rx,o=rx ../
RUN mkdir ./tmp


##################
## Build Go App ##
##################
# RUN go build -o ./bin/threemix-executable ./cmd/main.go



#####################
## Copy Entrypoint ##
#####################
EXPOSE 8080
ADD ./docker/entrypoint-dev.sh /app/entrypoint.sh
RUN chmod u=rx,g=r,o=r /app/entrypoint.sh

ENTRYPOINT ["/app/entrypoint.sh"]

