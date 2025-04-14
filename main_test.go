package main

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

// Test_LeerTransaccionesDesdeCSV verifica la correcta lectura de transacciones desde un archivo CSV
func Test_LeerTransaccionesDesdeCSV(t *testing.T) {
	// Crear un archivo CSV temporal para las pruebas
	contenido := `id,tipo,monto
1,Crédito,100.00
2,Débito,50.00
3,Crédito,200.00
`
	archivoTemp := crearArchivoTemporal(t, "test_transacciones*.csv", contenido)
	defer os.Remove(archivoTemp) // Limpiar después de la prueba

	// Resultado esperado
	esperado := []Transaccion{
		{ID: "1", Tipo: "Crédito", Monto: 100.00},
		{ID: "2", Tipo: "Débito", Monto: 50.00},
		{ID: "3", Tipo: "Crédito", Monto: 200.00},
	}

	// Ejecutar la función a probar
	transacciones, err := leerTransaccionesDesdeCSV(archivoTemp)
	if err != nil {
		t.Fatalf("Error inesperado al leer el archivo CSV: %v", err)
	}

	// Verificar resultados
	if !reflect.DeepEqual(transacciones, esperado) {
		t.Errorf("Las transacciones leídas no coinciden con las esperadas.\nObtenido: %+v\nEsperado: %+v", transacciones, esperado)
	}
}

// Test_LeerTransaccionesDesdeCSVArchivoNoExistente verifica el manejo de errores cuando el archivo no existe
func Test_LeerTransaccionesDesdeCSVArchivoNoExistente(t *testing.T) {
	_, err := leerTransaccionesDesdeCSV("archivo_inexistente.csv")
	if err == nil {
		t.Error("Se esperaba un error al intentar leer un archivo inexistente, pero no se recibió ninguno")
	}
}

// Test_LeerTransaccionesDesdeCSVFormatoInvalido verifica el manejo de errores cuando el formato del CSV es inválido
func Test_LeerTransaccionesDesdeCSVFormatoInvalido(t *testing.T) {
	// Crear un archivo CSV temporal con un formato inválido
	contenido := `id,tipo,monto
1,Crédito,no_es_un_numero
`
	archivoTemp := crearArchivoTemporal(t, "test_formato_invalido*.csv", contenido)
	defer os.Remove(archivoTemp)

	_, err := leerTransaccionesDesdeCSV(archivoTemp)
	if err == nil {
		t.Error("Se esperaba un error al intentar procesar un CSV con formato inválido, pero no se recibió ninguno")
	}
}

// Test_LeerTransaccionesDesdeCSVVacio verifica el comportamiento cuando el CSV solo tiene encabezados
func Test_LeerTransaccionesDesdeCSVVacio(t *testing.T) {
	// Crear un archivo CSV temporal solo con encabezados
	contenido := "id,tipo,monto"
	archivoTemp := crearArchivoTemporal(t, "test_vacio*.csv", contenido)
	defer os.Remove(archivoTemp)

	transacciones, err := leerTransaccionesDesdeCSV(archivoTemp)
	if err != nil {
		t.Fatalf("Error inesperado al leer el archivo CSV vacío: %v", err)
	}

	if len(transacciones) != 0 {
		t.Errorf("Se esperaba una lista vacía de transacciones, pero se obtuvieron %d transacciones", len(transacciones))
	}
}

// Test_GenerarReporte verifica que el reporte se genera correctamente
func Test_GenerarReporte(t *testing.T) {
	// Configurar transacciones de prueba
	transacciones := []Transaccion{
		{ID: "1", Tipo: "Crédito", Monto: 100.00},
		{ID: "2", Tipo: "Débito", Monto: 50.00},
		{ID: "3", Tipo: "Crédito", Monto: 200.00},
		{ID: "4", Tipo: "Débito", Monto: 75.00},
		{ID: "5", Tipo: "Crédito", Monto: 150.00},
	}

	// Capturar la salida estándar
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Ejecutar la función
	generarReporte(transacciones)

	// Restaurar stdout y leer la salida capturada
	w.Close()
	os.Stdout = oldStdout
	var buf bytes.Buffer
	io.Copy(&buf, r)
	salida := buf.String()

	// Verificar que la salida contiene la información esperada
	if !strings.Contains(salida, "Balance Final: 325.00") {
		t.Error("El reporte no contiene el balance final correcto")
	}
	if !strings.Contains(salida, "Transacción de Mayor Monto: ID 3 - 200.00") {
		t.Error("El reporte no identifica correctamente la transacción de mayor monto")
	}
	if !strings.Contains(salida, "Conteo de Transacciones: Crédito: 3 Débito: 2") {
		t.Error("El reporte no muestra correctamente el conteo de transacciones")
	}
}

// Test_GenerarReporteVariantesTipo verifica que el reporte maneja correctamente diferentes variantes de tipo
func Test_GenerarReporteVariantesTipo(t *testing.T) {
	// Transacciones con diferentes variantes de escritura para "Crédito" y "Débito"
	transacciones := []Transaccion{
		{ID: "1", Tipo: "Crédito", Monto: 100.00},
		{ID: "2", Tipo: "Debito", Monto: 50.00},   // Sin tilde
		{ID: "3", Tipo: "credito", Monto: 200.00}, // Minúsculas
		{ID: "4", Tipo: "débito", Monto: 75.00},   // Minúsculas con tilde
	}

	// Capturar la salida estándar
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Ejecutar la función
	generarReporte(transacciones)

	// Restaurar stdout y leer la salida capturada
	w.Close()
	os.Stdout = oldStdout
	var buf bytes.Buffer
	io.Copy(&buf, r)
	salida := buf.String()

	// Verificar que la salida contiene la información esperada
	if !strings.Contains(salida, "Balance Final: 175.00") {
		t.Error("El reporte no maneja correctamente las variantes de tipo")
	}
	if !strings.Contains(salida, "Conteo de Transacciones: Crédito: 2 Débito: 2") {
		t.Error("El reporte no cuenta correctamente las transacciones con diferentes variantes de tipo")
	}
}

// Test_EjecutarAplicacionSinArgumentos verifica que la aplicación maneje correctamente la falta de argumentos
func Test_EjecutarAplicacionSinArgumentos(t *testing.T) {
	// Capturar la salida estándar
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Ejecutar la función con argumentos insuficientes
	code := ejecutarAplicacion([]string{"bancli"})

	// Restaurar stdout y leer la salida capturada
	w.Close()
	os.Stdout = oldStdout
	var buf bytes.Buffer
	io.Copy(&buf, r)
	salida := buf.String()

	// Verificar el resultado
	if code != 1 {
		t.Errorf("Se esperaba un código de salida 1, pero se obtuvo %d", code)
	}

	if !strings.Contains(salida, "Uso: go run main.go") {
		t.Error("No se mostró el mensaje de uso cuando no hay suficientes argumentos")
	}
}

// Test_EjecutarAplicacionConArchivoInvalido verifica que la aplicación maneje correctamente un archivo inválido
func Test_EjecutarAplicacionConArchivoInvalido(t *testing.T) {
	// Capturar la salida estándar y de error
	oldStdout := os.Stdout
	oldStderr := os.Stderr
	rOut, wOut, _ := os.Pipe()
	rErr, wErr, _ := os.Pipe()
	os.Stdout = wOut
	os.Stderr = wErr

	// Ejecutar la función con un archivo inexistente
	code := ejecutarAplicacion([]string{"bancli", "archivo_inexistente.csv"})

	// Restaurar stdout y stderr
	wOut.Close()
	wErr.Close()
	os.Stdout = oldStdout
	os.Stderr = oldStderr

	// Leer la salida capturada
	var outBuf, errBuf bytes.Buffer
	io.Copy(&outBuf, rOut)
	io.Copy(&errBuf, rErr)

	// Verificar el resultado
	if code != 1 {
		t.Errorf("Se esperaba un código de salida 1, pero se obtuvo %d", code)
	}
}

// Test_EjecutarAplicacionExitoso verifica que la aplicación funcione correctamente con un archivo válido
func Test_EjecutarAplicacionExitoso(t *testing.T) {
	// Crear un archivo CSV temporal para las pruebas
	contenido := `id,tipo,monto
1,Crédito,100.00
2,Débito,50.00
3,Crédito,200.00
`
	archivoTemp := crearArchivoTemporal(t, "test_transacciones*.csv", contenido)
	defer os.Remove(archivoTemp)

	// Capturar la salida estándar
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Ejecutar la función con un archivo válido
	code := ejecutarAplicacion([]string{"bancli", archivoTemp})

	// Restaurar stdout y leer la salida capturada
	w.Close()
	os.Stdout = oldStdout
	var buf bytes.Buffer
	io.Copy(&buf, r)
	salida := buf.String()

	// Verificar el resultado
	if code != 0 {
		t.Errorf("Se esperaba un código de salida 0, pero se obtuvo %d", code)
	}

	if !strings.Contains(salida, "Balance Final: 250.00") {
		t.Error("El reporte no contiene el balance final correcto")
	}
}

// Función auxiliar para crear archivos temporales para pruebas
func crearArchivoTemporal(t *testing.T, patrón, contenido string) string {
	t.Helper()
	tempDir := t.TempDir()
	tempFile, err := os.CreateTemp(tempDir, patrón)
	if err != nil {
		t.Fatalf("Error al crear archivo temporal: %v", err)
	}

	if _, err := tempFile.WriteString(contenido); err != nil {
		t.Fatalf("Error al escribir en archivo temporal: %v", err)
	}

	if err := tempFile.Close(); err != nil {
		t.Fatalf("Error al cerrar archivo temporal: %v", err)
	}

	return filepath.Join(tempDir, filepath.Base(tempFile.Name()))
}
