package utils

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestExecuteAt(t *testing.T) {
	testHour := 3
	testDay := 2
	testMonth := 1
	testTime := time.Date(2024, time.Month(testMonth), testDay, testHour, 0, 0, 0, time.Local)

	// If the time is further in the day, return time object that is appropriately that hour 0mins 0secs
	resultTime := executeAtXMath(testHour+1, testTime)
	expectedTime := time.Date(2024, testTime.Month(), testDay, testHour+1, 0, 0, 0, time.Local)
	assert.True(t, resultTime.Equal(expectedTime))
	//Ensure the logic used in the sleep function is sound, sleep for one hour
	assert.Equal(t, resultTime.Sub(testTime), time.Hour)

	// If the specefied hour has already passed for the day, return a time object that is one day ahead
	resultTime = executeAtXMath(testHour-1, testTime)
	expectedTime = time.Date(2024, testTime.Month(), testDay+1, testHour-1, 0, 0, 0, time.Local)
	assert.True(t, resultTime.Equal(expectedTime))
	//Ensure the logic used in the sleep function is sound, sleep for 23 hours
	assert.Equal(t, resultTime.Sub(testTime), time.Hour*23)

}
