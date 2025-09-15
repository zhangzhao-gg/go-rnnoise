package rnnoise

import (
	"testing"
)

func TestConvertBytesToFloat32(t *testing.T) {
	tests := []struct {
		name     string
		data     []byte
		bitDepth int
		expected int // expected number of samples
	}{
		{
			name:     "16-bit audio",
			data:     []byte{0x00, 0x01, 0xFF, 0x7F}, // 2 samples
			bitDepth: 16,
			expected: 2,
		},
		{
			name:     "24-bit audio",
			data:     []byte{0x00, 0x01, 0x02, 0x03, 0x04, 0x05}, // 2 samples
			bitDepth: 24,
			expected: 2,
		},
		{
			name:     "32-bit audio",
			data:     []byte{0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07}, // 2 samples
			bitDepth: 32,
			expected: 2,
		},
		{
			name:     "empty data",
			data:     []byte{},
			bitDepth: 16,
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ConvertBytesToFloat32(tt.data, tt.bitDepth)
			if len(result) != tt.expected {
				t.Errorf("ConvertBytesToFloat32() got %d samples, want %d", len(result), tt.expected)
			}
		})
	}
}

func TestConvertFloat32ToBytes(t *testing.T) {
	tests := []struct {
		name     string
		samples  []float32
		bitDepth int
	}{
		{
			name:     "16-bit conversion",
			samples:  []float32{0.0, 1.0, -1.0, 0.5},
			bitDepth: 16,
		},
		{
			name:     "24-bit conversion",
			samples:  []float32{0.0, 1.0, -1.0},
			bitDepth: 24,
		},
		{
			name:     "32-bit conversion",
			samples:  []float32{0.0, 1.0, -1.0},
			bitDepth: 32,
		},
		{
			name:     "empty samples",
			samples:  []float32{},
			bitDepth: 16,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ConvertFloat32ToBytes(tt.samples, tt.bitDepth)

			// 验证返回的数据不为空（除非输入为空）
			if len(tt.samples) > 0 && len(result) == 0 {
				t.Error("ConvertFloat32ToBytes() returned empty result for non-empty input")
			}

			// 验证返回的数据长度合理
			expectedBytes := len(tt.samples) * (tt.bitDepth / 8)
			if len(tt.samples) > 0 && len(result) != expectedBytes {
				t.Errorf("ConvertFloat32ToBytes() got %d bytes, want %d", len(result), expectedBytes)
			}
		})
	}
}

func TestAudioData(t *testing.T) {
	// 测试 AudioData 结构体的基本功能
	audioData := &AudioData{
		Samples:    []float32{0.1, 0.2, 0.3, 0.4, 0.5},
		SampleRate: 48000,
		Channels:   1,
		BitDepth:   16,
	}

	if len(audioData.Samples) != 5 {
		t.Errorf("Expected 5 samples, got %d", len(audioData.Samples))
	}

	if audioData.SampleRate != 48000 {
		t.Errorf("Expected sample rate 48000, got %d", audioData.SampleRate)
	}

	if audioData.Channels != 1 {
		t.Errorf("Expected 1 channel, got %d", audioData.Channels)
	}

	if audioData.BitDepth != 16 {
		t.Errorf("Expected bit depth 16, got %d", audioData.BitDepth)
	}
}
