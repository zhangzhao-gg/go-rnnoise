package rnnoise

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"os"

	"github.com/go-audio/audio"
	"github.com/go-audio/wav"
	"github.com/sirupsen/logrus"
)

// AudioProcessor 音频处理器，用于处理WAV文件和音频数据
//
// AudioProcessor 提供了音频文件的读写、格式转换和帧处理功能。
// 它支持多种音频格式的自动转换，将各种格式统一转换为RNNoise支持的48kHz单声道格式。
type AudioProcessor struct {
	rnnoise *RNNoise
}

// NewAudioProcessor 创建新的音频处理器
func NewAudioProcessor(rnnoise *RNNoise) *AudioProcessor {
	return &AudioProcessor{
		rnnoise: rnnoise,
	}
}

// AudioData 音频数据结构
//
// AudioData 封装了音频的基本信息，包括样本数据、采样率、声道数和位深度。
// 这是整个音频处理流程中的核心数据结构。
type AudioData struct {
	Samples    []float32 // 音频样本数据
	SampleRate int       // 采样率
	Channels   int       // 声道数
	BitDepth   int       // 位深度
}

// ReadWAV 读取WAV文件
func (ap *AudioProcessor) ReadWAV(filename string) (*AudioData, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("无法打开文件 %s: %v", filename, err)
	}
	defer file.Close()

	decoder := wav.NewDecoder(file)
	if !decoder.IsValidFile() {
		return nil, fmt.Errorf("无效的WAV文件: %s", filename)
	}
	// 获取音频格式信息
	format := decoder.Format()
	// 读取所有音频数据
	buf := &audio.IntBuffer{
		Data:   make([]int, 0),
		Format: format,
	}
	// 分块读取音频数据
	for {
		chunk := &audio.IntBuffer{
			Data:   make([]int, 1024),
			Format: format,
		}

		n, err := decoder.PCMBuffer(chunk)
		if err != nil {
			if err == io.EOF {
				break
			}
			return nil, fmt.Errorf("读取音频数据失败: %v", err)
		}

		// 如果读取到0字节且没有错误，说明已经读完了
		if n == 0 {
			break
		}

		buf.Data = append(buf.Data, chunk.Data[:n]...)
	}

	// 转换为float32格式
	samples := make([]float32, len(buf.Data))
	bitDepth := 16 // 默认16位，因为go-audio没有直接的SampleSize字段
	maxVal := float32(int32(1) << uint(bitDepth-1))
	for i, sample := range buf.Data {
		samples[i] = float32(sample) / maxVal
	}

	return &AudioData{
		Samples:    samples,
		SampleRate: int(format.SampleRate),
		Channels:   int(format.NumChannels),
		BitDepth:   bitDepth,
	}, nil
}

// WriteWAV 写入WAV文件
func (ap *AudioProcessor) WriteWAV(filename string, audioData *AudioData) error {
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("无法创建文件 %s: %v", filename, err)
	}
	defer file.Close()

	// 设置音频格式
	format := &audio.Format{
		NumChannels: audioData.Channels,
		SampleRate:  audioData.SampleRate,
	}

	encoder := wav.NewEncoder(file, int(format.SampleRate), audioData.BitDepth, int(format.NumChannels), 1)
	defer encoder.Close()

	// 转换float32样本为int格式
	maxVal := float32(int32(1) << uint(audioData.BitDepth-1))
	intSamples := make([]int, len(audioData.Samples))
	for i, sample := range audioData.Samples {
		// 限制范围并转换
		if sample > 1.0 {
			sample = 1.0
		} else if sample < -1.0 {
			sample = -1.0
		}
		intSamples[i] = int(sample * maxVal)
	}

	// 创建音频缓冲区
	buf := &audio.IntBuffer{
		Data:   intSamples,
		Format: format,
	}

	// 写入音频数据
	if err := encoder.Write(buf); err != nil {
		return fmt.Errorf("写入音频数据失败: %v", err)
	}

	return nil
}

// ConvertToRNNoiseFormat 将音频转换为RNNoise支持的格式（48kHz, 单声道, 16位）
// 输入音频格式: 8000Hz, 1声道, 16位深, 160样本
// 转换后音频格式: 48000Hz, 1声道, 16位深, 960样本
func (ap *AudioProcessor) ConvertToRNNoiseFormat(audioData *AudioData) (*AudioData, error) {
	logrus.Debugf("输入音频格式: %dHz, %d声道, %d位深, %d样本",
		audioData.SampleRate, audioData.Channels, audioData.BitDepth, len(audioData.Samples))

	result := &AudioData{
		SampleRate: 48000,
		Channels:   1,
		BitDepth:   16,
	}

	samples := audioData.Samples

	// 1. 转换声道（如果是立体声，取平均值）
	if audioData.Channels == 2 {
		monoSamples := make([]float32, len(samples)/2)
		for i := 0; i < len(monoSamples); i++ {
			monoSamples[i] = (samples[i*2] + samples[i*2+1]) / 2.0
		}
		samples = monoSamples
	} else if audioData.Channels > 2 {
		// 多声道转单声道（取平均值）
		monoSamples := make([]float32, len(samples)/audioData.Channels)
		for i := 0; i < len(monoSamples); i++ {
			var sum float32
			for ch := 0; ch < audioData.Channels; ch++ {
				sum += samples[i*audioData.Channels+ch]
			}
			monoSamples[i] = sum / float32(audioData.Channels)
		}
		samples = monoSamples
	}

	// 2. 重采样到48kHz（简单的线性插值）
	if audioData.SampleRate != 48000 {
		ratio := float64(48000) / float64(audioData.SampleRate)
		newLength := int(float64(len(samples)) * ratio)
		resampledSamples := make([]float32, newLength)

		for i := 0; i < newLength; i++ {
			srcIndex := float64(i) / ratio
			srcIndexInt := int(srcIndex)

			if srcIndexInt >= len(samples)-1 {
				resampledSamples[i] = samples[len(samples)-1]
			} else {
				// 线性插值
				frac := float32(srcIndex - float64(srcIndexInt))
				resampledSamples[i] = samples[srcIndexInt]*(1-frac) + samples[srcIndexInt+1]*frac
			}
		}
		samples = resampledSamples
	}

	result.Samples = samples
	logrus.Debugf("转换后音频格式: %dHz, %d声道, %d位深, %d样本",
		result.SampleRate, result.Channels, result.BitDepth, len(result.Samples))
	return result, nil
}

// GetFrames 将音频数据分割为10毫秒的帧
func (ap *AudioProcessor) GetFrames(audioData *AudioData) ([][]float32, error) {
	if audioData.SampleRate != 48000 {
		return nil, fmt.Errorf("音频采样率必须为48kHz，当前为%dHz", audioData.SampleRate)
	}

	frameSize := int(48000 * 0.01) // 10ms at 48kHz = 480 samples
	samples := audioData.Samples

	// 如果样本数不是帧大小的整数倍，用零填充
	if len(samples)%frameSize != 0 {
		padding := frameSize - (len(samples) % frameSize)
		samples = append(samples, make([]float32, padding)...)
	}

	// 分割为帧
	numFrames := len(samples) / frameSize
	frames := make([][]float32, numFrames)

	for i := 0; i < numFrames; i++ {
		start := i * frameSize
		end := start + frameSize
		frames[i] = make([]float32, frameSize)
		copy(frames[i], samples[start:end])
	}

	return frames, nil
}

// ConvertBytesToFloat32 将字节数组转换为float32样本
func ConvertBytesToFloat32(data []byte, bitDepth int) []float32 {
	logrus.Debugf("ConvertBytesToFloat32: 输入%d字节, 位深%d位", len(data), bitDepth)
	var samples []float32

	switch bitDepth {
	case 16:
		samples = make([]float32, len(data)/2)
		for i := 0; i < len(samples); i++ {
			sample := int16(binary.LittleEndian.Uint16(data[i*2 : i*2+2]))
			samples[i] = float32(sample) / 32768.0
		}
	case 24:
		samples = make([]float32, len(data)/3)
		for i := 0; i < len(samples); i++ {
			// 24位采样点，需要特殊处理
			bytes := data[i*3 : i*3+3]
			var sample int32
			if bytes[2]&0x80 != 0 { // 负数
				sample = int32(-16777216) | int32(bytes[2])<<16 | int32(bytes[1])<<8 | int32(bytes[0])
			} else { // 正数
				sample = int32(bytes[2])<<16 | int32(bytes[1])<<8 | int32(bytes[0])
			}
			samples[i] = float32(sample) / 8388608.0 // 2^23
		}
	case 32:
		samples = make([]float32, len(data)/4)
		for i := 0; i < len(samples); i++ {
			sample := int32(binary.LittleEndian.Uint32(data[i*4 : i*4+4]))
			samples[i] = float32(sample) / 2147483648.0
		}
	default:
		logrus.Errorf("不支持的位深度: %d", bitDepth)
		return []float32{}
	}

	logrus.Debugf("ConvertBytesToFloat32: 输出%d个样本", len(samples))
	return samples
}

// ConvertFloat32ToBytes 将float32样本转换为字节数组
func ConvertFloat32ToBytes(samples []float32, bitDepth int) []byte {
	var buf bytes.Buffer

	switch bitDepth {
	case 16:
		for _, sample := range samples {
			// 限制范围
			if sample > 1.0 {
				sample = 1.0
			} else if sample < -1.0 {
				sample = -1.0
			}

			intSample := int16(sample * 32767)
			binary.Write(&buf, binary.LittleEndian, intSample)
		}
	case 24:
		for _, sample := range samples {
			// 限制范围
			if sample > 1.0 {
				sample = 1.0
			} else if sample < -1.0 {
				sample = -1.0
			}

			intSample := int32(sample * 8388607) // 2^23 - 1

			// 写入3字节
			buf.WriteByte(byte(intSample & 0xFF))
			buf.WriteByte(byte((intSample >> 8) & 0xFF))
			buf.WriteByte(byte((intSample >> 16) & 0xFF))
		}
	case 32:
		for _, sample := range samples {
			// 限制范围
			if sample > 1.0 {
				sample = 1.0
			} else if sample < -1.0 {
				sample = -1.0
			}

			intSample := int32(sample * 2147483647)
			binary.Write(&buf, binary.LittleEndian, intSample)
		}
	}

	return buf.Bytes()
}
