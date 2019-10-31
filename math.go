package main

import (
	"fmt"
	"math"

	"github.com/patrickbr/gtfsparser/gtfs"
)

// Degrees -
type Degrees float64

// PolarPoint -
type PolarPoint struct {
	distance float64
	bearing  Degrees
}

// Cartesian -
type Cartesian struct {
	x, y float64
}

func hypotenuse(x, y float64, precision int) float64 {
	return toFixed(math.Sqrt(x*x+y*y), precision)
}

// Degrees -
func toDegrees(x, y float64, precision int) Degrees {
	return Degrees(toFixed(math.Mod((math.Atan2(x, y)*180/math.Pi)+360, 360), precision))
}

func toRadians(degrees Degrees) (radians float64) {
	radians = float64(degrees * math.Pi / 180)
	return radians
}

func round(num float64) int {
	return int(num + math.Copysign(0.5, num))
}

func toFixed(num float64, precision int) float64 {
	output := math.Pow(10, float64(precision))
	return float64(round(num*output)) / output
}

// PercentDiff -
func PercentDiff(x, y float64, precision int) (value float64) {
	if toFixed(y, precision) == 0.0 {
		value = 0
	} else {
		value = (x - y) / y
	}
	return toFixed(value*100, precision)
}

func toKm(distance float64, precision int) float64 {
	return toFixed(distance/1000, precision)
}

// Coord -
type Coord struct {
	Lat float64
	Lon float64
}

func toCoord(lat float32, lon float32, precision int) (coord Coord) {
	coord.Lat = toFixed(float64(lat), precision)
	coord.Lon = toFixed(float64(lon), precision)
	return coord
}

// SphericalDistance -
func SphericalDistance(a Coord, b Coord, precision int) (metres float64) {
	cosine := math.Cos(toRadians(Degrees(a.Lat)))
	degrees := hypotenuse(a.Lat-b.Lat, (a.Lon-b.Lon)*cosine, precision*2)
	metres = toFixed(degrees*111111, precision)
	return metres
}

// HaversineFormula -
func HaversineFormula(p1, p2 Coord, precision int) (metres float64) {
	R := 6371e3 // metres
	φ1 := toRadians(Degrees(p1.Lat))
	φ2 := toRadians(Degrees(p2.Lat))
	Δφ := toRadians(Degrees(p2.Lat - p1.Lat))
	Δλ := toRadians(Degrees(p2.Lon - p1.Lon))
	a := math.Sin(Δφ/2)*math.Sin(Δφ/2) + math.Cos(φ1)*math.Cos(φ2)*math.Sin(Δλ/2)*math.Sin(Δλ/2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
	metres = toFixed(R*c, precision)
	return metres
}

func isLeapYear(year int) bool {
	return (year%4 == 0) && !(year%100 == 0)
}

// DayOfYear -
func DayOfYear(date gtfs.Date) (dayOfYear int) {
	daysOfYear := []int{0, 31, 59, 90, 120, 151, 181, 212, 242, 273, 304, 334}
	dayOfYear = daysOfYear[date.Month] + int(date.Day)
	if date.Month >= 3 && isLeapYear(int(date.Year)) {
		dayOfYear++
	}
	return dayOfYear
}

func daysSince(startDate, endDate gtfs.Date) (days int) {
	startDayOfYear := DayOfYear(startDate)
	endDayOfYear := DayOfYear(endDate)
	yearsDiff := int(endDate.Year - startDate.Year)
	days = (endDayOfYear - startDayOfYear) + (yearsDiff * 365) + (yearsDiff / 4)
	return days
}

func inDateRange(startDate, today, endDate gtfs.Date) (state bool) {
	return daysSince(startDate, today) < daysSince(startDate, endDate)
}

func inWeekdayRange(weekday int, days [7]bool) (state bool) {
	return days[weekday]
}

//Timestamp -
func Timestamp(a gtfs.Time) (timestamp string) {
	return fmt.Sprintf("%02d:%02d:%02d", a.Hour, a.Minute, a.Second)
}

//Datestamp -
func Datestamp(a gtfs.Date) (datestamp string) {
	return fmt.Sprintf("%04d-%02d-%02d", a.Year, a.Month, a.Day)
}
