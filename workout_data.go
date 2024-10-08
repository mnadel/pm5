package main

import (
	"encoding/json"
	"time"
)

const (
	WorkoutTypeJustRowNoSplits               = WorkoutType(0)
	WorkoutTypeJustRowSplits                 = WorkoutType(1)
	WorkoutTypeFixedDistNoSplits             = WorkoutType(2)
	WorkoutTypeFixedDistSplits               = WorkoutType(3)
	WorkoutTypeFixedTimeNoSplits             = WorkoutType(4)
	WorkoutTypeFixedTimeSplits               = WorkoutType(5)
	WorkoutTypeFixedTimeInterval             = WorkoutType(6)
	WorkoutTypeFixedDistInterval             = WorkoutType(7)
	WorkoutTypeVariableInterval              = WorkoutType(8)
	WorkoutTypeVariableUndefinedRestInterval = WorkoutType(9)
	WorkoutTypeFixedCalorie                  = WorkoutType(10)
	WorkoutTypeFixedWattMinutes              = WorkoutType(11)
	WorkoutTypeFixedCalsInterval             = WorkoutType(12)
	WorkoutTypeNum                           = WorkoutType(13)
)

const (
	ElapsedTimeFactor = 0.01
	AvgPaceFactor     = 0.1
)

type WorkoutType int

// Raw Workout (0x39) bytes
type RawWorkoutData struct {
	Data              []byte
	LogEntry          []byte
	ElapsedTime       []byte
	Distance          []byte
	AverageStrokeRate byte
	AvgDragFactor     byte
	WorkoutType       byte
	AvgPace           []byte
}

// Decoded Workout (0x39) data
type WorkoutData struct {
	LogEntry          time.Time
	ElapsedTime       time.Duration
	Distance          float32
	AverageStrokeRate uint8
	AvgDragFactor     uint8
	WorkoutType       WorkoutType
	AvgPace           time.Duration
}

// ReadWorkoutData reads an array of bytes into a raw struct of Workout data.
func ReadWorkoutData(data []byte) *RawWorkoutData {
	return &RawWorkoutData{
		Data:              data,
		LogEntry:          data[0:4],
		ElapsedTime:       data[4:7],
		Distance:          data[7:10],
		AverageStrokeRate: data[10],
		AvgDragFactor:     data[15],
		WorkoutType:       data[17],
		AvgPace:           data[18:20],
	}
}

// Decode returns a parsed/decoded set of bytes into a WorkoutData.
func (rd *RawWorkoutData) Decode() *WorkoutData {
	return &WorkoutData{
		LogEntry:          DecodeDateTime(rd.LogEntry),
		ElapsedTime:       DecodeDuration(float32(DecodeThreeByteNumber(rd.ElapsedTime)), ElapsedTimeFactor),
		Distance:          float32(DecodeThreeByteNumber(rd.Distance)) * 0.1,
		AverageStrokeRate: DecodeByteNumber(rd.AverageStrokeRate),
		AvgDragFactor:     DecodeByteNumber(rd.AvgDragFactor),
		WorkoutType:       WorkoutType(int(rd.WorkoutType)),
		AvgPace:           DecodeDuration(float32(DecodeTwoByteNumber(rd.AvgPace)), AvgPaceFactor),
	}
}

func (wd *WorkoutData) AsJSON() string {
	var m map[string]interface{}

	b, _ := json.Marshal(wd)
	json.Unmarshal(b, &m)

	m["ElapsedTime"] = uint64(wd.ElapsedTime.Seconds() / ElapsedTimeFactor)
	m["AvgPace"] = uint64(wd.ElapsedTime.Seconds() / AvgPaceFactor)

	b, _ = json.Marshal(m)

	return string(b)
}

func (wt WorkoutType) AsString() string {
	var t string

	switch wt {
	case WorkoutTypeJustRowNoSplits:
	case WorkoutTypeJustRowSplits:
		t = "JustRow"
	case WorkoutTypeFixedDistSplits:
		t = "FixedDistanceSplits"
	case WorkoutTypeFixedDistInterval:
		t = "FixedDistanceInterval"
	case WorkoutTypeFixedTimeSplits:
		t = "FixedTimeSplits"
	case WorkoutTypeFixedTimeInterval:
	case WorkoutTypeFixedTimeNoSplits:
		t = "FixedTimeInterval"
	case WorkoutTypeFixedCalsInterval:
	case WorkoutTypeFixedCalorie:
		t = "FixedCalorie"
	case WorkoutTypeVariableInterval:
		t = "VariableInterval"
	case WorkoutTypeVariableUndefinedRestInterval:
		t = "VariableIntervalUndefinedRest"
	default:
		t = "unknown"
	}

	return t
}

/*
	1, // Log Entry Date Lo, -- 0
	2, // Log Entry Date Hi, -- 1
	3, // Log Entry Time Lo, -- 2
	4, // Log Entry Time Hi, -- 3
	5, // Elapsed Time Lo (0.01 sec lsb), -- 4
	6, // Elapsed Time Mid, -- 5
	7, // Elapsed Time High, -- 6
	8, // Distance Lo (0.1 m lsb), -- 7
	9, // Distance Mid, -- 8
	10, // Distance High, -- 9
	11, // Average Stroke Rate, -- 10
	12, // Ending Heart Rate, -- 11
	13, // Average Heart Rate, -- 12
	14, // Min Heart Rate, -- 13
	15, // Max Heartrate, -- 14
	16, // Drag Factor Average, -- 15
	17, // Recovery Heart Rate, (zero = not valid data. After 1 minute of rest/recovery, PM5 sends this data as a revised End Of Workout summary data characteristic unless the monitor has been turned off or a new workout started) -- 16
	18, // Workout Type, -- 17
	19, // Avg Pace Lo (0.1 sec lsb) -- 18
	20, // Avg Pace Hi -- 19
*/
