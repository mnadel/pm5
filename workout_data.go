package main

import (
	"time"

	log "github.com/sirupsen/logrus"
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

type WorkoutType int

// WorkoutData
type RawWorkoutData struct {
	LogEntry          []byte
	ElapsedTime       []byte
	Distance          []byte
	AverageStrokeRate byte
	AvgDragFactor     byte
	WorkoutType       byte
	AvgPace           []byte
}

// WorkoutData
type WorkoutData struct {
	LogEntry          time.Time
	ElapsedTime       time.Duration
	Distance          float32
	AverageStrokeRate uint8
	AvgDragFactor     uint8
	WorkoutType       WorkoutType
	AvgPace           time.Duration
}

func ReadWorkoutData(data []byte) *RawWorkoutData {
	return &RawWorkoutData{
		LogEntry:          data[0:4],
		ElapsedTime:       data[4:7],
		Distance:          data[7:10],
		AverageStrokeRate: data[10],
		AvgDragFactor:     data[15],
		WorkoutType:       data[17],
		AvgPace:           data[18:20],
	}
}

func (rd *RawWorkoutData) Decode() *WorkoutData {
	tz, err := time.LoadLocation("America/Chicago")
	if err != nil {
		log.Fatal("cannot load timezone")
	}

	decodedLogEntry := time.Date(
		time.Now().Year(), 0, 0, // y m d
		int(rd.LogEntry[3]), int(rd.LogEntry[2]), 0, 0, // h m s ns
		tz)

	return &WorkoutData{
		LogEntry:          decodedLogEntry,
		ElapsedTime:       DecodeDuration(float32(DecodeThreeByteNumber(rd.ElapsedTime)), 0.01),
		Distance:          float32(DecodeThreeByteNumber(rd.Distance)) * 0.1,
		AverageStrokeRate: DecodeByteNumber(rd.AverageStrokeRate),
		AvgDragFactor:     DecodeByteNumber(rd.AvgDragFactor),
		WorkoutType:       WorkoutType(int(rd.WorkoutType)),
		AvgPace:           DecodeDuration(float32(DecodeTwoByteNumber(rd.AvgPace)), 0.1),
	}
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