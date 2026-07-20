package main

import "testing"

func Test_validateHostPort(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		// Named input parameters for target function.
		address string
		wantErr bool
	}{
		{
			name:    "valid address",
			address: "localhost:8080",
			wantErr: false,
		},
		{
			name:    "invalid address",
			address: "localhost:abc",
			wantErr: true,
		},
		// Address with schema should fail
		{
			name:    "invalid address with schema",
			address: "http://localhost:8080",
			wantErr: true,
		},
		{
			name:    "invalid address with empty port",
			address: "localhost:",
			wantErr: true,
		},
		{
			name:    "invalid address with empty host",
			address: ":8080",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotErr := validateHostPort(tt.address)
			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("validateHostPort() failed: %v", gotErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("validateHostPort() succeeded unexpectedly")
			}
		})
	}
}
