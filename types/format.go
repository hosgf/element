package types

func FormatCpu(data int64, unit string) int64 {
	formatData := int64(0)
	switch unit {
	case "":
		formatData = data
	case "m":
		formatData = data
	case "μ":
		formatData = data / 1000
	case "n":
		formatData = data / 1000 / 1000
	}
	return formatData
}

func FormatMemory(data int64, unit string) int64 {
	formatData := int64(0)
	switch unit {
	case "":
		formatData = data
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

func FormatDataSize(data int64, unit string) int64 {
	formatData := int64(0)
	switch unit {
	case "":
		formatData = data
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
