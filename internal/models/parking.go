package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Point struct {
	Type        string    `json:"type" bson:"type"`
	Coordinates []float64 `json:"coordinates" bson:"coordinates"`
}

type VehicleType string

const (
	PAS  VehicleType = "PAS"
	COMM VehicleType = "COMM"
)

type MeterHours struct {
	StartTime time.Time `json:"startTime" bson:"startTime"`
	EndTime   time.Time `json:"endTime" bson:"endTime"`
}

type MeterDays struct {
	Monday    bool `json:"monday" bson:"monday"`
	Tuesday   bool `json:"tuesday" bson:"tuesday"`
	Wednesday bool `json:"wednesday" bson:"wednesday"`
	Thursday  bool `json:"thursday" bson:"thursday"`
	Friday    bool `json:"friday" bson:"friday"`
	Saturday  bool `json:"saturday" bson:"saturday"`
	Sunday    bool `json:"sunday" bson:"sunday"`
}

type ParkingMeter struct {
	ID           primitive.ObjectID `json:"_id" bson:"_id,omitempty"`
	ObjectID     int                `json:"objectId" bson:"objectId"`
	MeterNumber  string             `json:"meterNumber" bson:"meter_number"`
	Status       string             `json:"status" bson:"status"`
	PayByCell    string             `json:"payByCell" bson:"pay_by_cell"`
	VehicleType  VehicleType        `json:"vehicleType" bson:"vehicle_type"`
	Duration     int                `json:"duration" bson:"duration"`
	MeterDays    MeterDays          `json:"meterDays" bson:"meter_days"`
	MeterHours   MeterHours         `json:"meterHours" bson:"meter_hours"`
	Facility     bool               `json:"facility" bson:"facility"`
	FacilityName string             `json:"facilityName" bson:"facility_name"`
	Borough      string             `json:"borough" bson:"borough"`
	OnStreet     string             `json:"onStreet" bson:"on_street"`
	FromStreet   string             `json:"fromStreet" bson:"from_street"`
	ToStreet     string             `json:"toStreet" bson:"to_street"`
	SideOfStreet string             `json:"sideOfStreet" bson:"side_of_street"`
	Location     Point              `json:"location" bson:"location"`
}
