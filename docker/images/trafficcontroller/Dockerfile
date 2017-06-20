FROM golang:latest

RUN apt-get update --yes && \
    apt-get install --yes dnsutils jq

COPY ./trafficcontroller /trafficcontroller
COPY ./config.json /config.json
COPY ./start_tc.sh /start_tc.sh

ENV DEPLOYMENT=foo

CMD /start_tc.sh
#CMD /traffficcontroller -disableAccessControll -config /config.json
