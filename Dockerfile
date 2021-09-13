FROM centos:centos7.9.2009
RUN yum update -y && yum install epel-release zip unzip wget curl -y 
WORKDIR /evanesco-miner
ADD . /evanesco-miner
RUN unzip miner-linux.zip && mv ./miner-linux miner && rm miner-linux.zip && mv ./QmNpJg4jDFE4LMNvZUzysZ2Ghvo4UJFcsjguYcx4dTfwKx ./miner
RUN rm Dockerfile
CMD ["/bin/bash"]
