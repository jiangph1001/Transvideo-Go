package config

import (
	"os"
	"path/filepath"
	"strings"
)

// EncoderType 编码器类型
type EncoderType string

const (
	EncoderSoftware EncoderType = "software" // 软件编码
	EncoderNVENC    EncoderType = "nvenc"    // NVIDIA 硬件编码
	EncoderQSV      EncoderType = "qsv"      // Intel QSV 硬件编码
	EncoderAMF      EncoderType = "amf"      // AMD AMF 硬件编码
)

// VideoCodec 视频编码格式
type VideoCodec string

const (
	CodecH264 VideoCodec = "h264"
	CodecHEVC VideoCodec = "hevc"
	CodecAV1  VideoCodec = "av1"
)

type Config struct {
	MaxConcurrency int
	DockerHost     string
	InputDir       string // 宿主机输入目录
	OutputDir      string // 宿主机输出目录

	// 编码器配置
	EncoderType   EncoderType // 编码器类型: software, nvenc, qsv, amf
	VideoCodec    VideoCodec  // 视频编码格式: h264, hevc, av1
	SoftwarePreset string     // 软件编码预设 (svtav1: 0-12)
	HardwarePreset string     // 硬件编码预设 (nvenc: p1-p7, qsv: veryfast-veryslow)
	CRF          string       // CRF 质量因子
	AudioCodec   string       // 音频编码器
	AudioBitrate string       // 音频码率
}

func Load() *Config {
	cfg := &Config{
		MaxConcurrency: 2,
		DockerHost:     os.Getenv("DOCKER_HOST"),
	}

	// 输入输出目录
	cfg.InputDir = getEnvDefault("INPUT_DIR", "/input")
	cfg.OutputDir = getEnvDefault("OUTPUT_DIR", "/output")

	// 转换为绝对路径
	if !filepath.IsAbs(cfg.InputDir) {
		cfg.InputDir, _ = filepath.Abs(cfg.InputDir)
	}
	if !filepath.IsAbs(cfg.OutputDir) {
		cfg.OutputDir, _ = filepath.Abs(cfg.OutputDir)
	}

	// 编码器类型（software, nvenc, qsv, amf）
	encoderType := getEnvDefault("ENCODER_TYPE", "software")
	cfg.EncoderType = EncoderType(encoderType)

	// 视频编码格式（h264, hevc, av1）
	videoCodec := getEnvDefault("VIDEO_CODEC", "av1")
	cfg.VideoCodec = VideoCodec(videoCodec)

	// 软件编码预设（svtav1: 0-12，默认6）
	cfg.SoftwarePreset = getEnvDefault("SOFTWARE_PRESET", "6")

	// 硬件编码预设
	// NVIDIA: p1-p7（p1最快，p7最慢质量最高）
	// Intel QSV: veryfast, faster, fast, medium, slow, slower, veryslow
	cfg.HardwarePreset = getEnvDefault("HARDWARE_PRESET", "medium")

	// CRF（质量因子）
	// H.264/H.265: 0-51（默认23）
	// AV1: 0-63（默认35）
	cfg.CRF = getEnvDefault("CRF", "")

	// 音频编码器
	cfg.AudioCodec = getEnvDefault("AUDIO_CODEC", "libopus")

	// 音频码率
	cfg.AudioBitrate = getEnvDefault("AUDIO_BITRATE", "128k")

	// 根据编码器类型和视频格式设置默认值
	setDefaults(cfg)

	return cfg
}

// getEnvDefault 获取环境变量，如果不存在返回默认值
func getEnvDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// setDefaults 根据编码器类型和视频格式设置合理的默认值
func setDefaults(cfg *Config) {
	// 如果 CRF 没有设置，根据编码格式设置合适的默认值
	if cfg.CRF == "" {
		switch cfg.VideoCodec {
		case CodecH264:
			cfg.CRF = "23" // H.264 的默认值
		case CodecHEVC:
			cfg.CRF = "28" // H.265 的默认值
		case CodecAV1:
			cfg.CRF = "35" // AV1 的默认值
		}
	}

	// 验证编码器类型是否有效
	validEncoders := []EncoderType{EncoderSoftware, EncoderNVENC, EncoderQSV, EncoderAMF}
	isValidEncoder := false
	for _, e := range validEncoders {
		if strings.ToLower(string(cfg.EncoderType)) == string(e) {
			isValidEncoder = true
			break
		}
	}
	if !isValidEncoder {
		cfg.EncoderType = EncoderSoftware
	}

	// 验证视频编码格式
	validCodecs := []VideoCodec{CodecH264, CodecHEVC, CodecAV1}
	isValidCodec := false
	for _, c := range validCodecs {
		if strings.ToLower(string(cfg.VideoCodec)) == string(c) {
			isValidCodec = true
			break
		}
	}
	if !isValidCodec {
		cfg.VideoCodec = CodecAV1
	}
}

