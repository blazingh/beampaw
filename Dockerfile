FROM golang:1.21

# install needed tools
RUN apt-get update && apt-get install xz-utils 

# install nodejs
RUN wget https://nodejs.org/dist/v20.9.0/node-v20.9.0-linux-x64.tar.xz && tar -xf node-v20.9.0-linux-x64.tar.xz -C /lib/

# add the nodejs path to env
ENV PATH=/lib/node-v20.9.0-linux-x64/bin:$PATH

# set a working directory
WORKDIR /app

# copy modules
COPY package.json ./
COPY go.mod ./
COPY go.sum ./
COPY Makefile ./

# bootstrap the project
RUN make init/modules

# copy rest of the files
COPY . ./

# start the project
CMD make run
