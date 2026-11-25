package stringutils

// Reverse возвращает строку, символы которой идут в обратном порядке.
// Использует []rune, чтобы корректно работать с Unicode (кириллица, эмодзи).
func Reverse(s string) string {
	runes := []rune(s)
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}
	return string(runes)
}
