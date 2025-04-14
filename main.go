package main

import (
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

// Transaccion representa una transacción bancaria
type Transaccion struct {
	ID    string
	Tipo  string
	Monto float64
}

// Variables que podemos modificar para pruebas
var (
	osExit = os.Exit
	osArgs = os.Args
)

func main() {
	inicio := time.Now() // ⏱️ Comienza a medir tiempo

	code := ejecutarAplicacion(osArgs)

	duracion := time.Since(inicio) // ⏱️ Calcula duración
	fmt.Printf("⏱️ Tiempo total de ejecución: %s\n", duracion)

	if code != 0 {
		osExit(code)
	}
}

func ejecutarAplicacion(args []string) int {
	if len(args) < 2 {
		fmt.Println("Uso: go run main.go [archivo_csv]")
		return 1
	}

	archivoCSV := args[1]
	transacciones, err := leerTransaccionesDesdeCSV(archivoCSV)
	if err != nil {
		log.Printf("Error al leer el archivo CSV: %v", err)
		return 1
	}

	// Probar diferentes configuraciones de workers
	testWorkers(transacciones)

	return 0
}

func leerTransaccionesDesdeCSV(rutaArchivo string) ([]Transaccion, error) {
	archivo, err := os.Open(rutaArchivo)
	if err != nil {
		return nil, fmt.Errorf("no se pudo abrir el archivo: %w", err)
	}
	defer archivo.Close()

	reader := csv.NewReader(archivo)

	if _, err := reader.Read(); err != nil {
		return nil, fmt.Errorf("error al leer encabezados: %w", err)
	}

	registros, err := reader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("error al leer CSV: %w", err)
	}

	transacciones := make([]Transaccion, 0, len(registros))
	for _, record := range registros {
		if len(record) < 3 {
			log.Printf("Registro inválido: %v", record)
			continue
		}
		monto, err := strconv.ParseFloat(record[2], 64)
		if err != nil {
			log.Printf("Error al convertir monto: %v", err)
			continue
		}
		transacciones = append(transacciones, Transaccion{
			ID:    record[0],
			Tipo:  record[1],
			Monto: monto,
		})
	}

	return transacciones, nil
}

func generarReporte(transacciones []Transaccion, workers int) {
	type resultadoParcial struct {
		balanceParcial        float64
		conteoCredito         int
		conteoDebito          int
		transaccionMayorMonto Transaccion
	}

	chunkSize := (len(transacciones) + workers - 1) / workers
	resultCh := make(chan resultadoParcial, workers)
	var wg sync.WaitGroup

	for i := 0; i < len(transacciones); i += chunkSize {
		end := i + chunkSize
		if end > len(transacciones) {
			end = len(transacciones)
		}
		wg.Add(1)
		go func(chunk []Transaccion) {
			defer wg.Done()
			var parcial resultadoParcial

			for _, t := range chunk {
				switch strings.ToLower(t.Tipo) {
				case "crédito", "credito":
					parcial.balanceParcial += t.Monto
					parcial.conteoCredito++
				case "débito", "debito":
					parcial.balanceParcial -= t.Monto
					parcial.conteoDebito++
				}
				if t.Monto > parcial.transaccionMayorMonto.Monto {
					parcial.transaccionMayorMonto = t
				}
			}
			resultCh <- parcial
		}(transacciones[i:end])
	}

	go func() {
		wg.Wait()
		close(resultCh)
	}()

	var balanceFinal float64
	var conteoCredito, conteoDebito int
	var transaccionMayorMonto Transaccion

	for parcial := range resultCh {
		balanceFinal += parcial.balanceParcial
		conteoCredito += parcial.conteoCredito
		conteoDebito += parcial.conteoDebito

		if parcial.transaccionMayorMonto.Monto > transaccionMayorMonto.Monto {
			transaccionMayorMonto = parcial.transaccionMayorMonto
		}
	}

	fmt.Println("Reporte de Transacciones (Paralelizado con", workers, "workers)")
	fmt.Println("------------------------------------------------")
	fmt.Printf("Balance Final: %.2f\n", balanceFinal)
	fmt.Printf("Transacción de Mayor Monto: ID %s - %.2f\n", transaccionMayorMonto.ID, transaccionMayorMonto.Monto)
	fmt.Printf("Conteo de Transacciones: Crédito: %d Débito: %d\n", conteoCredito, conteoDebito)
}

func testWorkers(transacciones []Transaccion) {
	workerCounts := []int{1, 4, 6, 8} // Workers para probar

	for _, workers := range workerCounts {
		inicio := time.Now()
		generarReporte(transacciones, workers)
		duracion := time.Since(inicio)
		fmt.Printf("⏱️ Tiempo con %d workers: %s\n", workers, duracion)
	}
}
