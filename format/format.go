package format

func Memory(data int64, unit string) int64 {
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

func MemoryOrDefault(data int64, unit string, dataDefault int64) int64 {
	if data > 0 {
		return Memory(data, unit)
	}
	return dataDefault
}

func DataSize(data int64, unit string) int64 {
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

func DataSizeOrDefault(data int64, unit string, dataDefault int64) int64 {
	if data > 0 {
		return DataSize(data, unit)
	}
	return dataDefault
}

func Cpu(data int64, unit string) int64 {
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

func CpuOrDefault(data int64, unit string, dataDefault int64) int64 {
	if data > 0 {
		return Cpu(data, unit)
	}
	return dataDefault
}
