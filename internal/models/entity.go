package models

import (
	"math"
	"time"
)

type User struct {
	Login     string    `json:"login"`
	Note      string    `json:"note"`
	CreatedAt time.Time `json:"created_at"`
}

type Location struct {
	Latitude  float64   `json:"lat"`
	Longitude float64   `json:"lon"`
	Azimuth   float64   `json:"azimuth"`
	Accuracy  float64   `json:"accuracy"`
	UpdatedAt time.Time `json:"ts"`
}

func (l *Location) Valid() bool {
	if math.IsNaN(l.Latitude) || math.IsNaN(l.Longitude) || math.IsInf(l.Latitude, 0) || math.IsInf(l.Longitude, 0) {
		return false
	}

	return l.Latitude >= -90 && l.Latitude <= 90 && l.Longitude >= -180 && l.Longitude <= 180
}

func (l *Location) NormalizeAzimuth() bool {
	if math.IsNaN(l.Azimuth) || math.IsInf(l.Azimuth, 0) {
		return false
	}

	l.Azimuth = math.Mod(math.Mod(l.Azimuth, 360)+360, 360)

	l.Azimuth = math.Round(l.Azimuth*10) / 10
	return true
}

type TargetLocation struct {
	Login  string    `json:"login"`
	Note   string    `json:"note"`
	NextAt time.Time `json:"next_at"`

	Location
}
