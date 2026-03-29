package hls

import "github.com/hellyaxs/miniyoutube/internal/domain/entity"

// HLSOutputFile representa um arquivo gerado pelo encoding HLS com seu tipo.
type HLSOutputFile struct {
	Path string // caminho absoluto do arquivo
	Type string // entity.FileTypeManifest ou entity.FileTypeSegment
}

// IsManifest retorna true se o tipo for manifest.
func (f HLSOutputFile) IsManifest() bool { return f.Type == entity.FileTypeManifest }

// IsSegment retorna true se o tipo for segment.
func (f HLSOutputFile) IsSegment() bool { return f.Type == entity.FileTypeSegment }

// Rendition define uma variante de qualidade no encoding multi-bitrate.
type Rendition struct {
	Name         string // nome do subdiretório (ex: "1080p")
	Width        int
	Height       int
	VideoBitrate string // ex: "5000k"
	AudioBitrate string // ex: "192k"
	Bandwidth    int    // largura de banda em bits para o master.m3u8 (ex: 5192000)
}

// DefaultRenditions contém as três variantes padrão de qualidade.
var DefaultRenditions = []Rendition{
	{Name: "1080p", Width: 1920, Height: 1080, VideoBitrate: "5000k", AudioBitrate: "192k", Bandwidth: 5192000},
	{Name: "720p", Width: 1280, Height: 720, VideoBitrate: "2800k", AudioBitrate: "128k", Bandwidth: 2928000},
	{Name: "480p", Width: 854, Height: 480, VideoBitrate: "1400k", AudioBitrate: "96k", Bandwidth: 1496000},
}

// HLSOptions configuração do encoding HLS multi-bitrate.
type HLSOptions struct {
	SegmentDurationSec float64     // duração alvo de cada segmento em segundos (ex: 6)
	ManifestName       string      // nome do manifest master (ex: "master.m3u8")
	Renditions         []Rendition // variantes de qualidade a gerar
}

// DefaultHLSOptions retorna opções padrão para HLS multi-bitrate.
func DefaultHLSOptions() HLSOptions {
	return HLSOptions{
		SegmentDurationSec: 6,
		ManifestName:       "master.m3u8",
		Renditions:         DefaultRenditions,
	}
}

// HLSResult resultado do encoding HLS com arquivos classificados por tipo.
type HLSResult struct {
	ManifestPath string          // caminho do master.m3u8
	Segments     []string        // caminhos de todos os arquivos .ts (ordenados)
	AllFiles     []HLSOutputFile // todos os arquivos com tipo (manifest ou segment)
}
