package main

import (
	"fmt"
	"math"
	"os"
	"sort"
	"time"

	"github.com/im7mortal/UTM"
	"github.com/patrickbr/gtfsparser" // "github.com/geops/gtfsparser" //
	"github.com/patrickbr/gtfsparser/gtfs"
)

// GtfsFeedInfo -
type GtfsFeedInfo struct {
	PublisherName string
	PublisherURL  string
	Lang          string
	StartDate     gtfs.Date
	EndDate       gtfs.Date
	Phone         string
	Version       string
	Active        bool
}

func toGtfsDate(year int, month time.Month, day int) (date gtfs.Date) {
	date.Year = int16(year)
	date.Month = int8(month)
	date.Day = int8(day)
	return date
}

// ValidateFeeds -
func ValidateFeeds(feed *gtfsparser.Feed, filename string) (feedInfo []GtfsFeedInfo, err error) {
	file, err := os.Create(filename)
	test(err)
	defer file.Close()

	file.WriteString(fmt.Sprintln("ParseFeeds:"))
	today := toGtfsDate(time.Now().Date())
	for k, v := range feed.FeedInfos {
		feedData := GtfsFeedInfo{}
		feedData.PublisherName = v.Publisher_name
		feedData.Version = v.Version
		feedData.StartDate = v.Start_date
		feedData.EndDate = v.End_date
		file.WriteString(fmt.Sprintf("[%d] %s : <%s> - %s - %s\n", k, v.Version, v.Publisher_name, Datestamp(feedData.StartDate), Datestamp(feedData.EndDate)))
		feedData.Active = inDateRange(feedData.StartDate, today, feedData.EndDate)
		feedInfo = append(feedInfo, feedData)
	}
	return feedInfo, err
}

// ValidateAgencies -
func ValidateAgencies(feed *gtfsparser.Feed, filename string) (valid bool, err error) {

	return valid, err
}

// ValidateRoutes -
func ValidateRoutes(feed *gtfsparser.Feed, filename string) (valid bool, err error) {

	return valid, err
}

// ValidateStops -
func ValidateStops(feed *gtfsparser.Feed, floatPrecision int, filename string) (valid bool, err error) {

	return valid, err
}

func inRange(min, value, max float64) bool {
	return min <= value && value <= max
}

// BySequence -
type BySequence struct{ gtfs.ShapePoints }

// ShapePoints -
type ShapePoints []gtfs.ShapePoint

// Len -
func (shapePoints ShapePoints) Len() int {
	return len(shapePoints)
}

// Less -
func (s BySequence) Less(i, j int) bool {
	return s.ShapePoints[i].Sequence < s.ShapePoints[j].Sequence
}

// Swap -
func (shapePoints ShapePoints) Swap(i, j int) {
	shapePoints[i], shapePoints[j] = shapePoints[j], shapePoints[i]
}

func toLocation(lat float32, lon float32, precision int) (location Cartesian, timeZone int, timeLetter string, err error) {
	x, y, timeZone, timeLetter, err := UTM.FromLatLon(float64(lat), float64(lon), lat > 0)
	location.x = toFixed(x, precision)
	location.y = toFixed(y, precision)
	return location, timeZone, timeLetter, err
}

const minimumSegmentLength = 0.001
const maximumSegmentLength = 1.000
const maximumSegmentError = 0.01
const maximumLengthVariance = 1.0

// ValidateShapes -
func ValidateShapes(feed *gtfsparser.Feed, precision int, filename string) (valid bool, err error) {
	file, err := os.Create(filename)
	test(err)
	defer file.Close()

	file.WriteString(fmt.Sprintln("ValidateShapes:"))
	file.WriteString(fmt.Sprintln("shape_id, shape_pt_lat, shape_pt_lon, shape_pt_sequence, shape_dist_traveled, shape_segment_length, shape_location_east, shape_location_north, calc_dist_traveled, calc_dist_variance, calc_segment_length, calc_segment_variance, sphere_dist_traveled, sphere_dist_variance, spherical_length, spherical_length_variance, shape_pt_name"))
	for _, v := range feed.Shapes {
		// fmt.Println("Shape:", v.Id)
		// Sort by sequence number
		sort.Sort(BySequence{v.Points})
		sequenceNumber := 0
		var calculatedDistTraveled, sphericalDistTraveled float64
		var previousDistTraveled float32
		var previousLocation Cartesian
		var lastCoord Coord
		for _, v1 := range v.Points {
			// Check for lat range
			if !inRange(-90.0, float64(v1.Lat), 90.0) {
				file.WriteString(fmt.Sprintln("Out of range: Lat - [", v.Id, v1.Sequence, "]=", v1.Lat))
			}
			// Check for lon range
			if !inRange(-180.0, float64(v1.Lon), 180.0) {
				file.WriteString(fmt.Sprintln("Out of range: Lon - [", v.Id, v1.Sequence, "]=", v1.Lon))
			}
			// Check for consecutive sequence number
			sequenceDifference := v1.Sequence - sequenceNumber
			if sequenceDifference != 1 {
				file.WriteString(fmt.Sprintln("Out of sequence: [", v.Id, v1.Sequence, "]"))
				if sequenceDifference == 0 {
					file.WriteString(fmt.Sprintf("Duplicate"))
				} else {
					file.WriteString(fmt.Sprintf("Missing: %d", sequenceDifference))
				}
			}
			// If shape distance traveled available, check for validity against lat/lon
			location, _, _, _ := toLocation(v1.Lat, v1.Lon, 1)
			if v1.Has_dist {
				if v1.Sequence == 1 {
					calculatedDistTraveled = 0.0
					previousLocation = location
					previousDistTraveled = 0.0
					sphericalDistTraveled = 0.0
					lastCoord = toCoord(v1.Lat, v1.Lon, 6)
					file.WriteString(fmt.Sprintf("%s, %.6f, %.6f, %d, %.4f\n", v.Id, v1.Lat, v1.Lon, v1.Sequence, v1.Dist_traveled))

				} else {
					calculatedSegmentLength := toKm(hypotenuse(previousLocation.x-location.x, previousLocation.y-location.y, precision), 4)
					shapeSegmentLength := toFixed(float64(v1.Dist_traveled-previousDistTraveled), 4)
					if !inRange(minimumSegmentLength, shapeSegmentLength, maximumSegmentLength) {
						file.WriteString(fmt.Sprintln("Segment length violation: [", v.Id, ",", v1.Sequence, "]", toFixed(shapeSegmentLength, precision)))
						if minimumSegmentLength > shapeSegmentLength {
							file.WriteString(fmt.Sprintf("Segment length less than minimum %.4f km - Suggestion: Remove unnecessary shape point\n", minimumSegmentLength))
						} else {
							file.WriteString(fmt.Sprintf("Segment length greater than maximum %.4f km - Suggestion: Add intermediate shape points as necessary\n", maximumSegmentLength))
						}
					}
					newCoord := toCoord(v1.Lat, v1.Lon, 6)
					haversineDistance := toKm(HaversineFormula(newCoord, lastCoord, 4), 4)
					sphericalDistance := toKm(SphericalDistance(newCoord, lastCoord, 4), 4)
					sphericalDistTraveled += sphericalDistance
					calculatedDistTraveled += calculatedSegmentLength
					percent1 := PercentDiff(shapeSegmentLength, calculatedSegmentLength, precision)
					percent2 := PercentDiff(float64(v1.Dist_traveled), calculatedDistTraveled, precision)
					percent3 := PercentDiff(shapeSegmentLength, sphericalDistance, precision)
					percent4 := PercentDiff(float64(v1.Dist_traveled), sphericalDistTraveled, precision)
					percent5 := PercentDiff(shapeSegmentLength, haversineDistance, precision)
					if math.Abs(percent1) > maximumLengthVariance && math.Abs(percent3) > maximumLengthVariance {
						file.WriteString(fmt.Sprintln("Segment length variation violation: [", v.Id, ",", v1.Sequence, "]", shapeSegmentLength, "km", percent1, "%", percent3, "% - Suggestion: Recalculate shape_distance_traveled"))
					}
					if math.Abs(percent2) > maximumLengthVariance && math.Abs(percent4) > maximumLengthVariance {
						file.WriteString(fmt.Sprintln("Shape distance traveled variation violation: [", v.Id, ",", v1.Sequence, "]", v1.Dist_traveled, "km", percent2, "%", percent4, "% - Suggestion: Recalculate shape_distance_traveled"))
					}
					file.WriteString(fmt.Sprintf("%s, %.6f, %.6f, %d, %.4f, %.4f, %.1f, %.1f, %.4f, %.1f, %.4f, %.1f, %.4f, %.1f, %.4f, %.1f, %.4f, %.1f\n", v.Id, v1.Lat, v1.Lon, v1.Sequence, v1.Dist_traveled, shapeSegmentLength, location.x, location.y, calculatedDistTraveled, percent2, calculatedSegmentLength, percent1, sphericalDistTraveled, percent4, sphericalDistance, percent3, haversineDistance, percent5))
					previousDistTraveled = v1.Dist_traveled
					previousLocation = location
					lastCoord = newCoord
				}
			}
			sequenceNumber++
		}
	}

	return valid, err
}

// ValidateTrips -
func ValidateTrips(feed *gtfsparser.Feed, filename string) (valid bool, err error) {

	return valid, err
}

// ValidateStopTimes -
func ValidateStopTimes(feed *gtfsparser.Feed, floatPrecision int, filename string) (valid bool, err error) {

	return valid, err
}
