package main

import (
	"flag"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/lukeroth/gdal"
)

func ExtractTifFromNetCDF(imagePath string, subdatasetNames []string, outputFolder string) []string {
	netCDFDs, err := gdal.Open(imagePath, gdal.ReadOnly)
	if err != nil {
		fmt.Printf("Unable to open NetCDF file: %s", imagePath)
		fmt.Println(err)
	}
	defer netCDFDs.Close()

	outSrs := gdal.CreateSpatialReference("")
	outSrs.FromWKT(netCDFDs.MetadataItem("Grid_Projection", ""))
	inSrs := gdal.CreateSpatialReference("")
	inSrs.FromEPSG(4326)

	imageRes, _ := strconv.ParseFloat(strings.Split(filepath.Base(imagePath), "_")[1][1:], 64)

	transform := gdal.CreateCoordinateTransform(inSrs, outSrs)

	latSubdataset := fmt.Sprintf("HDF5:\"%s\"://lat", imagePath)
	lonSubdataset := fmt.Sprintf("HDF5:\"%s\"://lon", imagePath)
	latDs, _ := gdal.Open(latSubdataset, gdal.ReadOnly)
	defer latDs.Close()
	lonDs, _ := gdal.Open(lonSubdataset, gdal.ReadOnly)
	defer lonDs.Close()

	uLLat := make([]float64, 1)
	uLLon := make([]float64, 1)
	bandMap := make([]int, 1)
	bandMap[0] = 1
	err = latDs.IO(gdal.Read, 0, 0, 1, 1, uLLat, 1, 1, 1, []int{1}, 0, 0, 0)
	err = lonDs.IO(gdal.Read, 0, 0, 1, 1, uLLon, 1, 1, 1, []int{1}, 0, 0, 0)

	pointGeom := gdal.CreateFromJson(fmt.Sprintf("{\"type\": \"Point\", \"coordinates\": [%f, %f] }", uLLon[0], uLLat[0]))
	pointGeom.Transform(transform)

	imageName := strings.Split(path.Base(imagePath), ".")[0]
	var outputFiles []string

	for _, item := range subdatasetNames {
		subdatasetName := fmt.Sprintf("HDF5:\"%s\"://%s", imagePath, item)
		subDs, err := gdal.Open(subdatasetName, gdal.ReadOnly)
		if err != nil {
			fmt.Printf("Unable to open subdataset %s\n", subdatasetName)
			fmt.Println(err)
		}

		outputFile := path.Join(outputFolder, imageName+"_"+item+".tif")
		outputFiles = append(outputFiles, outputFile)

		outDs := gdal.GDALTranslate(outputFile, subDs, []string{})
		// outDs, err := gdal.Open(outputFile, gdal.Update)
		outSrsWKT, _ := outSrs.ToWKT()
		outDs.SetProjection(outSrsWKT)

		// gt := [6]float64{envelope.MinX(), imageRes, 0, envelope.MinY(), 0, imageRes}
		gt := [6]float64{pointGeom.X(0), imageRes, 0, pointGeom.Y(0), 0, imageRes}
		outDs.SetGeoTransform(gt)
		outDs.Close()
		subDs.Close()

	}

	return outputFiles

}

func main() {

	netCDFPathPtr := flag.String("i", "", "a netCDF file")
	outputFolderPtr := flag.String("o", "", "output folder")
	subdatasetNamePtr := flag.String("s", "", "subdataset names separated with comma (e.g. B4,B8)")
	allBandsPtr := flag.Bool("a", false, "extract all bands")
	naturalPtr := flag.Bool("n", false, "extract natural color image")

	flag.Parse()

	os.Setenv("GDAL_SKIP", "netCDF")

	if *netCDFPathPtr == "" {
		flag.Usage()
		os.Exit(1)
	}

	if *outputFolderPtr == "" {
		*outputFolderPtr = os.TempDir()
	}
	var subdatasetNames []string
	if *subdatasetNamePtr == "" {
		subdatasetNames = []string{"B4", "B3", "B2"}
	} else {
		subdatasetNames = strings.Split(*subdatasetNamePtr, ",")
	}
	if *allBandsPtr != false {
		subdatasetNames = []string{}
		netCDFDs, _ := gdal.Open(*netCDFPathPtr, gdal.ReadOnly)
		gdalSubdatasets := netCDFDs.Metadata("SUBDATASETS")
		for _, gdalSubdataset := range gdalSubdatasets {
			var nameExp = regexp.MustCompile("SUBDATASET_[0-9]+_NAME.*")
			if nameExp.MatchString(gdalSubdataset) {
				subdatasetNames = append(subdatasetNames, strings.Split(gdalSubdataset, "://")[1])
			}
		}
		netCDFDs.Close()
	}
	if *naturalPtr != false {
		// TODO
	}

	ExtractTifFromNetCDF(*netCDFPathPtr, subdatasetNames, *outputFolderPtr)
}
