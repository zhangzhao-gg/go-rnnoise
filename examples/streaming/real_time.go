package main

import (
	"fmt"
	"log"
	"time"

	"github.com/zhangzhao-gg/go-rnnoise/rnnoise"
)

// 实时流式音频处理示例
func main() {
	fmt.Println("实时流式音频处理示例")

	// 创建噪声过滤器
	filter, err := rnnoise.NewNoiseFilter("")
	if err != nil {
		log.Fatal("创建噪声过滤器失败:", err)
	}
	defer filter.Destroy()

	// 模拟实时音频流
	voiceThreshold := float32(0.3)
	frameDuration := 10 * time.Millisecond // 10ms帧

	fmt.Printf("开始实时处理，帧大小: %v\n", frameDuration)
	fmt.Println("模拟处理音频流...")

	// 模拟处理多个音频帧
	for i := 0; i < 20; i++ {
		// 生成模拟的480个样本（10ms @ 48kHz）
		frame := make([]float32, 480)
		for j := range frame {
			// 生成模拟的音频数据（正弦波 + 噪声）
			// 模拟语音和噪声的混合
			voiceSignal := float32(0.3 * float64(i%3))               // 模拟语音信号
			noiseSignal := float32(0.1 * (float64(j%50)/50.0 - 0.5)) // 模拟噪声
			frame[j] = voiceSignal + noiseSignal
		}

		// 处理帧
		denoisedFrame, voiceProb, keepFrame, err := filter.FilterStream(frame, voiceThreshold)
		if err != nil {
			log.Printf("帧 %d 处理失败: %v", i+1, err)
			continue
		}

		// 显示结果
		status := "丢弃"
		if keepFrame {
			status = "保留"
		}

		fmt.Printf("帧 %2d: 语音概率=%.3f, 状态=%s, 输出样本数=%d\n",
			i+1, voiceProb, status, len(denoisedFrame))

		// 模拟实时处理延迟
		time.Sleep(frameDuration)
	}

	fmt.Println("\n实时流式处理演示完成")
	fmt.Println("在实际应用中，您可以将音频流数据逐帧传入FilterStream函数进行实时处理")
}
