package rnnoise

import (
	"fmt"

	"github.com/sirupsen/logrus"
)

// NoiseFilter 噪声过滤器，整合RNNoise和音频处理功能
//
// NoiseFilter 提供了完整的音频降噪解决方案，包括：
// - 音频格式转换（支持多种采样率和声道配置）
// - 批量音频文件处理
// - 原始PCM数据处理
// - 流式音频处理
// - 语音概率分析和统计
type NoiseFilter struct {
	rnnoise   *RNNoise
	processor *AudioProcessor
}

// NewNoiseFilter 创建新的噪声过滤器
//
// 参数:
//   - libPath: RNNoise动态库文件路径，如果为空则自动查找
//
// 返回:
//   - *NoiseFilter: 噪声过滤器实例
//   - error: 创建失败时的错误信息
//
// 示例:
//
//	filter, err := NewNoiseFilter("")
//	if err != nil {
//	    log.Fatal(err)
//	}
//	defer filter.Destroy()
func NewNoiseFilter(libPath string) (*NoiseFilter, error) {
	rnnoise, err := NewRNNoise(libPath)
	if err != nil {
		return nil, err
	}

	processor := NewAudioProcessor(rnnoise)

	return &NoiseFilter{
		rnnoise:   rnnoise,
		processor: processor,
	}, nil
}

// Destroy 销毁噪声过滤器，释放资源
func (nf *NoiseFilter) Destroy() {
	if nf.rnnoise != nil {
		nf.rnnoise.Destroy()
	}
}

// Reset 重置RNNoise状态
func (nf *NoiseFilter) Reset() error {
	return nf.rnnoise.Reset()
}

// GetRNNoise 获取RNNoise实例（用于创建AudioProcessor）
func (nf *NoiseFilter) GetRNNoise() *RNNoise {
	return nf.rnnoise
}

// FilterResult 过滤结果
type FilterResult struct {
	DenoisedAudio      *AudioData // 降噪后的音频
	VoiceProbabilities []float32  // 每帧的语音概率
	ProcessedFrames    int        // 处理的帧数
}

// FilterAudio 对音频进行降噪处理
func (nf *NoiseFilter) FilterAudio(audioData *AudioData, voiceProbThreshold float32) (*FilterResult, error) {
	// 1. 转换音频格式为RNNoise支持的格式
	convertedAudio, err := nf.processor.ConvertToRNNoiseFormat(audioData)

	if err != nil {
		return nil, fmt.Errorf("音频格式转换失败: %v", err)
	}

	// 2. 分割为帧
	frames, err := nf.processor.GetFrames(convertedAudio)
	if err != nil {
		return nil, fmt.Errorf("音频分帧失败: %v", err)
	}
	logrus.Debugf("分帧结果: %d帧, 每帧%d样本", len(frames), func() int {
		if len(frames) > 0 {
			return len(frames[0])
		}
		return 0
	}())

	// 3. 处理每个帧
	var denoisedFrames [][]float32
	var voiceProbabilities []float32
	for _, frame := range frames {
		voiceProb, denoisedFrame, err := nf.rnnoise.ProcessFrame(frame)
		logrus.Debugf("RNNoise当前帧概率为:%v", voiceProb)
		if err != nil {
			return nil, fmt.Errorf("帧处理失败: %v", err)
		}

		voiceProbabilities = append(voiceProbabilities, voiceProb)

		// 根据语音概率阈值决定是否保留帧
		if voiceProb >= voiceProbThreshold {
			denoisedFrames = append(denoisedFrames, denoisedFrame)
		}
	}
	// 4. 重新组合音频
	var allSamples []float32
	for _, frame := range denoisedFrames {
		allSamples = append(allSamples, frame...)
	}

	denoisedAudio := &AudioData{
		Samples:    allSamples,
		SampleRate: 48000,
		Channels:   1,
		BitDepth:   16,
	}

	// 5. 如果需要，转换回原始采样率
	if audioData.SampleRate != 48000 { //肯定不是48000，因为之前是8000。所以需要转换回原始采样率
		logrus.Debugf("RNN-FilterAudio转换回原始采样率: %d", audioData.SampleRate)
		denoisedAudio, err = nf.convertSampleRate(denoisedAudio, audioData.SampleRate)
		if err != nil {
			return nil, fmt.Errorf("采样率转换失败: %v", err)
		}
	}

	return &FilterResult{
		DenoisedAudio:      denoisedAudio,
		VoiceProbabilities: voiceProbabilities,
		ProcessedFrames:    len(frames),
	}, nil
}

// FilterAudioFile 直接处理音频文件
func (nf *NoiseFilter) FilterAudioFile(inputFile, outputFile string, voiceProbThreshold float32) (*FilterResult, error) {

	// 读取输入文件
	audioData, err := nf.processor.ReadWAV(inputFile)
	if err != nil {
		return nil, fmt.Errorf("读取音频文件失败: %v", err)
	}

	// 进行降噪处理
	result, err := nf.FilterAudio(audioData, voiceProbThreshold)
	if err != nil {
		return nil, err
	}

	// 写入输出文件
	err = nf.processor.WriteWAV(outputFile, result.DenoisedAudio)
	if err != nil {
		return nil, fmt.Errorf("写入音频文件失败: %v", err)
	}

	return result, nil
}

// FilterAudioBytes 处理音频字节数据，对原始PCM音频进行RNNoise降噪处理
//
// 这是处理原始PCM音频数据的主要方法，支持多种音频格式的自动转换和处理。
// 内部会将音频转换为48kHz单声道进行处理，然后转换回原始格式。
//
// 参数:
//   - audioBytes: 原始PCM音频字节数据（小端序格式）
//   - sampleRate: 音频采样率（Hz），支持8000, 16000, 44100, 48000等
//   - channels: 声道数（1=单声道, 2=立体声）
//   - bitDepth: 位深度（支持16, 24, 32位）
//   - voiceProbThreshold: 语音概率阈值（0.0-1.0），低于此值的帧会被过滤
//
// 返回:
//   - []byte: 降噪后的PCM音频字节数据（与输入格式相同）
//   - []float32: 每帧的语音概率数组（0.0-1.0，每个值对应10ms音频帧）
//   - error: 处理过程中的错误信息，成功时为nil
//
// 示例:
//
//	// 处理8000Hz单声道16位PCM数据
//	denoisedBytes, voiceProbs, err := filter.FilterAudioBytes(
//	    pcmData, 8000, 1, 16, 0.3)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Printf("检测到 %d 帧语音\n", len(voiceProbs))
func (nf *NoiseFilter) FilterAudioBytes(audioBytes []byte, sampleRate, channels, bitDepth int, voiceProbThreshold float32) ([]byte, []float32, error) {
	// 转换字节为float32样本
	//输入320字节，16位深，8000hz，单声道
	//输出为160长度的float32数组,48000hz，单声道，16位深
	samples := ConvertBytesToFloat32(audioBytes, bitDepth)

	audioData := &AudioData{
		Samples:    samples,
		SampleRate: sampleRate,
		Channels:   channels,
		BitDepth:   bitDepth,
	}

	// 进行降噪处理
	result, err := nf.FilterAudio(audioData, voiceProbThreshold)
	if err != nil {
		return nil, nil, err
	}

	// 转换回字节格式
	denoisedBytes := ConvertFloat32ToBytes(result.DenoisedAudio.Samples, bitDepth)

	return denoisedBytes, result.VoiceProbabilities, nil
}

// FilterStream 流式处理音频（每次处理一个10ms的帧）
func (nf *NoiseFilter) FilterStream(frame []float32, voiceProbThreshold float32) ([]float32, float32, bool, error) {
	if len(frame) != 480 {
		return nil, 0, false, fmt.Errorf("流式处理要求帧大小为480个样本（10ms @ 48kHz），当前为%d", len(frame))
	}

	voiceProb, denoisedFrame, err := nf.rnnoise.ProcessFrame(frame)
	if err != nil {
		return nil, 0, false, err
	}

	// 判断是否保留该帧
	keepFrame := voiceProb >= voiceProbThreshold

	return denoisedFrame, voiceProb, keepFrame, nil
}

// convertSampleRate 简单的采样率转换（线性插值）
func (nf *NoiseFilter) convertSampleRate(audioData *AudioData, targetSampleRate int) (*AudioData, error) {
	if audioData.SampleRate == targetSampleRate {
		return audioData, nil
	}

	ratio := float64(targetSampleRate) / float64(audioData.SampleRate)
	newLength := int(float64(len(audioData.Samples)) * ratio)
	resampledSamples := make([]float32, newLength)

	for i := 0; i < newLength; i++ {
		srcIndex := float64(i) / ratio
		srcIndexInt := int(srcIndex)

		if srcIndexInt >= len(audioData.Samples)-1 {
			resampledSamples[i] = audioData.Samples[len(audioData.Samples)-1]
		} else {
			// 线性插值
			frac := float32(srcIndex - float64(srcIndexInt))
			resampledSamples[i] = audioData.Samples[srcIndexInt]*(1-frac) + audioData.Samples[srcIndexInt+1]*frac
		}
	}

	return &AudioData{
		Samples:    resampledSamples,
		SampleRate: targetSampleRate,
		Channels:   audioData.Channels,
		BitDepth:   audioData.BitDepth,
	}, nil
}

// FrameStatistics 获取帧处理统计信息
type FrameStatistics struct {
	TotalFrames      int     // 总帧数
	VoiceFrames      int     // 语音帧数
	NoiseFrames      int     // 噪声帧数
	AverageVoiceProb float32 // 平均语音概率
	MaxVoiceProb     float32 // 最大语音概率
	MinVoiceProb     float32 // 最小语音概率
}

// AnalyzeFrames 分析音频帧的统计信息
func (nf *NoiseFilter) AnalyzeFrames(audioData *AudioData, voiceProbThreshold float32) (*FrameStatistics, error) {
	// 转换音频格式
	convertedAudio, err := nf.processor.ConvertToRNNoiseFormat(audioData)
	if err != nil {
		return nil, err
	}

	// 分割为帧
	frames, err := nf.processor.GetFrames(convertedAudio)
	if err != nil {
		return nil, err
	}

	stats := &FrameStatistics{
		TotalFrames:  len(frames),
		MinVoiceProb: 1.0,
	}

	var totalProb float32
	for _, frame := range frames {
		voiceProb, _, err := nf.rnnoise.ProcessFrame(frame)
		if err != nil {
			return nil, err
		}

		totalProb += voiceProb

		if voiceProb >= voiceProbThreshold {
			stats.VoiceFrames++
		} else {
			stats.NoiseFrames++
		}

		if voiceProb > stats.MaxVoiceProb {
			stats.MaxVoiceProb = voiceProb
		}
		if voiceProb < stats.MinVoiceProb {
			stats.MinVoiceProb = voiceProb
		}
	}

	if stats.TotalFrames > 0 {
		stats.AverageVoiceProb = totalProb / float32(stats.TotalFrames)
	}

	return stats, nil
}
