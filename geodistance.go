package main

import (
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"math"
	"os"
	"strconv"
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

func sphericalTrigonometry(p1 geoPoint, p2 geoPoint) (float64, float64) {
	return earthRadius * math.Acos((math.Sin(toRad(p1.Latitude))*math.Sin(toRad(p2.Latitude)))+
			(math.Cos(toRad(p1.Latitude))*math.Cos(toRad(p2.Latitude))*math.Cos(toRad(p2.Longitude-p1.Longitude)))),
		toDeg(math.Atan2(math.Sin(toRad(p2.Longitude-p1.Longitude)), (math.Cos(toRad(p1.Latitude))*math.Tan(toRad(p2.Latitude)))-
			(math.Sin(toRad(p1.Latitude))*math.Cos(toRad(p2.Longitude-p1.Longitude)))))
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

func openLog() *os.File {
	logfile, err := os.OpenFile("./geodistance.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		panic("cannot open test.log:" + err.Error())
	}

	log.SetOutput(io.MultiWriter(logfile, os.Stdout))
	log.SetFlags(log.Ldate | log.Ltime)

	return logfile
}

func openFile(filePath string) []float64 {
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

func main() {
	f := openLog()
	defer f.Close()

	lines := openFile("./location.csv")
	var stdist, staz, hfdist, hfaz []float64

	for i := 0; i < len(lines)-2; i += 2 {
		p1 := geoPoint{lines[i], lines[i+1]}
		p2 := geoPoint{lines[i+2], lines[i+3]}

		st, az1 := sphericalTrigonometry(p1, p2)
		hf, az2 := hubenyFormula(p1, p2)

		stdist = append(stdist, st)
		staz = append(staz, az1)
		hfdist = append(hfdist, hf)
		hfaz = append(hfaz, az2)

		fmt.Println("始点：", p1, "終点：", p2)
		fmt.Println("球面三角法 距離：", st, "km 方位角：", az1)
		fmt.Println("Hubenyの公式 距離：", hf, "km 方位角：", az2)
		fmt.Println("距離差異：", math.Abs(st-hf), "km 方位角差異：", math.Abs(az1-az2))
	}

}
