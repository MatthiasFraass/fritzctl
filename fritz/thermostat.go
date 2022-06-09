package fritz

import "strconv"

// HkrErrorDescriptions has a translation of error code to a warning/error/status description.
var HkrErrorDescriptions = map[string]string{
	"":  "",
	"0": "",
	"1": " Thermostat adjustment not possible. Is the device mounted correctly?",
	"2": " Valve plunger cannot be driven far enough. Possible solutions: Open and close the plunger a couple of times by hand. Check if the battery is too weak.",
	"3": " Valve plunger cannot be moved. Is it blocked?",
	"4": " Preparing installation.",
	"5": " Device in mode 'INSTALLATION'. It can be mounted now.",
	"6": " Device is adjusting to the valve plunger.",
}

// Thermostat models the "HKR" device.
// codebeat:disable[TOO_MANY_IVARS]
type Thermostat struct {
	Measured           string     `xml:"tist"`            // Measured temperature.
	Goal               string     `xml:"tsoll"`           // Desired temperature, user controlled.
	Saving             string     `xml:"absenk"`          // Energy saving temperature.
	Comfort            string     `xml:"komfort"`         // Comfortable temperature.
	NextChange         NextChange `xml:"nextchange"`      // The next scheduled temperature change.
	Lock               string     `xml:"lock"`            // Switch locked (box defined)? 1/0 (empty if not known or if there was an error).
	DeviceLock         string     `xml:"devicelock"`      // Switch locked (device defined)? 1/0 (empty if not known or if there was an error).
	ErrorCode          string     `xml:"errorcode"`       // Error codes: 0 = OK, 1 = ... see https://avm.de/fileadmin/user_upload/Global/Service/Schnittstellen/AHA-HTTP-Interface.pdf.
	BatteryLow         string     `xml:"batterylow"`      // "0" if the battery is OK, "1" if it is running low on capacity. FIXME: With at least FritzOS 7.21 this field is also part f the device element.
	WindowOpen         string     `xml:"windowopenactiv"` // "1" if detected an open window (usually turns off heating), "0" if not.
	BatteryChargeLevel string     `xml:"battery"`         // Battery charge level in percent. FIXME: With at least FritzOS 7.21 this field is also part f the device element.
}

// codebeat:enable[TOO_MANY_IVARS]

// FmtMeasuredTemperature formats the value of t.Measured as obtained on the xml-over http interface to a floating
// point string, units in °C.
// If the value cannot be parsed an empty string is returned.
// If the value if 255, 254 or 253, "?", "ON" or "OFF" is returned, respectively.
// If the value is greater (less) than 50 (0) a cut-off "50" ("0") is returned.
func (t *Thermostat) FmtMeasuredTemperature() string {
	return fmtTemperatureHkr(t.Measured, 0, 100)
}

// FmtGoalTemperature formats the value of t.Goal as obtained on the xml-over http interface to a floating
// point string, units in °C.
// If the value cannot be parsed an empty string is returned.
// If the value if 255, 254 or 253, "?", "ON" or "OFF" is returned, respectively.
// If the value is greater (less) than 56 (16) a cut-off "28" ("8") is returned.
func (t *Thermostat) FmtGoalTemperature() string {
	return fmtTemperatureHkr(t.Goal, 16, 56)
}

// FmtSavingTemperature formats the value of t.Saving as obtained on the xml-over http interface to a floating
// point string, units in °C.
// If the value cannot be parsed an empty string is returned.
// If the value if 255, 254 or 253, "?", "ON" or "OFF" is returned, respectively.
// If the value is greater (less) than 56 (16) a cut-off "28" ("8") is returned.
func (t *Thermostat) FmtSavingTemperature() string {
	return fmtTemperatureHkr(t.Saving, 16, 56)
}

// FmtComfortTemperature formats the value of t.Comfort as obtained on the xml-over http interface to a floating
// point string, units in °C.
// If the value cannot be parsed an empty string is returned.
// If the value if 255, 254 or 253, "?", "ON" or "OFF" is returned, respectively.
// If the value is greater (less) than 56 (16) a cut-off "28" ("8") is returned.
func (t *Thermostat) FmtComfortTemperature() string {
	return fmtTemperatureHkr(t.Comfort, 16, 56)
}

// State returns 1 in case the thermostat is considered ON, 0 if it is considered OFF
// or -1 in case of error or undefined state.
func (t *Thermostat) State() int {
	var err error
	var f float64
	var state int = -1 // Default to error state

	for _, val := range []string{t.Measured, t.Goal, t.Saving, t.Comfort} {
		f, err = strconv.ParseFloat(val, 64)
		if err != nil {
			continue
		}
		if f < 253 || f == 254 {
			// If there was at least one successful read of a temperature,
			// or a temperate is set to the special value for ON,
			// consider the thermostat ON
			state = 1
		} else if f == 253 {
			// Return state OFF in case one of the values is 253
			state = 0
			break
		} else if f >= 255 {
			// Return error state in case one of the values is undefined
			state = -1
			break
		}
	}
	return state
}
