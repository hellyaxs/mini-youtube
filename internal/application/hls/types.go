package hls

import "github.com/hellyaxs/miniyoutube/internal/domain/entity"

// HLSOutputFile representa um arquivo gerado pelo encoding HLS com seu tipo.
type HLSOutputFile struct {
	Path string // caminho absoluto ou relativo do arquivo
	Type string // entity.FileTypeManifest ou entity.FileTypeSegment
}

// HLSOptions configuração do encoding HLS.
type HLSOptions struct {
	SegmentDurationSec float64 // duração alvo de cada segmento em segundos (ex: 10)
	SegmentFilename    string  // padrão para nomes dos segmentos (ex: "segment_%03d.ts")
	ManifestName       string  // nome do arquivo manifest (ex: "index.m3u8")
	CodecCopy          bool    // se true, usa -codec copy (rápido, sem re-encode)
}

// DefaultHLSOptions retorna opções padrão para HLS.
func DefaultHLSOptions() HLSOptions {
	return HLSOptions{
		SegmentDurationSec: 10,
		SegmentFilename:    "segment_%03d.ts",
		ManifestName:       "index.m3u8",
		CodecCopy:          true,
	}
}

// HLSResult resultado do encoding HLS com arquivos classificados por tipo.
type HLSResult struct {
	ManifestPath string          // caminho do arquivo .m3u8
	Segments     []string        // caminhos dos arquivos .ts
	AllFiles     []HLSOutputFile // todos os arquivos com tipo (manifest ou segment)
}

// IsManifest retorna true se o tipo for manifest.
func (f HLSOutputFile) IsManifest() bool { return f.Type == entity.FileTypeManifest }

// IsSegment retorna true se o tipo for segment.
func (f HLSOutputFile) IsSegment() bool { return f.Type == entity.FileTypeSegment }
