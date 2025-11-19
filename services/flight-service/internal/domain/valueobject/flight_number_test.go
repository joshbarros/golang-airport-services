package valueobject

import (
	"testing"
)

func TestNewFlightNumber(t *testing.T) {
	tests := []struct {
		name    string
		value   string
		wantErr bool
	}{
		{
			name:    "valid flight number with 3 digits",
			value:   "AA123",
			wantErr: false,
		},
		{
			name:    "valid flight number with 4 digits",
			value:   "DL1234",
			wantErr: false,
		},
		{
			name:    "valid flight number with lowercase",
			value:   "aa123",
			wantErr: false,
		},
		{
			name:    "invalid empty flight number",
			value:   "",
			wantErr: true,
		},
		{
			name:    "invalid flight number too short",
			value:   "A1",
			wantErr: true,
		},
		{
			name:    "invalid flight number without digits",
			value:   "ABCD",
			wantErr: true,
		},
		{
			name:    "invalid flight number with special chars",
			value:   "AA@123",
			wantErr: true,
		},
		{
			name:    "invalid flight number too long",
			value:   "AA123456",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewFlightNumber(tt.value)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewFlightNumber() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got.String() == "" {
				t.Error("NewFlightNumber() returned empty value for valid input")
			}
		})
	}
}

func TestFlightNumber_String(t *testing.T) {
	fn, _ := NewFlightNumber("AA123")
	want := "AA123"

	if got := fn.String(); got != want {
		t.Errorf("String() = %v, want %v", got, want)
	}
}

func TestFlightNumber_Airline(t *testing.T) {
	tests := []struct {
		name         string
		flightNumber string
		want         string
	}{
		{
			name:         "extracts two-letter airline code",
			flightNumber: "AA123",
			want:         "AA",
		},
		{
			name:         "extracts three-letter airline code",
			flightNumber: "DAL1234",
			want:         "DAL",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fn, _ := NewFlightNumber(tt.flightNumber)
			if got := fn.Airline(); got != tt.want {
				t.Errorf("Airline() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFlightNumber_Number(t *testing.T) {
	tests := []struct {
		name         string
		flightNumber string
		want         string
	}{
		{
			name:         "extracts three-digit number",
			flightNumber: "AA123",
			want:         "123",
		},
		{
			name:         "extracts four-digit number",
			flightNumber: "DAL1234",
			want:         "1234",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fn, _ := NewFlightNumber(tt.flightNumber)
			if got := fn.Number(); got != tt.want {
				t.Errorf("Number() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFlightNumber_Equals(t *testing.T) {
	fn1, _ := NewFlightNumber("AA123")
	fn2, _ := NewFlightNumber("AA123")
	fn3, _ := NewFlightNumber("DL456")

	if !fn1.Equals(fn2) {
		t.Error("Equal flight numbers should be equal")
	}

	if fn1.Equals(fn3) {
		t.Error("Different flight numbers should not be equal")
	}
}
