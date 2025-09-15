package main

import (
	"fmt"
	"log"

	"github.com/zhangzhao-gg/go-rnnoise/rnnoise"
)

// 简单的音频降噪示例
func main() {
	// 创建噪声过滤器（自动查找库文件）
	filter, err := rnnoise.NewNoiseFilter("")
	if err != nil {
		log.Fatal("创建噪声过滤器失败:", err)
	}
	defer filter.Destroy()

	// 处理音频文件
	inputFile := "test.wav"
	outputFile := "denoised.wav"
	voiceThreshold := float32(0.3) // 语音概率阈值

	fmt.Printf("开始处理音频文件: %s\n", inputFile)

	result, err := filter.FilterAudioFile(inputFile, outputFile, voiceThreshold)
	if err != nil {
		log.Fatal("音频处理失败:", err)
	}

	fmt.Printf("处理完成！\n")
	fmt.Printf("- 输出文件: %s\n", outputFile)
	fmt.Printf("- 处理帧数: %d\n", result.ProcessedFrames)
	fmt.Printf("- 音频时长: %.2f 秒\n",
		float64(len(result.DenoisedAudio.Samples))/float64(result.DenoisedAudio.SampleRate))

	// 计算语音帧统计
	voiceFrames := 0
	for _, prob := range result.VoiceProbabilities {
		if prob >= voiceThreshold {
			voiceFrames++
		}
	}

	fmt.Printf("- 语音帧: %d/%d (%.1f%%)\n",
		voiceFrames,
		len(result.VoiceProbabilities),
		float32(voiceFrames)/float32(len(result.VoiceProbabilities))*100)
}
