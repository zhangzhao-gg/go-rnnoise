# Go RNNoise

[![Go Version](https://img.shields.io/badge/Go-%3E%3D1.19-blue.svg)](https://golang.org/)
[![License](https://img.shields.io/badge/license-MIT-green.svg)](LICENSE)
[![Go Report Card](https://goreportcard.com/badge/github.com/zhangzhao-gg/go-rnnoise)](https://goreportcard.com/report/github.com/zhangzhao-gg/go-rnnoise)

Go RNNoise 是一个基于 [RNNoise](https://github.com/xiph/rnnoise) 的 Go 语言音频降噪库。RNNoise 是 Mozilla 开发的一个基于深度学习的实时噪声抑制库，专门用于语音通话和音频处理。

## 特性

- 🎯 **实时音频降噪**: 基于深度学习的高质量噪声抑制
- 🚀 **高性能**: 优化的 Go 实现，支持批量处理
- 🎵 **多格式支持**: 支持 WAV 文件和原始 PCM 数据
- 📊 **语音检测**: 提供每帧的语音概率分析
- 🔧 **灵活配置**: 支持自定义语音概率阈值
- 📱 **跨平台**: 支持 Linux、macOS 和 Windows

## 安装

```bash
go get github.com/zhangzhao-gg/go-rnnoise
```

## 快速开始

### 基本用法

```go
package main

import (
    "fmt"
    "log"
    
    "github.com/zhangzhao-gg/go-rnnoise/rnnoise"
)

func main() {
    // 创建噪声过滤器
    filter, err := rnnoise.NewNoiseFilter("")
    if err != nil {
        log.Fatal(err)
    }
    defer filter.Destroy()

    // 处理音频文件
    result, err := filter.FilterAudioFile("input.wav", "output.wav", 0.3)
    if err != nil {
        log.Fatal(err)
    }

    fmt.Printf("处理完成！处理了 %d 帧\n", result.ProcessedFrames)
}
```

### 处理原始音频数据

```go
// 读取原始 PCM 数据
audioBytes := []byte{...} // 你的音频数据

// 处理音频（8000Hz, 单声道, 16位）
denoisedBytes, voiceProbs, err := filter.FilterAudioBytes(
    audioBytes, 
    8000,  // 采样率
    1,     // 声道数
    16,    // 位深度
    0.3,   // 语音概率阈值
)

if err != nil {
    log.Fatal(err)
}

fmt.Printf("降噪完成，检测到 %d 帧语音\n", len(voiceProbs))
```

### 流式处理

```go
// 处理单个音频帧（480个样本，10ms @ 48kHz）
frame := make([]float32, 480)
// ... 填充音频数据

denoisedFrame, voiceProb, keepFrame, err := filter.FilterStream(frame, 0.3)
if err != nil {
    log.Fatal(err)
}

if keepFrame {
    // 保留这个帧
    fmt.Printf("语音概率: %.3f\n", voiceProb)
}
```

## 项目结构

```
go-rnnoise/
├── cmd/
│   └── rnnoise-cli/          # 命令行工具
├── examples/                 # 使用示例
│   ├── basic/               # 基础示例
│   ├── advanced/            # 高级示例
│   └── streaming/           # 流式处理示例
├── rnnoise/                 # 核心库
├── lib/                     # RNNoise 动态库
└── ... (配置文件)
```

## 命令行工具

项目包含一个完整的命令行工具，支持多种操作：

```bash
# 对音频文件进行降噪
go run ./cmd/rnnoise-cli denoise input.wav output.wav 0.3

# 分析音频文件的语音/噪声统计
go run ./cmd/rnnoise-cli analyze input.wav

# 批量处理目录中的所有 WAV 文件
go run ./cmd/rnnoise-cli batch ./input_dir ./output_dir

# 性能测试
go run ./cmd/rnnoise-cli test test.wav 1 10

# 流式处理演示
go run ./cmd/rnnoise-cli stream
```

## API 文档

### 核心类型

#### `NoiseFilter`
主要的噪声过滤器结构体。

```go
type NoiseFilter struct {
    rnnoise   *RNNoise
    processor *AudioProcessor
}
```

#### `AudioData`
音频数据结构。

```go
type AudioData struct {
    Samples    []float32 // 音频样本数据
    SampleRate int       // 采样率
    Channels   int       // 声道数
    BitDepth   int       // 位深度
}
```

#### `FilterResult`
过滤结果结构。

```go
type FilterResult struct {
    DenoisedAudio      *AudioData // 降噪后的音频
    VoiceProbabilities []float32  // 每帧的语音概率
    ProcessedFrames    int        // 处理的帧数
}
```

### 主要方法

#### `NewNoiseFilter(libPath string) (*NoiseFilter, error)`
创建新的噪声过滤器。如果 `libPath` 为空，会自动查找库文件。

#### `FilterAudioFile(inputFile, outputFile string, voiceProbThreshold float32) (*FilterResult, error)`
直接处理音频文件。

#### `FilterAudioBytes(audioBytes []byte, sampleRate, channels, bitDepth int, voiceProbThreshold float32) ([]byte, []float32, error)`
处理原始 PCM 音频数据。

#### `FilterStream(frame []float32, voiceProbThreshold float32) ([]float32, float32, bool, error)`
流式处理单个音频帧。

#### `AnalyzeFrames(audioData *AudioData, voiceProbThreshold float32) (*FrameStatistics, error)`
分析音频帧的统计信息。

## 技术细节

### 音频格式要求

- **输入**: 支持多种采样率（8000Hz, 16000Hz, 44100Hz, 48000Hz 等）
- **处理**: 内部转换为 48kHz 单声道进行处理
- **输出**: 可转换回原始采样率

### 帧处理

- 每帧包含 480 个样本（10ms @ 48kHz）
- 支持语音概率阈值过滤
- 提供实时语音检测

### 性能优化

- 批量处理支持
- 内存优化的音频转换
- 高效的 C 库绑定

## 依赖项

- [go-audio](https://github.com/go-audio/audio) - 音频格式支持
- [go-audio/wav](https://github.com/go-audio/wav) - WAV 文件处理
- [logrus](https://github.com/sirupsen/logrus) - 日志记录

## 构建要求

- Go 1.19 或更高版本
- C 编译器（用于 CGO）
- RNNoise 动态库文件

## 库文件

项目包含预编译的 RNNoise 库文件：
- `lib/librnnoise_5h_b_500k.so.0.4.1` (Linux/macOS)

如果需要其他平台的库文件，请从 [RNNoise 官方仓库](https://github.com/xiph/rnnoise) 获取。

## 示例

项目提供了多个使用示例：

### 基础示例
```bash
# 简单音频降噪
go run ./examples/basic/simple.go
```

### 高级示例
```bash
# 批量处理示例
go run ./examples/advanced/batch_processing.go
```

### 流式处理示例
```bash
# 实时流式处理
go run ./examples/streaming/real_time.go
```

### 构建所有示例
```bash
make examples
```

## 贡献

欢迎提交 Issue 和 Pull Request！

## 许可证

本项目基于 MIT 许可证开源。详见 [LICENSE](LICENSE) 文件。

## 致谢

- [RNNoise](https://github.com/xiph/rnnoise) - Mozilla 的原始 RNNoise 实现
- [Mozilla Research](https://research.mozilla.org/) - 深度学习音频处理研究

## 相关链接

- [RNNoise 论文](https://arxiv.org/abs/1709.08243)
- [Mozilla DeepSpeech](https://github.com/mozilla/DeepSpeech)
- [WebRTC](https://webrtc.org/) - 实时通信技术
