// Пакет mathutils содержит математические утилиты.
package mathutils

// Factorial вычисляет факториал числа n (n!).
// Для n < 0 возвращает 0, так как факториал не определён.
// Для n = 0 и n = 1 возвращает 1 (по определению 0! = 1).
func Factorial(n int) int {
	if n < 0 {
		return 0
	}
	result := 1
	for i := 2; i <= n; i++ {
		result *= i
	}
	return result
}
