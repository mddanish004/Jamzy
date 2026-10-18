package textlen

func UTF16(value string) int {
	units := 0
	for _, char := range value {
		if char > 0xFFFF {
			units += 2
			continue
		}
		units++
	}
	return units
}
