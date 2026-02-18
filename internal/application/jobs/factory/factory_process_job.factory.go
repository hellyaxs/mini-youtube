package factory

import (
	"context"
	"fmt"
	"log/slog"
	"reflect"

	"github.com/hellyaxs/miniyoutube/pkg/workerpool"
)

// JobProcessor registry de handlers por tipo de job; usa generics para registrar qualquer job.
type JobProcessor struct {
	handlers map[reflect.Type]func(context.Context, workerpool.Job) workerpool.Result
}

// NewJobProcessor cria um registry vazio para registrar handlers de jobs.
func NewJobProcessor() *JobProcessor {
	return &JobProcessor{
		handlers: make(map[reflect.Type]func(context.Context, workerpool.Job) workerpool.Result),
	}
}

// Register associa um handler ao job do tipo J. Pode ser chamado várias vezes para tipos diferentes.
func Register[J workerpool.Job](p *JobProcessor, handler func(context.Context, J) workerpool.Result) {
	var zero J
	t := reflect.TypeOf(zero)
	p.handlers[t] = func(ctx context.Context, job workerpool.Job) workerpool.Result {
		j, ok := job.(J)
		if !ok {
			return nil
		}
		return handler(ctx, j)
	}
}

// Build retorna uma ProcessorFunc que despacha cada job para o handler registrado para aquele tipo.
func (p *JobProcessor) Build(logger *slog.Logger) workerpool.ProcessorFunc {
	if logger == nil {
		logger = slog.Default()
	}
	return func(ctx context.Context, job workerpool.Job) workerpool.Result {
		t := reflect.TypeOf(job)
		h, ok := p.handlers[t]
		if !ok {
			logger.Error("tipo de job não registrado", "type", fmt.Sprintf("%v", t))
			return nil
		}
		return h(ctx, job)
	}
}
