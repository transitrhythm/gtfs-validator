package main

import (
	"fmt"
	"github.com/patrickbr/gtfsparser" // "github.com/geops/gtfsparser" //
	"os"
)

const floatPrecision = 3

func test(err error) {
	if err != nil {
		fmt.Println(err)
	}
}
func main() {
	if len(os.Args) != 2 {
		fmt.Printf("Usage : %s <ZIPfile> \n", os.Args[0])
		os.Exit(0)
	}
	// Identify the input GTFS Zip file
	zipFile := os.Args[1]

	// Parse the GTFS Zip file
	feed := gtfsparser.NewFeed()
	feed.Parse(zipFile)
	fmt.Printf("Done, parsed: %d feeds, %d agencies, %d stops, %d routes, %d trips, %d shapes\n\n",
		len(feed.FeedInfos), len(feed.Agencies), len(feed.Stops), len(feed.Routes), len(feed.Trips), len(feed.Shapes))

	feeds, err := ValidateFeeds(feed, "feeds.csv")
	test(err)
	for _, v := range feeds {
		if v.Active == true {
			agencies, err := ValidateAgencies(feed, "agency.csv")
			test(err)
			routes, err := ValidateRoutes(feed, "routes.csv")
			test(err)
			stops, err := ValidateStops(feed, floatPrecision, "stops.csv")
			test(err)
			shapes, err := ValidateShapes(feed, floatPrecision, "shapes.csv")
			test(err)
			trips, err := ValidateTrips(feed, "trips.csv")
			test(err)
			stopTimes, err := ValidateStopTimes(feed, floatPrecision, "stoptimes.csv")
			test(err)
			if !agencies || !routes || !stops || !shapes || !trips || !stopTimes {
				fmt.Println("Fail")
			}
		}
	}
}
