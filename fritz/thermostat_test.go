package fritz

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestFormattingOfTemperaturesRegularRange tests formatting of the temperature values obtained by AHA interface.
func TestFormattingOfTemperaturesRegularRange(t *testing.T) {
	th := Thermostat{Measured: "47", Saving: "40", Goal: "42", Comfort: "44"}
	assert.Equal(t, "23.5", th.FmtMeasuredTemperature())
	assert.Equal(t, "20", th.FmtSavingTemperature())
	assert.Equal(t, "21", th.FmtGoalTemperature())
	assert.Equal(t, "22", th.FmtComfortTemperature())
	assert.Equal(t, 1, th.State())
}

// TestFormattingOfTemperaturesRegularRange tests formatting of the temperature values obtained by AHA interface.
func TestFormattingOfTemperaturesParseError(t *testing.T) {
	th := Thermostat{Measured: "assafsa", Saving: "dghdafhf", Goal: "dfahfh", Comfort: "rheeh"}
	assert.Equal(t, "", th.FmtMeasuredTemperature())
	assert.Equal(t, "", th.FmtSavingTemperature())
	assert.Equal(t, "", th.FmtGoalTemperature())
	assert.Equal(t, "", th.FmtComfortTemperature())
	assert.Equal(t, -1, th.State())
}

// TestFormattingOfTemperaturesSpecialValueOff tests formatting of the temperature values obtained by AHA interface.
func TestFormattingOfTemperaturesSpecialValueOff(t *testing.T) {
	th := Thermostat{Measured: "253", Saving: "253", Goal: "253", Comfort: "253"}
	assert.Equal(t, "OFF", th.FmtMeasuredTemperature())
	assert.Equal(t, "OFF", th.FmtSavingTemperature())
	assert.Equal(t, "OFF", th.FmtGoalTemperature())
	assert.Equal(t, "OFF", th.FmtComfortTemperature())
	assert.Equal(t, 0, th.State())
}

// TestFormattingOfTemperaturesSpecialValueOn tests formatting of the temperature values obtained by AHA interface.
func TestFormattingOfTemperaturesSpecialValueOn(t *testing.T) {
	th := Thermostat{Measured: "254", Saving: "254", Goal: "254", Comfort: "254"}
	assert.Equal(t, "ON", th.FmtMeasuredTemperature())
	assert.Equal(t, "ON", th.FmtSavingTemperature())
	assert.Equal(t, "ON", th.FmtGoalTemperature())
	assert.Equal(t, "ON", th.FmtComfortTemperature())
	assert.Equal(t, 1, th.State())
}

// TestFormattingOfTemperaturesOutOfRangeHigh tests formatting of the temperature values obtained by AHA interface.
func TestFormattingOfTemperaturesOutOfRangeHigh(t *testing.T) {
	th := Thermostat{Measured: "112", Saving: "110", Goal: "111", Comfort: "56"}
	assert.Equal(t, "50", th.FmtMeasuredTemperature())
	assert.Equal(t, "28", th.FmtSavingTemperature())
	assert.Equal(t, "28", th.FmtGoalTemperature())
	assert.Equal(t, "28", th.FmtComfortTemperature())
	assert.Equal(t, 1, th.State())
}

// TestFormattingOfTemperaturesOutOfRangeLow tests formatting of the temperature values obtained by AHA interface.
func TestFormattingOfTemperaturesOutOfRangeLow(t *testing.T) {
	th := Thermostat{Measured: "0", Saving: "2", Goal: "3", Comfort: "16"}
	assert.Equal(t, "0", th.FmtMeasuredTemperature())
	assert.Equal(t, "8", th.FmtSavingTemperature())
	assert.Equal(t, "8", th.FmtGoalTemperature())
	assert.Equal(t, "8", th.FmtComfortTemperature())
	assert.Equal(t, 1, th.State())
}

// TestState tests if a proper state is returned from all values that might cary special valued
func TestState(t *testing.T) {
	tests := []struct {
		th   Thermostat
		want int
	}{
		{th: Thermostat{Measured: "0", Saving: "0", Goal: "0", Comfort: "0"}, want: 1},
		{th: Thermostat{Measured: "255", Saving: "255", Goal: "255", Comfort: "255"}, want: -1},
		{th: Thermostat{Measured: "254", Saving: "254", Goal: "254", Comfort: "254"}, want: 1},
		{th: Thermostat{Measured: "253", Saving: "253", Goal: "253", Comfort: "253"}, want: 0},
		// Tests for special ON value (254)
		{th: Thermostat{Measured: "254", Saving: "0", Goal: "0", Comfort: "0"}, want: 1},
		{th: Thermostat{Measured: "0", Saving: "254", Goal: "0", Comfort: "0"}, want: 1},
		{th: Thermostat{Measured: "0", Saving: "0", Goal: "254", Comfort: "0"}, want: 1},
		{th: Thermostat{Measured: "0", Saving: "0", Goal: "0", Comfort: "254"}, want: 1},
		// Tests for undefined values
		{th: Thermostat{Measured: "255", Saving: "0", Goal: "0", Comfort: "0"}, want: -1},
		{th: Thermostat{Measured: "0", Saving: "355", Goal: "0", Comfort: "0"}, want: -1},
		{th: Thermostat{Measured: "0", Saving: "0", Goal: "355", Comfort: "0"}, want: -1},
		{th: Thermostat{Measured: "0", Saving: "0", Goal: "0", Comfort: "255"}, want: -1},
		// Tests for at least one valid value (Thermostat considered ON)
		{th: Thermostat{Measured: "no-float", Saving: "0", Goal: "0", Comfort: "0"}, want: 1},
		{th: Thermostat{Measured: "0", Saving: "no-float", Goal: "0", Comfort: "0"}, want: 1},
		{th: Thermostat{Measured: "0", Saving: "0", Goal: "no-float", Comfort: "0"}, want: 1},
		{th: Thermostat{Measured: "0", Saving: "0", Goal: "0", Comfort: "no-float"}, want: 1},
		// Tests for at least one special OFF value (Thermostat considered OFF
		{th: Thermostat{Measured: "253", Saving: "0", Goal: "0", Comfort: "0"}, want: 0},
		{th: Thermostat{Measured: "0", Saving: "253", Goal: "0", Comfort: "0"}, want: 0},
		{th: Thermostat{Measured: "0", Saving: "0", Goal: "253", Comfort: "0"}, want: 0},
		{th: Thermostat{Measured: "0", Saving: "0", Goal: "0", Comfort: "253"}, want: 0},
	}
	for i, tt := range tests {
		t.Run(fmt.Sprintf("TestState-%d", i), func(t *testing.T) {
			if got := tt.th.State(); got != tt.want {
				t.Errorf("Thermostat.State() = %v, want %v", got, tt.want)
			}
		})

	}
}
