FROM golang:latest

COPY ./dummymetron /dummymetron

CMD /dummymetron -key /loggregator-certs/metron.key -cert /loggregator-certs/metron.crt -ca /loggregator-certs/loggregator-ca.crt
