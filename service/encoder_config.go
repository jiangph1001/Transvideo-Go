package service

import (
	"os"

	"video-transcoder/config"
)

// VideoEncoder 视频编码器配置
type VideoEncoder struct {
	Codec      string // FFmpeg 编码器名称
	Preset     string // 预设值参数
	CRF        string // CRF 值
	Bitrate    string // 码率
	ExtraArgs  []string // 额外参数
}

// AudioEncoder 音频编码器配置
type AudioEncoder struct {
	Codec   string
	Bitrate string
}

// GetVideoEncoder 根据配置获取视频编码器
func GetVideoEncoder(cfg *config.Config, bitrate string) *VideoEncoder {
	encoder := &VideoEncoder{
		Bitrate: bitrate,
		CRF:     cfg.CRF,
	}

	// 根据编码器类型和视频格式选择编码器
	switch cfg.EncoderType {
	case config.EncoderSoftware:
		return getSoftwareEncoder(cfg, encoder, bitrate)
	case config.EncoderNVENC:
		return getNVENCEncoder(cfg, encoder, bitrate)
	case config.EncoderQSV:
		return getQSVEncoder(cfg, encoder, bitrate)
	case config.EncoderAMF:
		return getAMFEncoder(cfg, encoder, bitrate)
	default:
		return getSoftwareEncoder(cfg, encoder, bitrate)
	}
}

// GetAudioEncoder 获取音频编码器
func GetAudioEncoder(cfg *config.Config) *AudioEncoder {
	return &AudioEncoder{
		Codec:   cfg.AudioCodec,
		Bitrate: cfg.AudioBitrate,
	}
}

// 软件编码器配置
func getSoftwareEncoder(cfg *config.Config, encoder *VideoEncoder, bitrate string) *VideoEncoder {
	switch cfg.VideoCodec {
	case config.CodecH264:
		encoder.Codec = "libx264"
		encoder.Preset = cfg.SoftwarePreset // 0-9, 0=ultrafast, 9=veryslow
	case config.CodecHEVC:
		encoder.Codec = "libx265"
		encoder.Preset = cfg.SoftwarePreset // 0-9
	case config.CodecAV1:
		encoder.Codec = "libsvtav1"
		encoder.Preset = cfg.SoftwarePreset // 0-12, 6=medium
	}
	return encoder
}

// NVIDIA NVENC 配置
func getNVENCEncoder(cfg *config.Config, encoder *VideoEncoder, bitrate string) *VideoEncoder {
	switch cfg.VideoCodec {
	case config.CodecH264:
		encoder.Codec = "h264_nvenc"
		encoder.Preset = cfg.HardwarePreset // p1-p7, p1=fastest, p7=best
	case config.CodecHEVC:
		encoder.Codec = "hevc_nvenc"
		encoder.Preset = cfg.HardwarePreset // p1-p7
	case config.CodecAV1:
		// 需要 RTX 40 系列
		encoder.Codec = "av1_nvenc"
		encoder.Preset = cfg.HardwarePreset // p1-p7
	}

	// NVENC 额外参数
	encoder.ExtraArgs = []string{
		"-rc:v", "vbr", // 可变码率
	}
	return encoder
}

// Intel QSV 配置
func getQSVEncoder(cfg *config.Config, encoder *VideoEncoder, bitrate string) *VideoEncoder {
	// QSV 需要初始化硬件设备
	encoder.ExtraArgs = []string{
		"-init_hw_device", "qsv=hw",
		"-filter_hw_device", "hw",
	}

	switch cfg.VideoCodec {
	case config.CodecH264:
		encoder.Codec = "h264_qsv"
		// QSV 没有 preset，使用 -preset 参数
		encoder.Preset = "" // QSV 的 preset 直接通过 -preset 参数传递
		encoder.ExtraArgs = append(encoder.ExtraArgs,
			"-preset", cfg.HardwarePreset) // veryfast, faster, fast, medium, slow, slower, veryslow
	case config.CodecHEVC:
		encoder.Codec = "hevc_qsv"
		encoder.Preset = ""
		encoder.ExtraArgs = append(encoder.ExtraArgs,
			"-preset", cfg.HardwarePreset)
	case config.CodecAV1:
		// AV1 QSV 需要较新的 Intel GPU（Arc 系列）
		encoder.Codec = "av1_qsv"
		encoder.Preset = ""
		encoder.ExtraArgs = append(encoder.ExtraArgs,
			"-preset", cfg.HardwarePreset)
	}

	return encoder
}

// AMD AMF 配置
func getAMFEncoder(cfg *config.Config, encoder *VideoEncoder, bitrate string) *VideoEncoder {
	switch cfg.VideoCodec {
	case config.CodecH264:
		encoder.Codec = "h264_amf"
		encoder.Preset = "" // AMF 使用 -quality 参数
		encoder.ExtraArgs = []string{
			"-quality", cfg.HardwarePreset, // speed, balanced, quality
		}
	case config.CodecHEVC:
		encoder.Codec = "hevc_amf"
		encoder.Preset = ""
		encoder.ExtraArgs = []string{
			"-quality", cfg.HardwarePreset,
		}
	case config.CodecAV1:
		// AV1 AMF 需要较新的 AMD GPU（RX 7000 系列）
		encoder.Codec = "av1_amf"
		encoder.Preset = ""
		encoder.ExtraArgs = []string{
			"-quality", cfg.HardwarePreset,
		}
	}

	// AMF 额外参数
	encoder.ExtraArgs = append(encoder.ExtraArgs,
		"-rc", "vbr_peak", // 峰值可变码率
	)

	return encoder
}

// BuildFFmpegArgs 构建 FFmpeg 命令行参数
func BuildFFmpegArgs(inputFile, outputFile string, videoEncoder *VideoEncoder, audioEncoder *AudioEncoder) []string {
	args := []string{
		ffmpegImage(),
	}

	// 视频输入
	args = append(args, "-i", inputFile)

	// 视频编码器参数
	args = append(args, "-c:v", videoEncoder.Codec)

	// 预设值（如果存在）
	if videoEncoder.Preset != "" {
		args = append(args, "-preset", videoEncoder.Preset)
	}

	// 额外参数
	if len(videoEncoder.ExtraArgs) > 0 {
		args = append(args, videoEncoder.ExtraArgs...)
	}

	// CRF（如果有）
	if videoEncoder.CRF != "" {
		args = append(args, "-crf", videoEncoder.CRF)
	}

	// 码率
	// 注意：SVT-AV1 在 CRF 模式（--rc 0）下不接受目标码率，会报错
	// "Target Bitrate only supported when --rc is 1/2" 并导致编码器启动失败
	isSvtAv1Crf := videoEncoder.Codec == "libsvtav1" && videoEncoder.CRF != ""
	if videoEncoder.Bitrate != "" && videoEncoder.Bitrate != "0" && !isSvtAv1Crf {
		args = append(args, "-b:v", videoEncoder.Bitrate)
	}

	// 音频编码器
	args = append(args, "-c:a", audioEncoder.Codec)

	// 音频码率
	if audioEncoder.Bitrate != "" {
		args = append(args, "-b:a", audioEncoder.Bitrate)
	}

	// 覆盖输出文件
	args = append(args, "-y")

	// 输出文件
	args = append(args, outputFile)

	return args
}

// ffmpegImage 返回 FFmpeg Docker 镜像名称
func ffmpegImage() string {
	// 可用环境变量 FFMPEG_IMAGE 覆盖，方便在不同架构/环境切换镜像
	if image := os.Getenv("FFMPEG_IMAGE"); image != "" {
		return image
	}
	// 默认使用 linuxserver/ffmpeg: 多架构（arm64 原生），且包含 libsvtav1/libx264/libx265/libopus
	return "linuxserver/ffmpeg:latest"
}
