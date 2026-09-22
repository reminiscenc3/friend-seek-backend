package models

import "time"

type LoginRequest struct {
	Login string `json:"login"`
}

type UpdateNoteRequest struct {
	Note string `json:"note"`
}

type ReportLocationRequest struct {
	Latitude  *float64   `json:"lat"`
	Longitude *float64   `json:"lon"`
	Azimuth   *float64   `json:"azimuth"`
	Accuracy  float64    `json:"accuracy"`
	Timestamp *time.Time `json:"ts"`
}

func (r *ReportLocationRequest) Location() (Location, error) {
	if r.Latitude == nil || r.Longitude == nil {
		return Location{}, ErrNoCoordinates
	}

	loc := Location{
		Latitude:  *r.Latitude,
		Longitude: *r.Longitude,
		Accuracy:  r.Accuracy,
	}

	if !loc.Valid() {
		return Location{}, ErrInvalidLocation
	}

	if r.Azimuth != nil {
		loc.Azimuth = *r.Azimuth

		if !loc.NormalizeAzimuth() {
			return Location{}, ErrInvalidAzimuth
		}
	}

	if r.Timestamp != nil {
		loc.UpdatedAt = r.Timestamp.UTC()
	}

	return loc, nil
}

func (r *ReportLocationRequest) Validate() error {
	_, err := r.Location()

	return err
}
