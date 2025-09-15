package main

import (
	"fmt"
	"log"
	"path/filepath"
	"time"

	"github.com/zhangzhao-gg/go-rnnoise/rnnoise"
)

// 批量音频处理示例
func main() {
	fmt.Println("批量音频处理示例")

	// 创建噪声过滤器
	filter, err := rnnoise.NewNoiseFilter("")
	if err != nil {
		log.Fatal("创建噪声过滤器失败:", err)
	}
	defer filter.Destroy()

	// 模拟批量处理
	inputFiles := []string{
		"audio1.wav",
		"audio2.wav",
		"audio3.wav",
	}

	voiceThreshold := float32(0.3)

	fmt.Printf("开始批量处理 %d 个文件...\n", len(inputFiles))

	for i, inputFile := range inputFiles {
		fmt.Printf("\n处理文件 [%d/%d]: %s\n", i+1, len(inputFiles), filepath.Base(inputFile))

		// 生成输出文件名
		outputFile := fmt.Sprintf("denoised_%s", inputFile)

		startTime := time.Now()

		// 处理音频文件
		result, err := filter.FilterAudioFile(inputFile, outputFile, voiceThreshold)
		if err != nil {
			log.Printf("  处理失败: %v", err)
			continue
		}

		elapsed := time.Since(startTime)

		// 显示处理结果
		audioDuration := float64(len(result.DenoisedAudio.Samples)) / float64(result.DenoisedAudio.SampleRate)
		fmt.Printf("  完成: %.2fs (%.1fx 实时)\n", elapsed.Seconds(), audioDuration/elapsed.Seconds())

		// 计算语音帧统计
		voiceFrames := 0
		for _, prob := range result.VoiceProbabilities {
			if prob >= voiceThreshold {
				voiceFrames++
			}
		}

		fmt.Printf("  语音帧: %d/%d (%.1f%%)\n",
			voiceFrames,
			len(result.VoiceProbabilities),
			float32(voiceFrames)/float32(len(result.VoiceProbabilities))*100)
	}

	fmt.Println("\n批量处理完成！")
}
