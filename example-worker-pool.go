package main

import (
	"context"
	"fmt"
	"math/rand"
	"sync"
	"time"
	"github.com/hellyaxs/miniyoutube/pkg/workerpool"
)

type NumeroJob struct {
	Numero int
}

type Resultado struct {
	valor int
	workerID int
	Timestamp time.Time
}

func procesarNumero(ctx context.Context,job workerpool.Job) workerpool.Result {
	numero := job.(NumeroJob).Numero
	workerId := numero % 3
 
	sleepTime := time.Duration(880 +rand.Intn(400)) * time.Millisecond
	time.Sleep(sleepTime)
	return Resultado{
		valor: numero,
		workerID: workerId,
		Timestamp: time.Now(),
	}
}

func main() {
	valorMaximo := 100
	bufferSize := 100
	pool := workerpool.NewWorkerPool(
		workerpool.Config{ WorkerCount: 100},
		procesarNumero,
	)
	inputCh := make(chan workerpool.Job, bufferSize)
	ctx := context.Background()
	resultCh, err := pool.Start(ctx, inputCh)
	if err != nil {
		panic(err)
	}

	var wg sync.WaitGroup
	wg.Add(valorMaximo)
	fmt.Println("Iniciando processamento")

	go func() {
		for i := range valorMaximo {
			inputCh <- NumeroJob{Numero: i}
		}
		close(inputCh)
	}()

	go func() {
		for result := range resultCh {
			resultado := result.(Resultado)
			fmt.Printf("Resultado: %d (worker %d) - %s\n", resultado.valor, resultado.workerID, resultado.Timestamp.Format(time.StampMilli))
			wg.Done()
		}
	}()

	wg.Wait()
	fmt.Println("Procesamiento finalizado")

    
}
