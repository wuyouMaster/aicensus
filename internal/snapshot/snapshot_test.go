package snapshot

import "testing"

func TestFormatBytesUsesReadableUnits(t *testing.T) {
	tests := []struct {
		name  string
		bytes int64
		want  string
	}{
		{name: "zero", bytes: 0, want: "0 B"},
		{name: "bytes", bytes: 1023, want: "1023 B"},
		{name: "kilobytes", bytes: 1024, want: "1.00 KB"},
		{name: "megabytes", bytes: 1024 * 1024, want: "1.00 MB"},
		{name: "gigabytes", bytes: 1024 * 1024 * 1024, want: "1.00 GB"},
		{name: "terabytes", bytes: 1 << 40, want: "1.00 TB"},
		{name: "petabytes", bytes: 1 << 50, want: "1.00 PB"},
		{name: "negative delta", bytes: -680246, want: "-664.30 KB"},
		{name: "minimum int64", bytes: -1 << 63, want: "-8.00 EB"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := FormatBytes(tt.bytes); got != tt.want {
				t.Fatalf("FormatBytes(%d) = %q, want %q", tt.bytes, got, tt.want)
			}
		})
	}
}

func TestFormatBytesCompactUsesAdaptiveUnits(t *testing.T) {
	tests := []struct {
		bytes int64
		want  string
	}{
		{bytes: 1023, want: "1023 B"},
		{bytes: 1024, want: "1 KB"},
		{bytes: 1 << 30, want: "1 GB"},
		{bytes: 1 << 40, want: "1 TB"},
	}

	for _, tt := range tests {
		if got := FormatBytesCompact(tt.bytes); got != tt.want {
			t.Fatalf("FormatBytesCompact(%d) = %q, want %q", tt.bytes, got, tt.want)
		}
	}
}
