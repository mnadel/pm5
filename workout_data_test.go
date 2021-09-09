package main

import (
	"testing"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestDecodeRealWorkoutData(t *testing.T) {
	// 2021-08-16 18:22 raw="&{[8 43 22 18] [184 96 0] [74 31 0] 0 0 1 [9 6]}"
	raw := &RawWorkoutData{
		LogEntry:          []byte{8, 43, 22, 18},
		ElapsedTime:       []byte{184, 96, 0},
		Distance:          []byte{74, 31, 0},
		AverageStrokeRate: byte(0),
		AvgDragFactor:     byte(0),
		WorkoutType:       byte(1),
		AvgPace:           []byte{9, 6},
	}
	parsed := raw.Decode()

	assert.Equal(t, MustParseDateTime("2021-08-16 18:22 CDT"), parsed.LogEntry)
	assert.Equal(t, MustParseDuration("4m7.6s"), parsed.ElapsedTime)
	assert.Equal(t, float32(801.0), parsed.Distance)
	assert.Equal(t, WorkoutTypeJustRowSplits, parsed.WorkoutType)
	assert.Equal(t, MustParseDuration("2m34.5s"), parsed.AvgPace)
}

func TestDecodeRealWorkoutData2(t *testing.T) {
	// 2021-09-05 17:39 raw="&{[89 42 38 17] [162 93 0] [20 30 0] 28 122 1 [20 6]}"
	raw := &RawWorkoutData{
		LogEntry:          []byte{89, 42, 38, 17},
		ElapsedTime:       []byte{162, 93, 0},
		Distance:          []byte{20, 30, 0},
		AverageStrokeRate: byte(28),
		AvgDragFactor:     byte(122),
		WorkoutType:       byte(1),
		AvgPace:           []byte{20, 6},
	}
	parsed := raw.Decode()

	assert.Equal(t, MustParseDateTime("2021-09-05 17:38 CDT"), parsed.LogEntry)
	assert.Equal(t, MustParseDuration("3m59.7s"), parsed.ElapsedTime)
	assert.Equal(t, float32(770.0), parsed.Distance)
	assert.Equal(t, WorkoutTypeJustRowSplits, parsed.WorkoutType)
	assert.Equal(t, MustParseDuration("2m35.6s"), parsed.AvgPace)
}

func testWorkoutData() []byte {
	return []byte{
		8,  // Log Entry Date Lo, -- 0
		43, // Log Entry Date Hi, -- 1
		3,  // Log Entry Time Lo, -- 2
		4,  // Log Entry Time Hi, -- 3
		5,  // Elapsed Time Lo (0.01 sec lsb), -- 4
		6,  // Elapsed Time Mid, -- 5
		7,  // Elapsed Time High, -- 6
		8,  // Distance Lo (0.1 m lsb), -- 7
		9,  // Distance Mid, -- 8
		10, // Distance High, -- 9
		11, // Average Stroke Rate, -- 10
		12, // Ending Heart Rate, -- 11
		13, // Average Heart Rate, -- 12
		14, // Min Heart Rate, -- 13
		15, // Max Heartrate, -- 14
		16, // Drag Factor Average, -- 15
		17, // Recovery Heart Rate, (zero = not valid data. After 1 minute of rest/recovery, PM5 sends this data as a revised End Of Workout summary data characteristic unless the monitor has been turned off or a new workout started) -- 16
		7,  // Workout Type, -- 17 (NB there aren't 17 workout types, using a smaller byte for this)
		19, // Avg Pace Lo (0.1 sec lsb) -- 18
		20, // Avg Pace Hi -- 19
	}
}

func TestReadWorkoutData(t *testing.T) {
	raw := ReadWorkoutData(testWorkoutData())
	assert.Equal(t, []byte{8, 43, 3, 4}, raw.LogEntry)
	assert.Equal(t, []byte{5, 6, 7}, raw.ElapsedTime)
	assert.Equal(t, []byte{8, 9, 10}, raw.Distance)
	assert.Equal(t, byte(11), raw.AverageStrokeRate)
	assert.Equal(t, byte(16), raw.AvgDragFactor)
	assert.Equal(t, byte(7), raw.WorkoutType)
	assert.Equal(t, []byte{19, 20}, raw.AvgPace)
}

func TestDecodeWorkoutData(t *testing.T) {
	raw := ReadWorkoutData(testWorkoutData())
	parsed := raw.Decode()

	assert.Equal(t, float32(65767.2), parsed.Distance)                 // 657672*0.1 = 65767.2
	assert.Equal(t, MustParseDuration("4602.93s"), parsed.ElapsedTime) // 460293*0.01 = 4602.93s
	assert.Equal(t, uint8(11), parsed.AverageStrokeRate)
	assert.Equal(t, uint8(16), parsed.AvgDragFactor)
	assert.Equal(t, WorkoutTypeFixedDistInterval, parsed.WorkoutType)
	assert.Equal(t, MustParseDuration("513.9s"), parsed.AvgPace) // 5139*0.1 = 513.9s
}

func TestMustParseDateTime(t *testing.T) {
	dt := MustParseDateTime("2021-09-05 17:38 CDT")

	assert.Equal(t, 2021, dt.Year())
	assert.Equal(t, time.Month(9), dt.Month())
	assert.Equal(t, 5, dt.Day())
	assert.Equal(t, 17, dt.Hour())
	assert.Equal(t, 38, dt.Minute())
}

func MustParseDuration(formatted string) time.Duration {
	duration, err := time.ParseDuration(formatted)
	if err != nil {
		log.WithError(err).WithField("dt", formatted).Fatalf("cannot parse duration")
	}
	return duration
}

func MustParseDateTime(formatted string) time.Time {
	t, err := time.Parse("2006-01-02 15:04 MST", formatted)
	if err != nil {
		log.WithError(err).WithField("dt", formatted).Fatal("cannot parse time")
	}
	return t.Local()
}
