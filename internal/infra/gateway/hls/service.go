package hls

import (
	"context"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"

	"github.com/hellyaxs/miniyoutube/internal/domain/entity"
)

// Service é um wrapper do ffmpeg para geração de HLS multi-bitrate.
type Service struct {
	ffmpegBin string
}

// NewService cria um Service com o binário ffmpeg informado. Passar "" usa "ffmpeg" do PATH.
func NewService(ffmpegBin string) *Service {
	if ffmpegBin == "" {
		ffmpegBin = "ffmpeg"
	}
	return &Service{ffmpegBin: ffmpegBin}
}

// EncodeToHLS converte inputPath para HLS multi-bitrate em outputDir.
// Gera um subdiretório por rendição (ex: outputDir/1080p/, outputDir/720p/) e um master.m3u8 na raiz.
func (s *Service) EncodeToHLS(ctx context.Context, inputPath, outputDir string, opts HLSOptions) (*HLSResult, error) {
	opts = applyDefaults(opts)

	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return nil, fmt.Errorf("criar outputDir: %w", err)
	}
	for _, r := range opts.Renditions {
		if err := os.MkdirAll(filepath.Join(outputDir, r.Name), 0755); err != nil {
			return nil, fmt.Errorf("criar dir %s: %w", r.Name, err)
		}
	}

	args := buildMultiBitrateArgs(inputPath, outputDir, opts)
	cmd := exec.CommandContext(ctx, s.ffmpegBin, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("executar ffmpeg: %w", err)
	}

	segments, allFiles, err := walkHLSOutputs(outputDir)
	if err != nil {
		return nil, fmt.Errorf("listar saídas HLS: %w", err)
	}

	return &HLSResult{
		ManifestPath: filepath.Join(outputDir, opts.ManifestName),
		Segments:     segments,
		AllFiles:     allFiles,
	}, nil
}

// applyDefaults preenche campos zero-value com os valores padrão.
func applyDefaults(opts HLSOptions) HLSOptions {
	if opts.SegmentDurationSec <= 0 {
		opts.SegmentDurationSec = 6
	}
	if opts.ManifestName == "" {
		opts.ManifestName = "master.m3u8"
	}
	if len(opts.Renditions) == 0 {
		opts.Renditions = DefaultRenditions
	}
	return opts
}

// buildMultiBitrateArgs monta os argumentos ffmpeg para encoding multi-bitrate com filter_complex.
// Estrutura de saída:
//
//	outputDir/
//	  master.m3u8
//	  1080p/playlist.m3u8  +  1080p/seg_000.ts ...
//	  720p/playlist.m3u8   +  720p/seg_000.ts  ...
//	  480p/playlist.m3u8   +  480p/seg_000.ts  ...
func buildMultiBitrateArgs(inputPath, outputDir string, opts HLSOptions) []string {
	n := len(opts.Renditions)
	args := []string{"-y", "-i", inputPath}

	// filter_complex: split do stream de vídeo + scale por rendição
	// Ex: [0:v]split=3[v0][v1][v2];[v0]scale=1920:1080[s0];[v1]scale=1280:720[s1];...
	filter := fmt.Sprintf("[0:v]split=%d", n)
	for i := range opts.Renditions {
		filter += fmt.Sprintf("[v%d]", i)
	}
	filter += ";"
	for i, r := range opts.Renditions {
		filter += fmt.Sprintf("[v%d]scale=%d:%d[s%d]", i, r.Width, r.Height, i)
		if i < n-1 {
			filter += ";"
		}
	}
	args = append(args, "-filter_complex", filter)

	// Mapeamento de streams: vídeo escalonado + áudio original por rendição
	for i, r := range opts.Renditions {
		args = append(args,
			"-map", fmt.Sprintf("[s%d]", i),
			"-map", "0:a",
			fmt.Sprintf("-c:v:%d", i), "libx264",
			fmt.Sprintf("-b:v:%d", i), r.VideoBitrate,
			fmt.Sprintf("-c:a:%d", i), "aac",
			fmt.Sprintf("-b:a:%d", i), r.AudioBitrate,
		)
	}

	// var_stream_map: associa cada par (vídeo, áudio) ao subdiretório da rendição
	varStreamMap := ""
	for i, r := range opts.Renditions {
		if i > 0 {
			varStreamMap += " "
		}
		varStreamMap += fmt.Sprintf("v:%d,a:%d,name:%s", i, i, r.Name)
	}

	// %v é substituído pelo ffmpeg pelo campo "name" de cada stream
	segmentPattern := filepath.Join(outputDir, "%v", "seg_%03d.ts")
	playlistPattern := filepath.Join(outputDir, "%v", "playlist.m3u8")

	args = append(args,
		"-var_stream_map", varStreamMap,
		"-hls_time", fmt.Sprintf("%.2f", opts.SegmentDurationSec),
		"-hls_list_size", "0",
		"-hls_segment_filename", segmentPattern,
		"-master_pl_name", opts.ManifestName,
		"-f", "hls",
		playlistPattern,
	)

	return args
}

// walkHLSOutputs percorre outputDir recursivamente e classifica .m3u8 e .ts por tipo.
func walkHLSOutputs(outputDir string) (segments []string, allFiles []HLSOutputFile, err error) {
	err = filepath.WalkDir(outputDir, func(path string, d fs.DirEntry, e error) error {
		if e != nil || d.IsDir() {
			return e
		}
		name := d.Name()
		switch {
		case strings.HasSuffix(name, ".m3u8"):
			allFiles = append(allFiles, HLSOutputFile{Path: path, Type: entity.FileTypeManifest})
		case strings.HasSuffix(name, ".ts"):
			segments = append(segments, path)
			allFiles = append(allFiles, HLSOutputFile{Path: path, Type: entity.FileTypeSegment})
		}
		return nil
	})
	sort.Strings(segments)
	return
}
