package hls

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"

	"github.com/hellyaxs/miniyoutube/internal/domain/entity"
)

// Service wrapper do ffmpeg para geração de HLS com output tipado (manifest vs segment).
type Service struct {
	ffmpegBin string // caminho do binário ffmpeg (ex: "ffmpeg")
}

// NewService cria um serviço com o binário ffmpeg informado. Passar "" usa "ffmpeg" do PATH.
func NewService(ffmpegBin string) *Service {
	if ffmpegBin == "" {
		ffmpegBin = "ffmpeg"
	}
	return &Service{ffmpegBin: ffmpegBin}
}

// EncodeToHLS gera HLS a partir do vídeo de entrada, gravando manifest e segmentos em outputDir.
// Retorna os caminhos do manifest e dos segmentos e a lista de todos os arquivos com tipo (manifest/segment).
func (s *Service) EncodeToHLS(ctx context.Context, inputPath, outputDir string, opts HLSOptions) (*HLSResult, error) {
	if opts.SegmentDurationSec <= 0 {
		opts = DefaultHLSOptions()
	}
	if opts.SegmentFilename == "" {
		opts.SegmentFilename = "segment_%03d.ts"
	}
	if opts.ManifestName == "" {
		opts.ManifestName = "index.m3u8"
	}

	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return nil, fmt.Errorf("criar diretório de saída: %w", err)
	}

	manifestPath := filepath.Join(outputDir, opts.ManifestName)
	segmentPattern := filepath.Join(outputDir, opts.SegmentFilename)

	args := []string{"-y", "-i", inputPath}
	if opts.CodecCopy {
		args = append(args, "-codec", "copy")
	}
	args = append(args,
		"-hls_time", fmt.Sprintf("%.2f", opts.SegmentDurationSec),
		"-hls_list_size", "0",
		"-hls_segment_filename", segmentPattern,
		"-f", "hls",
		manifestPath,
	)

	cmd := exec.CommandContext(ctx, s.ffmpegBin, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("executar ffmpeg: %w", err)
	}

	// Coleta arquivos gerados e classifica por tipo
	segments, allFiles, err := listHLSOutputs(outputDir, opts.ManifestName, opts.SegmentFilename)
	if err != nil {
		return nil, fmt.Errorf("listar saídas HLS: %w", err)
	}

	return &HLSResult{
		ManifestPath: manifestPath,
		Segments:     segments,
		AllFiles:     allFiles,
	}, nil
}

// listHLSOutputs percorre outputDir e retorna segmentos ordenados e todos os arquivos com tipo.
func listHLSOutputs(outputDir, manifestName, _ string) (segments []string, allFiles []HLSOutputFile, err error) {
	entries, err := os.ReadDir(outputDir)
	if err != nil {
		return nil, nil, err
	}

	var segmentPaths []string
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		fullPath := filepath.Join(outputDir, name)
		if name == manifestName || strings.HasSuffix(name, ".m3u8") {
			allFiles = append(allFiles, HLSOutputFile{Path: fullPath, Type: entity.FileTypeManifest})
			continue
		}
		if strings.HasSuffix(name, ".ts") {
			segmentPaths = append(segmentPaths, fullPath)
			allFiles = append(allFiles, HLSOutputFile{Path: fullPath, Type: entity.FileTypeSegment})
		}
	}
	sort.Slice(segmentPaths, func(i, j int) bool { return segmentPaths[i] < segmentPaths[j] })
	return segmentPaths, allFiles, nil
}
