package main

import (
	"bufio"
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"math"
	"os"
	"strconv"

	"github.com/gonum/plot"
	"github.com/gonum/plot/plotter"
	"github.com/gonum/plot/plotutil"
	"github.com/gonum/plot/vg"
)

const (
	earthRadius      = 6378.137
	equatorialRadius = 6378.1370
	polarRadius      = 6356.752314
)

type geoPoint struct {
	Latitude  float64
	Longitude float64
}

func main() {
	e := openErrorLog()
	r := openResult()
	defer e.Close()
	defer r.Close()
	writer := bufio.NewWriter(r)

	lines := openLocationFile("./location.csv")
	var stdist, staz, hfdist, hfaz []float64

	for i := 0; i < len(lines)-2; i += 2 {
		p1 := geoPoint{lines[i], lines[i+1]}
		p2 := geoPoint{lines[i+2], lines[i+3]}

		st, az1 := sphericalTrigonometry(p1, p2)
		stdist = append(stdist, st)
		staz = append(staz, az1)

		hf, az2 := hubenyFormula(p1, p2)
		hfdist = append(hfdist, hf)
		hfaz = append(hfaz, az2)

		fmt.Fprintln(writer, "\n", "始点：", p1, "終点：", p2)
		fmt.Fprintln(writer, "球面三角法 距離：", st, "km 方位角：", az1)
		fmt.Fprintln(writer, "Hubenyの公式 距離：", hf, "km 方位角：", az2)
		fmt.Fprintln(writer, "距離差異：", math.Abs(st-hf), "km 方位角差異：", math.Abs(az1-az2))
	}
	writer.Flush()
	plotDist(stdist, hfdist)
	plotAz(staz, hfaz)
}

func sphericalTrigonometry(p1 geoPoint, p2 geoPoint) (float64, float64) {
	var dlong, distance, azimuth float64
	dlong = p2.Longitude - p1.Longitude
	distance = earthRadius * math.Acos((math.Sin(toRad(p1.Latitude))*math.Sin(toRad(p2.Latitude)))+
		(math.Cos(toRad(p1.Latitude))*math.Cos(toRad(p2.Latitude))*math.Cos(toRad(dlong))))
	azimuth = toDeg(math.Atan2(math.Sin(toRad(dlong)), (math.Cos(toRad(p1.Latitude))*math.Tan(toRad(p2.Latitude)))-
		(math.Sin(toRad(p1.Latitude))*math.Cos(toRad(dlong)))))
	if azimuth < 0 {
		azimuth = 360 + azimuth
	}

	return distance, azimuth
}

func hubenyFormula(p1 geoPoint, p2 geoPoint) (float64, float64) {
	dx := toRad(p2.Longitude - p1.Longitude)
	dy := toRad(p2.Latitude - p1.Latitude)
	myu := toRad((p1.Latitude + p2.Latitude) / 2)

	e := math.Sqrt((math.Pow(equatorialRadius, 2) - math.Pow(polarRadius, 2)) / math.Pow(equatorialRadius, 2))
	W := math.Sqrt(1 - (math.Pow(e, 2) * math.Pow(math.Sin(myu), 2)))
	M := (equatorialRadius * (1 - math.Pow(e, 2))) / math.Pow(W, 3)
	N := equatorialRadius / W

	return math.Hypot(dy*M, dx*N*math.Cos(myu)), 90 - toDeg(math.Atan2(dy*M, dx*N*math.Cos(myu)))
}

func toRad(deg float64) float64 {
	return deg * math.Pi / 180
}

func toDeg(rad float64) float64 {
	return rad * 180 / math.Pi
}

func (p *geoPoint) toRad() {
	toRad(p.Latitude)
	toRad(p.Longitude)
}

func openErrorLog() *os.File {
	logfile, err := os.OpenFile("./error.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		panic("cannot open error.log:" + err.Error())
	}

	log.SetOutput(io.MultiWriter(logfile, os.Stdout))
	log.SetFlags(log.Ldate | log.Ltime)

	return logfile
}

func openResult() *os.File {
	logfile, err := os.OpenFile("./result.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		log.Panic("cannot open result.txt:" + err.Error())
	}

	return logfile
}

func openLocationFile(filePath string) []float64 {
	file, err := os.Open(filePath)
	if err != nil {
		log.Fatal("error opning locationfile")
	}
	defer file.Close()

	lines := make([]float64, 0, 100)
	reader := csv.NewReader(file)
	reader.Comma = ','

	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		} else if err != nil {
			log.Fatal(err)
		}
		for _, v := range record {
			var f, _ = strconv.ParseFloat(v, 64)
			lines = append(lines, f)
		}
	}
	return lines
}

func plotPoints(list []float64) plotter.XYs {
	pts := make(plotter.XYs, len(list))
	for i := range pts {
		pts[i].X = float64(i)
		pts[i].Y = list[i]
	}
	return pts
}

func plotDist(stdist, hfdist []float64) {
	p, err := plot.New()
	if err != nil {
		log.Panic(err)
	}

	p.Title.Text = "Distance between 2 points"
	p.X.Label.Text = "X"
	p.Y.Label.Text = "Distance [km]"

	err = plotutil.AddLinePoints(p,
		"SphericalTrigonometry", plotPoints(stdist),
		"hubenyFormula", plotPoints(hfdist))
	if err != nil {
		log.Panic(err)
	}

	if err := p.Save(8*vg.Inch, 8*vg.Inch, "distance.png"); err != nil {
		log.Panic(err)
	}
}

func plotAz(staz, hfaz []float64) {
	p, err := plot.New()
	if err != nil {
		log.Panic(err)
	}

	p.Title.Text = "Azimuth between 2 points"
	p.X.Label.Text = "X"
	p.Y.Label.Text = "Azimuth [°]"

	err = plotutil.AddLinePoints(p,
		"SphericalTrigonometry", plotPoints(staz),
		"hubenyFormula", plotPoints(hfaz))
	if err != nil {
		log.Panic(err)
	}

	if err := p.Save(8*vg.Inch, 8*vg.Inch, "azimuth.png"); err != nil {
		log.Panic(err)
	}
}
