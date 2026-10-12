package config

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParseGeneralTCPConnectTimeout(t *testing.T) {
	const maxInt64 int64 = 1<<63 - 1

	tests := []struct {
		name    string
		timeout int64
		want    int64
		wantErr bool
	}{
		{
			name:    "default",
			timeout: DefaultRawConfig().TCPConnectTimeout,
			want:    5000,
		},
		{
			name:    "custom",
			timeout: 15000,
			want:    15000,
		},
		{
			name:    "zero",
			timeout: 0,
			wantErr: true,
		},
		{
			name:    "negative",
			timeout: -1,
			wantErr: true,
		},
		{
			name:    "below minimum",
			timeout: 99,
			wantErr: true,
		},
		{
			name:    "minimum",
			timeout: 100,
			want:    100,
		},
		{
			name:    "maximum",
			timeout: 600000,
			want:    600000,
		},
		{
			name:    "above maximum",
			timeout: 600001,
			wantErr: true,
		},
		{
			name:    "duration overflow",
			timeout: maxInt64,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			raw := DefaultRawConfig()
			raw.TCPConnectTimeout = tt.timeout

			general, err := parseGeneral(raw)
			if tt.wantErr {
				require.Error(t, err)
				require.Nil(t, general)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, general)
			require.Equal(t, tt.want, general.TCPConnectTimeout)
		})
	}
}

func TestUnmarshalRawConfigTCPConnectTimeout(t *testing.T) {
	tests := []struct {
		name    string
		yaml    string
		want    int64
		wantErr bool
	}{
		{
			name: "omitted uses default",
			yaml: "mode: rule\n",
			want: 5000,
		},
		{
			name: "custom value",
			yaml: "tcp-connect-timeout: 15000\n",
			want: 15000,
		},
		{
			name:    "seconds instead of milliseconds is rejected",
			yaml:    "tcp-connect-timeout: 15\n",
			wantErr: true,
		},
		{
			name:    "explicit zero is rejected",
			yaml:    "tcp-connect-timeout: 0\n",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			raw, err := UnmarshalRawConfig([]byte(tt.yaml))
			require.NoError(t, err)

			general, err := parseGeneral(raw)
			if tt.wantErr {
				require.Error(t, err)
				require.Nil(t, general)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, general)
			require.Equal(t, tt.want, general.TCPConnectTimeout)
		})
	}
}
