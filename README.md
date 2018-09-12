convertnetcdf
=============
Command-line tool to extract raster bands from netCDF/HDF5 file.
Written in go, it uses [gdal binding for go](https://github.com/lukeroth/gdal) and gdal 2.3.1.

```
$ docker run --rm thhomas/convertnetcdf --help
Usage of /go/bin/convert:
  -a	extract all bands
  -i string
    	a netCDF file
  -n	extract natural color image
  -o string
    	output folder
  -s string
    	subdataset names separated with comma (e.g. B4,B8)
```
Example of command line:
```
docker run --rm -v /path/to/local/netCDF/folder/:/data/ -v /path/to/local/output/folder/:/output/ thhomas/convertnetcdf -i /data/netcdfFileName.nc -o /output/ -s B4,B8
```