package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/stilldavid/gopro-utils/telemetry"
	"github.com/tkrajina/gpxgo/gpx"
)

func main() {
	gpxData := new(gpx.GPX)

	videoName := flag.String("v", "", "Required: video file to read timestamp")
	inName := flag.String("i", "", "Required: telemetry file to read")
	outName := flag.String("o", "", "Required: gpx file to write")
	flag.Parse()


	if *videoName == "" {
		flag.Usage()
		return
	}

	/*
	// Opening the file handle
	f, err := os.Open(*videoName)
	if err != nil {
		fmt.Printf("Cannot access video file %s.\n", *videoName)
		return
	}
	defer f.Close()

	// Checking if the opened handle is really a file
	statinfo, err := f.Stat()
	if err != nil {
		fmt.Printf("stat() failure on %s.\n", *videoName)
		return
	}
	duration := statinfo.ModTime().Unix()

	fmt.Printf("duration: %d.\n", duration)
	*/


	lastTime := time.Now()
	fmt.Printf("lastTime start: %s\n", lastTime)


	if *inName == "" {
		flag.Usage()
		return
	}

	telemFile, err := os.Open(*inName)
	if err != nil {
		fmt.Printf("Cannot access telemetry file %s.\n", *inName)
		os.Exit(1)
	}
	defer telemFile.Close()

	t := &telemetry.TELEM{}
	t_prev := &telemetry.TELEM{}

	track := new(gpx.GPXTrack)
	segment := new(gpx.GPXTrackSegment)

	for {
		t, err = telemetry.Read(telemFile)
		if err != nil && err != io.EOF {
			fmt.Println("Error reading telemetry file", err)
			os.Exit(1)
		} else if err == io.EOF || t == nil {
			break
		}

		// first full, guess it's about a second
		if t_prev.IsZero() {
			*t_prev = *t
			t.Clear()
			continue
		}

		// process until t.Time
		t_prev.FillTimes(t.Time.Time)
		telems := t_prev.ShitJson()


		for i, _ := range telems {
			segment.AppendPoint(
				&gpx.GPXPoint{
					Point: gpx.Point{
						Latitude:  telems[i].Latitude,
						Longitude: telems[i].Longitude,
						Elevation: *gpx.NewNullableFloat64(telems[i].Altitude),
					},
					Timestamp: time.Unix(telems[i].TS/1000/1000, telems[i].TS%(1000*1000)),
				},
			)

			//fmt.Printf("raw time: %s, %s\n", telems[i].TS/1000/1000, telems[i].TS%(1000*1000))
			//fmt.Printf("Timestamp: %s\n", time.Unix(telems[i].TS/1000/1000, telems[i].TS%(1000*1000)))
			lastTime = time.Unix(telems[i].TS/1000/1000, telems[i].TS%(1000*1000))
		}

		*t_prev = *t
		t = &telemetry.TELEM{}
	}

	fmt.Printf("lastTime end: %s\n", lastTime)

	if err := os.Chtimes(*videoName, time.Now(), lastTime); err != nil {
		fmt.Printf("os.Chtimes() failure: %s.\n", err)
	}


	track.AppendSegment(segment)
	gpxData.AppendTrack(track)

	gpxFile, err := os.Create(*outName)
	if err != nil {
		fmt.Printf("Cannot make output file %s.\n", *outName)
		os.Exit(1)
	}

	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			fmt.Printf("Cannot close gpx file %s: %s", file.Name(), err)
			os.Exit(1)
		}
	}(gpxFile)

	xml, err := gpxData.ToXml(gpx.ToXmlParams{Version: "1.1", Indent: true})
	gpxFile.Write(xml)
}
