package format

func Memory(format string, data int64) int64 {
	formatData := int64(0)
	switch format {
	case "Ki":
		formatData = data / 1024
	case "Mi":
		formatData = data
	case "Gi":
		formatData = data * 1024
	case "Ti":
		formatData = data * 1024 * 1024
	}
	return formatData
}

func DataSize(format string, data int64) int64 {
	formatData := int64(0)
	switch format {
	case "Ki":
		formatData = data
	case "Mi":
		formatData = data * 1024
	case "Gi":
		formatData = data * 1024 * 1024
	case "Ti":
		formatData = data * 1024 * 1024 * 1024
	}
	return formatData
}

func Cpu(format string, data int64) int64 {
	formatData := int64(0)
	switch format {
	case "m":
		formatData = data
	case "μ":
		formatData = data / 1000
	case "n":
		formatData = data / 1000 / 1000
	}
	return formatData
}
