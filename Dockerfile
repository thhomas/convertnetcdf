FROM debian:buster

RUN apt-get update -y && apt-get install -y libgdal-dev golang git

ENV GOPATH=/go/
ENV GDAL_SKIP=netCDF
# prevent gdal to use netCDF driver to open file instead of HDF5

COPY ./convert.go /go/src/convert/convert.go

RUN cd /go/src/convert && go get "github.com/lukeroth/gdal" && go install .

ENTRYPOINT ["/go/bin/convert"]


