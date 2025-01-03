package services

import "testing"

func TestGetChargeStationId(t *testing.T) {
	pattern := `^([A-Z]{2})\*([A-Z0-9]{3})\*E([0-9]+)\*?(.*)$`

	tests := []struct {
		name      string
		evseId    string
		want      string
		expectErr bool
	}{
		{
			name:      "valid evseId",
			evseId:    "DE*GCE*E00188*001",
			want:      "00188",
			expectErr: false,
		},
		{
			name:      "valid evseId",
			evseId:    "DE*GCE*E00188",
			want:      "00188",
			expectErr: false,
		},
		{
			name:      "invalid evseId",
			evseId:    "DEGCEEACC00161571",
			want:      "",
			expectErr: true,
		},
		{
			name:      "invalid evseId",
			evseId:    "abcd",
			want:      "",
			expectErr: true,
		},
		{
			name:      "empty evseId",
			evseId:    "",
			want:      "",
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := NewEvseUIDService(pattern)
			got, err := service.GetChargeStationId(tt.evseId)
			if (err != nil) != tt.expectErr {
				t.Errorf("GetChargeStationId() error = %v, expectErr %v", err, tt.expectErr)
				return
			}
			if got != tt.want {
				t.Errorf("GetChargeStationId() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetOperatorId(t *testing.T) {
	service := &EvseUIDService{pattern: `^([A-Z]{2})\*([A-Z0-9]{3})\*E([0-9]+)\*?(.*)$`}

	tests := []struct {
		evseId      string
		expectedId  string
		expectedErr bool
	}{
		{"DE*ABC*E1234*5678", "ABC", false},
		{"US*XYZ*E9876*5432", "XYZ", false},
		{"INVALID*ID", "", true},
		{"TOO*SHORT", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.evseId, func(t *testing.T) {
			id, err := service.GetOperatorId(tt.evseId)
			if (err != nil) != tt.expectedErr {
				t.Errorf("expected error: %v, got: %v", tt.expectedErr, err)
			}
			if id != tt.expectedId {
				t.Errorf("expected id: %s, got: %s", tt.expectedId, id)
			}
		})
	}
}

func TestGetCountryCode(t *testing.T) {
	service := &EvseUIDService{pattern: `^([A-Z]{2})\*([A-Z0-9]{3})\*E([0-9]+)\*?(.*)$`}

	tests := []struct {
		evseId       string
		expectedCode string
		expectedErr  bool
	}{
		{"DE*ABC*E1234*5678", "DE", false},
		{"US*XYZ*E9876*5432", "US", false},
		{"INVALID*ID", "", true},
		{"TOO*SHORT", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.evseId, func(t *testing.T) {
			code, err := service.GetCountryCode(tt.evseId)
			if (err != nil) != tt.expectedErr {
				t.Errorf("expected error: %v, got: %v", tt.expectedErr, err)
			}
			if code != tt.expectedCode {
				t.Errorf("expected code: %s, got: %s", tt.expectedCode, code)
			}
		})
	}
}
