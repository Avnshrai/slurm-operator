// SPDX-FileCopyrightText: Copyright (C) SchedMD LLC.
// SPDX-License-Identifier: Apache-2.0

package utils

import (
	"testing"
	"time"
)

func Test_validFirstDigit(t *testing.T) {
	type args struct {
		str string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "Zero length string",
			args: args{
				str: "",
			},
			want: false,
		},
		{
			name: "Starts with '-'",
			args: args{
				str: "-foo",
			},
			want: true,
		},
		{
			name: "Starts with '0'",
			args: args{
				str: "0foo",
			},
			want: false,
		},
		{
			name: "Is '0'",
			args: args{
				str: "0",
			},
			want: true,
		},
		{
			name: "Starts with '1'",
			args: args{
				str: "1foo",
			},
			want: true,
		},
		{
			name: "Starts with '9'",
			args: args{
				str: "9foo",
			},
			want: true,
		},
		{
			name: "Starts with '5'",
			args: args{
				str: "5foo",
			},
			want: true,
		},
		{
			name: "Starts with 'a'",
			args: args{
				str: "afoo",
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := validFirstDigit(tt.args.str); got != tt.want {
				t.Errorf("validFirstDigit() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetNumberFromAnnotations(t *testing.T) {
	type args struct {
		annotations map[string]string
		key         string
	}
	tests := []struct {
		name    string
		args    args
		want    int32
		wantErr bool
	}{
		{
			name: "Get number from map key: 1",
			args: args{
				annotations: map[string]string{"foo": "1"},
				key:         "foo",
			},
			want:    1,
			wantErr: false,
		},
		{
			name: "Get number from map key: missing key",
			args: args{
				annotations: map[string]string{"bar": "1"},
				key:         "foo",
			},
			want:    0,
			wantErr: false,
		},
		{
			name: "Get number from map key: parse error",
			args: args{
				annotations: map[string]string{"foo": "1_2"},
				key:         "foo",
			},
			want:    0,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetNumberFromAnnotations(tt.args.annotations, tt.args.key)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetNumberFromAnnotations() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("GetNumberFromAnnotations() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetBoolFromAnnotations(t *testing.T) {
	type args struct {
		annotations map[string]string
		key         string
	}
	tests := []struct {
		name    string
		args    args
		want    bool
		wantErr bool
	}{
		{
			name: "Get number from map key: True",
			args: args{
				annotations: map[string]string{"foo": "True"},
				key:         "foo",
			},
			want:    true,
			wantErr: false,
		},
		{
			name: "Get number from map key: False",
			args: args{
				annotations: map[string]string{"foo": "False"},
				key:         "foo",
			},
			want:    false,
			wantErr: false,
		},
		{
			name: "Get number from map key: 1",
			args: args{
				annotations: map[string]string{"foo": "1"},
				key:         "foo",
			},
			want:    true,
			wantErr: false,
		},
		{
			name: "Get number from map key: 0",
			args: args{
				annotations: map[string]string{"foo": "0"},
				key:         "foo",
			},
			want:    false,
			wantErr: false,
		},
		{
			name: "Get number from map key: missing key",
			args: args{
				annotations: map[string]string{"bar": "true"},
				key:         "foo",
			},
			want:    false,
			wantErr: false,
		},
		{
			name: "Get number from map key: parse error",
			args: args{
				annotations: map[string]string{"foo": " "},
				key:         "foo",
			},
			want:    false,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetBoolFromAnnotations(tt.args.annotations, tt.args.key)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetBoolFromAnnotations() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("GetBoolFromAnnotations() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetTimeFromAnnotations(t *testing.T) {
	now := time.Now()
	unitTime := time.Time{}
	type args struct {
		annotations map[string]string
		key         string
	}
	tests := []struct {
		name    string
		args    args
		want    time.Time
		wantErr bool
	}{
		{
			name: "Get number from map key: Now",
			args: args{
				annotations: map[string]string{"foo": now.Format(time.RFC3339)},
				key:         "foo",
			},
			want:    now,
			wantErr: false,
		},
		{
			name: "Get number from map key: parse error",
			args: args{
				annotations: map[string]string{"foo": " "},
				key:         "foo",
			},
			want:    unitTime,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetTimeFromAnnotations(tt.args.annotations, tt.args.key)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetTimeFromAnnotations() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			// Compare Unix to avoid nanosecond precision loss due to (un)marshaling
			if tt.want.Unix() != got.Unix() {
				t.Errorf("GetTimeFromAnnotations() = %v, want %v", got, tt.want)
			}
		})
	}
}
