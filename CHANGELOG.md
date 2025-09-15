# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- 初始版本发布
- 支持 WAV 文件音频降噪处理
- 支持原始 PCM 音频数据处理
- 流式音频处理支持
- 语音概率检测和分析
- 批量音频文件处理
- 命令行工具支持
- 跨平台库文件支持 (Linux, macOS, Windows)
- 完整的 API 文档和示例

### Technical Details
- 基于 RNNoise C 库的 Go 绑定
- 支持多种采样率自动转换 (8000Hz, 16000Hz, 44100Hz, 48000Hz)
- 支持单声道和立体声音频处理
- 支持 16位、24位、32位音频格式
- 10ms 音频帧处理 (480 samples @ 48kHz)
- 高性能批量处理支持
- 内存优化的音频转换算法

### Dependencies
- Go 1.19+
- go-audio/audio v1.0.0
- go-audio/wav v1.1.0
- sirupsen/logrus v1.9.3

## [1.0.0] - 2024-01-XX

### Added
- 🎯 基于深度学习的实时音频降噪
- 🚀 高性能 Go 实现
- 🎵 多格式音频支持
- 📊 语音概率检测
- 🔧 灵活的配置选项
- 📱 跨平台支持
- 📖 完整的文档和示例
