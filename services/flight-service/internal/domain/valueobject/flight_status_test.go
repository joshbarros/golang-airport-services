package valueobject

import (
	"testing"
)

func TestNewFlightStatus(t *testing.T) {
	tests := []struct {
		name    string
		value   string
		wantErr bool
	}{
		{
			name:    "valid scheduled status",
			value:   "scheduled",
			wantErr: false,
		},
		{
			name:    "valid boarding status",
			value:   "boarding",
			wantErr: false,
		},
		{
			name:    "valid departed status",
			value:   "departed",
			wantErr: false,
		},
		{
			name:    "valid in_flight status",
			value:   "in_flight",
			wantErr: false,
		},
		{
			name:    "valid landed status",
			value:   "landed",
			wantErr: false,
		},
		{
			name:    "valid arrived status",
			value:   "arrived",
			wantErr: false,
		},
		{
			name:    "valid delayed status",
			value:   "delayed",
			wantErr: false,
		},
		{
			name:    "valid cancelled status",
			value:   "cancelled",
			wantErr: false,
		},
		{
			name:    "valid with uppercase",
			value:   "SCHEDULED",
			wantErr: false,
		},
		{
			name:    "invalid empty status",
			value:   "",
			wantErr: true,
		},
		{
			name:    "invalid unknown status",
			value:   "unknown",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewFlightStatus(tt.value)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewFlightStatus() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got.String() == "" {
				t.Error("NewFlightStatus() returned empty value for valid input")
			}
		})
	}
}

func TestFlightStatus_String(t *testing.T) {
	status, _ := NewFlightStatus("scheduled")
	want := "scheduled"

	if got := status.String(); got != want {
		t.Errorf("String() = %v, want %v", got, want)
	}
}

func TestFlightStatus_Equals(t *testing.T) {
	status1, _ := NewFlightStatus("scheduled")
	status2, _ := NewFlightStatus("scheduled")
	status3, _ := NewFlightStatus("cancelled")

	if !status1.Equals(status2) {
		t.Error("Equal statuses should be equal")
	}

	if status1.Equals(status3) {
		t.Error("Different statuses should not be equal")
	}
}

func TestFlightStatus_IsFinal(t *testing.T) {
	tests := []struct {
		name   string
		status string
		want   bool
	}{
		{
			name:   "arrived is final",
			status: "arrived",
			want:   true,
		},
		{
			name:   "cancelled is final",
			status: "cancelled",
			want:   true,
		},
		{
			name:   "scheduled is not final",
			status: "scheduled",
			want:   false,
		},
		{
			name:   "boarding is not final",
			status: "boarding",
			want:   false,
		},
		{
			name:   "in_flight is not final",
			status: "in_flight",
			want:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			status, _ := NewFlightStatus(tt.status)
			if got := status.IsFinal(); got != tt.want {
				t.Errorf("IsFinal() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFlightStatus_CanTransitionTo(t *testing.T) {
	tests := []struct {
		name       string
		from       string
		to         string
		wantCanDo  bool
	}{
		{
			name:      "scheduled can transition to boarding",
			from:      "scheduled",
			to:        "boarding",
			wantCanDo: true,
		},
		{
			name:      "scheduled can transition to delayed",
			from:      "scheduled",
			to:        "delayed",
			wantCanDo: true,
		},
		{
			name:      "scheduled can transition to cancelled",
			from:      "scheduled",
			to:        "cancelled",
			wantCanDo: true,
		},
		{
			name:      "boarding can transition to departed",
			from:      "boarding",
			to:        "departed",
			wantCanDo: true,
		},
		{
			name:      "departed can transition to in_flight",
			from:      "departed",
			to:        "in_flight",
			wantCanDo: true,
		},
		{
			name:      "in_flight can transition to landed",
			from:      "in_flight",
			to:        "landed",
			wantCanDo: true,
		},
		{
			name:      "landed can transition to arrived",
			from:      "landed",
			to:        "arrived",
			wantCanDo: true,
		},
		{
			name:      "arrived cannot transition to scheduled",
			from:      "arrived",
			to:        "scheduled",
			wantCanDo: false,
		},
		{
			name:      "cancelled cannot transition to boarding",
			from:      "cancelled",
			to:        "boarding",
			wantCanDo: false,
		},
		{
			name:      "boarding cannot transition to arrived",
			from:      "boarding",
			to:        "arrived",
			wantCanDo: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			from, _ := NewFlightStatus(tt.from)
			to, _ := NewFlightStatus(tt.to)

			if got := from.CanTransitionTo(to); got != tt.wantCanDo {
				t.Errorf("CanTransitionTo() = %v, want %v", got, tt.wantCanDo)
			}
		})
	}
}

func TestFlightStatus_IsActive(t *testing.T) {
	tests := []struct {
		name   string
		status string
		want   bool
	}{
		{
			name:   "scheduled is active",
			status: "scheduled",
			want:   true,
		},
		{
			name:   "boarding is active",
			status: "boarding",
			want:   true,
		},
		{
			name:   "departed is active",
			status: "departed",
			want:   true,
		},
		{
			name:   "in_flight is active",
			status: "in_flight",
			want:   true,
		},
		{
			name:   "landed is active",
			status: "landed",
			want:   true,
		},
		{
			name:   "delayed is active",
			status: "delayed",
			want:   true,
		},
		{
			name:   "arrived is not active",
			status: "arrived",
			want:   false,
		},
		{
			name:   "cancelled is not active",
			status: "cancelled",
			want:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			status, _ := NewFlightStatus(tt.status)
			if got := status.IsActive(); got != tt.want {
				t.Errorf("IsActive() = %v, want %v", got, tt.want)
			}
		})
	}
}
