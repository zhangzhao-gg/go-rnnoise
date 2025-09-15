package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/zhangzhao-gg/go-rnnoise/rnnoise"
)

func main() {
	// 显示使用帮助
	if len(os.Args) < 2 {
		showUsage()
		return
	}

	command := os.Args[1]
	switch command {
	case "denoise":
		if len(os.Args) < 4 {
			fmt.Println("用法: go run . denoise <输入文件> <输出文件> [语音概率阈值]")
			return
		}
		runDenoise(os.Args[2], os.Args[3], os.Args[4:])
	case "analyze":
		if len(os.Args) < 3 {
			fmt.Println("用法: go run . analyze <输入文件> [语音概率阈值]")
			return
		}
		runAnalyze(os.Args[2], os.Args[3:])
	case "stream":
		runStreamExample()
	case "batch":
		if len(os.Args) < 3 {
			fmt.Println("用法: go run . batch <输入目录> [输出目录]")
			return
		}
		runBatch(os.Args[2], os.Args[3:])
	case "test":
		if len(os.Args) < 5 {
			fmt.Println("用法: go run . test <音频文件> <批处理大小> <测试次数>")
			return
		}
		runPerformanceTest(os.Args[2], os.Args[3], os.Args[4])
	default:
		showUsage()
	}
}

func showUsage() {
	fmt.Println("Go RNNoise 音频降噪工具")
	fmt.Println()
	fmt.Println("用法:")
	fmt.Println("  go run . denoise <输入文件> <输出文件> [语音概率阈值]")
	fmt.Println("    - 对单个音频文件进行降噪处理")
	fmt.Println("    - 语音概率阈值: 0.0-1.0，默认0.0（保留所有帧）")
	fmt.Println()
	fmt.Println("  go run . analyze <输入文件> [语音概率阈值]")
	fmt.Println("    - 分析音频文件的语音/噪声统计信息")
	fmt.Println()
	fmt.Println("  go run . stream")
	fmt.Println("    - 演示流式音频处理")
	fmt.Println()
	fmt.Println("  go run . batch <输入目录> [输出目录]")
	fmt.Println("    - 批量处理目录中的所有WAV文件")
	fmt.Println()
	fmt.Println("  go run . test <音频文件> <批处理大小> <测试次数>")
	fmt.Println("    - 性能测试（模拟Python版本的测试逻辑）")
	fmt.Println()
	fmt.Println("示例:")
	fmt.Println("  go run . denoise input.wav output.wav 0.3")
	fmt.Println("  go run . analyze noisy_audio.wav")
	fmt.Println("  go run . batch ./test_audio ./output")
	fmt.Println("  go run . test test.wav 1 10")
}

func runDenoise(inputFile, outputFile string, args []string) {
	// 解析语音概率阈值
	var voiceProbThreshold float32 = 0.0
	if len(args) > 0 {
		if _, err := fmt.Sscanf(args[0], "%f", &voiceProbThreshold); err != nil {
			log.Printf("无效的语音概率阈值，使用默认值0.0: %v", err)
			voiceProbThreshold = 0.0
		}
	}

	fmt.Printf("开始处理文件: %s\n", inputFile)
	fmt.Printf("输出文件: %s\n", outputFile)
	fmt.Printf("语音概率阈值: %.2f\n", voiceProbThreshold)

	// 创建噪声过滤器
	filter, err := rnnoise.NewNoiseFilter("")

	if err != nil {
		log.Fatalf("创建噪声过滤器失败: %v", err)
	}
	defer filter.Destroy()

	// 记录开始时间
	startTime := time.Now()

	// 处理音频文件
	result, err := filter.FilterAudioFile(inputFile, outputFile, voiceProbThreshold)

	if err != nil {
		log.Fatalf("音频处理失败: %v", err)
	}

	// 记录处理时间
	elapsed := time.Since(startTime)

	// 显示结果
	fmt.Printf("\n处理完成！\n")
	fmt.Printf("处理时间: %.2f 秒\n", elapsed.Seconds())
	fmt.Printf("处理帧数: %d\n", result.ProcessedFrames)
	fmt.Printf("音频时长: %.2f 秒\n", float64(len(result.DenoisedAudio.Samples))/float64(result.DenoisedAudio.SampleRate))
	fmt.Printf("处理速度: %.1fx 实时\n", (float64(len(result.DenoisedAudio.Samples))/float64(result.DenoisedAudio.SampleRate))/elapsed.Seconds())

	// 计算语音帧统计
	voiceFrames := 0
	var avgVoiceProb float32
	for _, prob := range result.VoiceProbabilities {
		avgVoiceProb += prob
		if prob >= voiceProbThreshold {
			voiceFrames++
		}
	}
	if len(result.VoiceProbabilities) > 0 {
		avgVoiceProb /= float32(len(result.VoiceProbabilities))
	}

	fmt.Printf("语音帧: %d/%d (%.1f%%)\n", voiceFrames, len(result.VoiceProbabilities),
		float32(voiceFrames)/float32(len(result.VoiceProbabilities))*100)
	fmt.Printf("平均语音概率: %.3f\n", avgVoiceProb)

	// 如果没有语音帧被保留，给出警告
	if voiceFrames == 0 && voiceProbThreshold > 0 {
		fmt.Printf("\n⚠️  警告: 没有帧达到语音概率阈值 %.2f\n", voiceProbThreshold)
		fmt.Printf("💡 建议: 尝试使用更低的阈值（如0.0-0.3）或检查音频内容\n")
		fmt.Printf("📊 当前输出文件可能为空或很短\n")
	}
}

func runAnalyze(inputFile string, args []string) {
	// 解析语音概率阈值
	var voiceProbThreshold float32 = 0.3
	if len(args) > 0 {
		if _, err := fmt.Sscanf(args[0], "%f", &voiceProbThreshold); err != nil {
			log.Printf("无效的语音概率阈值，使用默认值0.3: %v", err)
			voiceProbThreshold = 0.3
		}
	}

	fmt.Printf("分析文件: %s\n", inputFile)
	fmt.Printf("语音概率阈值: %.2f\n", voiceProbThreshold)

	// 创建噪声过滤器
	filter, err := rnnoise.NewNoiseFilter("")
	if err != nil {
		log.Fatalf("创建噪声过滤器失败: %v", err)
	}
	defer filter.Destroy()

	// 读取音频文件
	processor := rnnoise.NewAudioProcessor(filter.GetRNNoise())
	audioData, err := processor.ReadWAV(inputFile)
	if err != nil {
		log.Fatalf("读取音频文件失败: %v", err)
	}

	fmt.Printf("\n音频信息:\n")
	fmt.Printf("  采样率: %d Hz\n", audioData.SampleRate)
	fmt.Printf("  声道数: %d\n", audioData.Channels)
	fmt.Printf("  位深度: %d 位\n", audioData.BitDepth)
	fmt.Printf("  时长: %.2f 秒\n", float64(len(audioData.Samples))/float64(audioData.SampleRate)/float64(audioData.Channels))

	// 分析帧统计信息
	startTime := time.Now()
	stats, err := filter.AnalyzeFrames(audioData, voiceProbThreshold)
	if err != nil {
		log.Fatalf("分析失败: %v", err)
	}
	elapsed := time.Since(startTime)

	fmt.Printf("\n分析结果:\n")
	fmt.Printf("  总帧数: %d\n", stats.TotalFrames)
	fmt.Printf("  语音帧: %d (%.1f%%)\n", stats.VoiceFrames, float32(stats.VoiceFrames)/float32(stats.TotalFrames)*100)
	fmt.Printf("  噪声帧: %d (%.1f%%)\n", stats.NoiseFrames, float32(stats.NoiseFrames)/float32(stats.TotalFrames)*100)
	fmt.Printf("  平均语音概率: %.3f\n", stats.AverageVoiceProb)
	fmt.Printf("  最大语音概率: %.3f\n", stats.MaxVoiceProb)
	fmt.Printf("  最小语音概率: %.3f\n", stats.MinVoiceProb)
	fmt.Printf("  分析时间: %.2f 秒\n", elapsed.Seconds())
}

func runStreamExample() {
	fmt.Println("流式音频处理演示")
	fmt.Println("注意: 这只是一个演示，实际流式处理需要音频输入源")

	// 创建噪声过滤器
	filter, err := rnnoise.NewNoiseFilter("")
	if err != nil {
		log.Fatalf("创建噪声过滤器失败: %v", err)
	}
	defer filter.Destroy()

	// 模拟音频流（生成一些测试数据）
	fmt.Println("\n模拟处理10个音频帧...")
	voiceProbThreshold := float32(0.3)

	for i := 0; i < 10; i++ {
		// 生成模拟的480个样本（10ms @ 48kHz）
		frame := make([]float32, 480)
		for j := range frame {
			// 生成一些模拟的音频数据（正弦波 + 噪声）
			frame[j] = float32(0.5*float64(i%2) + 0.1*(float64(j%100)/100.0-0.5))
		}

		// 处理帧
		denoisedFrame, voiceProb, keepFrame, err := filter.FilterStream(frame, voiceProbThreshold)
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
	}

	fmt.Println("\n流式处理演示完成")
	fmt.Println("在实际应用中，您可以将音频流数据逐帧传入FilterStream函数进行实时处理")
}

func runBatch(inputDir string, args []string) {
	// 确定输出目录
	outputDir := inputDir + "_denoised"
	if len(args) > 0 {
		outputDir = args[0]
	}

	fmt.Printf("批量处理目录: %s\n", inputDir)
	fmt.Printf("输出目录: %s\n", outputDir)

	// 创建输出目录
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		log.Fatalf("创建输出目录失败: %v", err)
	}

	// 查找所有WAV文件
	wavFiles := []string{}
	err := filepath.Walk(inputDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if filepath.Ext(path) == ".wav" && !info.IsDir() {
			wavFiles = append(wavFiles, path)
		}
		return nil
	})

	if err != nil {
		log.Fatalf("扫描目录失败: %v", err)
	}

	if len(wavFiles) == 0 {
		fmt.Println("未找到WAV文件")
		return
	}

	fmt.Printf("找到 %d 个WAV文件\n\n", len(wavFiles))

	// 创建噪声过滤器
	filter, err := rnnoise.NewNoiseFilter("")
	if err != nil {
		log.Fatalf("创建噪声过滤器失败: %v", err)
	}
	defer filter.Destroy()

	// 处理每个文件
	totalStartTime := time.Now()
	successCount := 0

	for i, inputFile := range wavFiles {
		fmt.Printf("处理 [%d/%d]: %s\n", i+1, len(wavFiles), filepath.Base(inputFile))

		// 生成输出文件名
		relPath, _ := filepath.Rel(inputDir, inputFile)
		outputFile := filepath.Join(outputDir, relPath)

		// 确保输出目录存在
		outputFileDir := filepath.Dir(outputFile)
		if err := os.MkdirAll(outputFileDir, 0755); err != nil {
			log.Printf("  创建目录失败: %v", err)
			continue
		}

		// 处理文件
		startTime := time.Now()
		result, err := filter.FilterAudioFile(inputFile, outputFile, 0.0)
		if err != nil {
			log.Printf("  处理失败: %v", err)
			continue
		}
		elapsed := time.Since(startTime)

		// 显示进度
		audioDuration := float64(len(result.DenoisedAudio.Samples)) / float64(result.DenoisedAudio.SampleRate)
		fmt.Printf("  完成: %.2fs (%.1fx 实时)\n", elapsed.Seconds(), audioDuration/elapsed.Seconds())

		successCount++
	}

	totalElapsed := time.Since(totalStartTime)
	fmt.Printf("\n批量处理完成!\n")
	fmt.Printf("总时间: %.2f 秒\n", totalElapsed.Seconds())
	fmt.Printf("成功处理: %d/%d 文件\n", successCount, len(wavFiles))
}

func runPerformanceTest(audioPath, batchSizeStr, numRunsStr string) {
	// 解析参数
	batchSize, err := strconv.Atoi(batchSizeStr)
	if err != nil {
		log.Fatalf("无效的批处理大小: %v", err)
	}

	numRuns, err := strconv.Atoi(numRunsStr)
	if err != nil {
		log.Fatalf("无效的测试次数: %v", err)
	}

	fmt.Printf("开始性能测试...\n")
	fmt.Printf("音频文件: %s\n", audioPath)
	fmt.Printf("批处理大小: %d\n", batchSize)
	fmt.Printf("测试次数: %d\n", numRuns)
	fmt.Println("---")

	var totalTime time.Duration

	for i := 0; i < numRuns; i++ {
		startTime := time.Now()

		err := performanceTestMain(audioPath, batchSize)
		if err != nil {
			log.Printf("第 %d 次测试失败: %v", i+1, err)
			continue
		}

		elapsed := time.Since(startTime)
		totalTime += elapsed

		fmt.Printf("第 %d 次: %.4f 秒\n", i+1, elapsed.Seconds())
	}

	avgTime := totalTime / time.Duration(numRuns)
	fmt.Println("---")
	fmt.Printf("总时间: %.4f 秒\n", totalTime.Seconds())
	fmt.Printf("平均时间: %.4f 秒\n", avgTime.Seconds())
	fmt.Printf("每秒处理次数: %.2f\n", 1.0/avgTime.Seconds())
}

// performanceTestMain 对应Python的main函数，进行批量音频处理
func performanceTestMain(audioPath string, batchSize int) error {
	// 读取音频文件
	audioData, err := ioutil.ReadFile(audioPath)
	if err != nil {
		return fmt.Errorf("无法读取音频文件: %v", err)
	}

	// 跳过WAV头部（44字节），获取PCM数据
	if len(audioData) < 44 {
		return fmt.Errorf("音频文件太小，不是有效的WAV文件")
	}
	pcmData := audioData[44:]

	// 分割为640字节的块（对应Python的chunk_size = 640）
	chunkSize := 640
	var pcmList [][]byte
	for i := 0; i < len(pcmData); i += chunkSize {
		end := i + chunkSize
		if end > len(pcmData) {
			end = len(pcmData)
		}
		pcmList = append(pcmList, pcmData[i:end])
	}

	// 创建噪声过滤器
	filter, err := rnnoise.NewNoiseFilter("")
	if err != nil {
		return fmt.Errorf("创建噪声过滤器失败: %v", err)
	}
	defer filter.Destroy()

	// 按批次处理
	for i := 0; i < len(pcmList); i += batchSize {
		end := i + batchSize
		if end > len(pcmList) {
			end = len(pcmList)
		}

		// 合并批次数据
		var batchData []byte
		for j := i; j < end; j++ {
			batchData = append(batchData, pcmList[j]...)
		}

		// 进行降噪处理（对应Python的filter调用）
		// 参数：音频字节、采样率8000、单声道、16位、语音概率阈值0.5
		_, _, err := filter.FilterAudioBytes(batchData, 8000, 1, 16, 0.5)
		if err != nil {
			return fmt.Errorf("音频处理失败: %v", err)
		}
	}

	return nil
}
