package rnnoise

/*
#include <stdlib.h>
#include <dlfcn.h>

// RNNoise C函数指针类型定义
typedef struct RNNoiseState RNNoiseState;
typedef RNNoiseState* (*rnnoise_create_func)(void *model);
typedef void (*rnnoise_destroy_func)(RNNoiseState *st);
typedef float (*rnnoise_process_frame_func)(RNNoiseState *st, float *out, const float *in);

// 全局变量存储动态库句柄和函数指针
static void* rnnoise_handle = NULL;
static rnnoise_create_func rnnoise_create_ptr = NULL;
static rnnoise_destroy_func rnnoise_destroy_ptr = NULL;
static rnnoise_process_frame_func rnnoise_process_frame_ptr = NULL;

// 加载RNNoise库
int load_rnnoise_library(const char* lib_path) {
    rnnoise_handle = dlopen(lib_path, RTLD_LAZY);
    if (!rnnoise_handle) {
        return 0; // 失败
    }

    rnnoise_create_ptr = (rnnoise_create_func)dlsym(rnnoise_handle, "rnnoise_create");
    rnnoise_destroy_ptr = (rnnoise_destroy_func)dlsym(rnnoise_handle, "rnnoise_destroy");
    rnnoise_process_frame_ptr = (rnnoise_process_frame_func)dlsym(rnnoise_handle, "rnnoise_process_frame");

    if (!rnnoise_create_ptr || !rnnoise_destroy_ptr || !rnnoise_process_frame_ptr) {
        dlclose(rnnoise_handle);
        rnnoise_handle = NULL;
        return 0; // 失败
    }

    return 1; // 成功
}

// 释放库
void unload_rnnoise_library() {
    if (rnnoise_handle) {
        dlclose(rnnoise_handle);
        rnnoise_handle = NULL;
    }
}

// 包装函数
RNNoiseState* rnnoise_create(void *model) {
    if (rnnoise_create_ptr) {
        return rnnoise_create_ptr(model);
    }
    return NULL;
}

void rnnoise_destroy(RNNoiseState *st) {
    if (rnnoise_destroy_ptr) {
        rnnoise_destroy_ptr(st);
    }
}

float rnnoise_process_frame(RNNoiseState *st, float *out, const float *in) {
    if (rnnoise_process_frame_ptr) {
        return rnnoise_process_frame_ptr(st, out, in);
    }
    return 0.0;
}
*/
import "C"
import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"unsafe"
)

// RNNoise 结构体，用于封装RNNoise降噪功能
// RNNoise 是 Mozilla 开发的一个基于深度学习的实时噪声抑制库
// 它专门用于语音通话和音频处理，能够有效去除背景噪声
type RNNoise struct {
	state           *C.RNNoiseState
	SampleWidth     int // 采样位宽（字节）
	Channels        int // 声道数
	SampleRate      int // 采样率
	FrameDurationMS int // 帧持续时间（毫秒）
}

// NewRNNoise 创建新的RNNoise实例
//
// 参数:
//   - libPath: RNNoise动态库文件路径，如果为空则自动查找
//
// 返回:
//   - *RNNoise: RNNoise实例指针
//   - error: 创建失败时的错误信息
//
// 示例:
//
//	// 自动查找库文件
//	rnn, err := NewRNNoise("")
//
//	// 指定库文件路径
//	rnn, err := NewRNNoise("/path/to/librnnoise.so")
func NewRNNoise(libPath string) (*RNNoise, error) {
	// 如果没有指定库路径，自动查找
	if libPath == "" {
		var err error
		libPath, err = findRNNoiseLib()
		if err != nil {
			return nil, fmt.Errorf("无法找到RNNoise库: %v", err)
		}
	}

	// 验证库文件是否存在
	if _, err := os.Stat(libPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("RNNoise库文件不存在: %s", libPath)
	}

	// 动态加载RNNoise库
	libPathC := C.CString(libPath)
	defer C.free(unsafe.Pointer(libPathC))

	if C.load_rnnoise_library(libPathC) == 0 {
		return nil, fmt.Errorf("无法加载RNNoise库: %s", libPath)
	}

	// 创建RNNoise状态对象
	state := C.rnnoise_create(nil)
	if state == nil {
		C.unload_rnnoise_library()
		return nil, fmt.Errorf("无法创建RNNoise状态对象")
	}

	return &RNNoise{
		state:           state,
		SampleWidth:     2,     // 16位 = 2字节
		Channels:        1,     // 单声道
		SampleRate:      48000, // 48kHz
		FrameDurationMS: 10,    // 10毫秒帧
	}, nil
}

// Destroy 销毁RNNoise实例，释放资源
//
// 这个方法会释放RNNoise状态对象和动态库资源
// 应该在不再使用RNNoise实例时调用此方法
func (r *RNNoise) Destroy() {
	if r.state != nil {
		C.rnnoise_destroy(r.state)
		r.state = nil
	}
	// 释放动态库（注意：这会影响所有实例，在实际应用中可能需要引用计数）
	C.unload_rnnoise_library()
}

// Reset 重置RNNoise状态，清除神经网络的内部状态
func (r *RNNoise) Reset() error {
	if r.state != nil {
		C.rnnoise_destroy(r.state)
	}

	r.state = C.rnnoise_create(nil)
	if r.state == nil {
		return fmt.Errorf("无法重置RNNoise状态对象")
	}

	return nil
}

// ProcessFrame 处理单个音频帧（10毫秒）
//
// 参数:
//   - frame: 480个float32样本（10ms @ 48kHz），范围-1.0到1.0
//
// 返回:
//   - float32: 语音概率 (0.0-1.0)
//   - []float32: 降噪后的音频帧（480个样本）
//   - error: 处理失败时的错误信息
//
// 注意: 输入帧必须恰好包含480个样本，对应48kHz采样率下的10毫秒音频
func (r *RNNoise) ProcessFrame(frame []float32) (float32, []float32, error) {
	if len(frame) != 480 {
		return 0, nil, fmt.Errorf("帧大小必须为480个样本（10ms @ 48kHz），当前为%d", len(frame))
	}

	// 将float32样本转换为RNNoise期望的格式（16位整数范围的float）
	rnnoiseInput := make([]float32, 480)
	rnnoiseOutput := make([]float32, 480)

	for i, sample := range frame {
		// 将-1.0到1.0范围转换为-32768到32767范围
		rnnoiseInput[i] = sample * 32768.0
	}

	// 调用RNNoise处理函数
	voiceProb := C.rnnoise_process_frame(
		r.state,
		(*C.float)(unsafe.Pointer(&rnnoiseOutput[0])),
		(*C.float)(unsafe.Pointer(&rnnoiseInput[0])),
	)

	// 将输出转换回-1.0到1.0范围
	output := make([]float32, 480)
	for i, sample := range rnnoiseOutput {
		output[i] = sample / 32768.0
	}

	return float32(voiceProb), output, nil
}

// findRNNoiseLib 自动查找RNNoise库文件
func findRNNoiseLib() (string, error) {
	// 根据操作系统确定库文件名
	var libName string
	switch runtime.GOOS {
	case "linux", "darwin":
		libName = "librnnoise_5h_b_500k.so.0.4.1"
	case "windows":
		libName = "librnnoise_5h_b_500k.dll"
	default:
		return "", fmt.Errorf("不支持的操作系统: %s", runtime.GOOS)
	}

	// 优先在lib目录中查找
	libPath := filepath.Join("lib", libName)
	if _, err := os.Stat(libPath); err == nil {
		absPath, err := filepath.Abs(libPath)
		if err != nil {
			return "", err
		}
		return absPath, nil
	}

	// 如果lib目录没有，再在整个项目中搜索
	var foundPath string
	err := filepath.Walk(".", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.Name() == libName {
			foundPath = path
			return filepath.SkipDir // 找到后停止搜索
		}
		return nil
	})

	if err != nil {
		return "", err
	}

	if foundPath == "" {
		return "", fmt.Errorf("未找到库文件: %s", libName)
	}

	// 返回绝对路径
	absPath, err := filepath.Abs(foundPath)
	if err != nil {
		return "", err
	}

	return absPath, nil
}
