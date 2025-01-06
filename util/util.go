package util

import (
	"context"
	rand2 "crypto/rand"
	"errors"
	"hash/crc32"
	"math"
	"math/big"
	"math/rand"
	"net"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/hosgf/element/logger"
)

func FilterDuplicates(datas []string) []string {
	var uniqueDatas []string
	seen := make(map[string]bool)
	for _, str := range datas {
		if len(str) < 1 {
			continue
		}
		if seen[str] {
			continue
		}
		uniqueDatas = append(uniqueDatas, str)
		seen[str] = true
	}
	return uniqueDatas
}

func Any(expr bool, a, b string) string {
	if expr {
		return a
	}
	return b
}

func AnyInt(expr bool, a, b int) int {
	if expr {
		return a
	}
	return b
}

func AnyInt64(expr bool, a, b int64) int64 {
	if expr {
		return a
	}
	return b
}

func GetPath(path string, args ...string) string {
	if args == nil || len(args) < 1 {
		return path
	}
	for _, arg := range args {
		if len(arg) > 0 {
			path += "/" + arg
		}
	}
	return path
}

// HashCode 计算hashcode唯一值
func HashCode(s string) int64 {
	v := int64(crc32.ChecksumIEEE([]byte(s)))
	if v >= 0 {
		return v
	}
	if -v >= 0 {
		return -v
	}
	return -1
}

// RandomNum 随机数
func RandomNum(length int) string {
	numberAttr := [10]int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}
	numberLen := len(numberAttr)
	rand.Seed(time.Now().UnixNano())
	var sb strings.Builder
	for i := 0; i < length; i++ {
		itemInt := numberAttr[rand.Intn(numberLen)]
		sb.WriteString(strconv.Itoa(itemInt))
	}
	randStr := sb.String()
	sb.Reset()
	return randStr
}

// RandomAround 范围随机数
func RandomAround(min, max int64) (int64, error) {
	if min > max {
		return 0, errors.New("the min is greater than max")
	}
	if min < 0 {
		f64Min := math.Abs(float64(min))
		i64Min := int64(f64Min)
		result, _ := rand2.Int(rand2.Reader, big.NewInt(max+1+i64Min))
		return result.Int64() - i64Min, nil
	} else {
		result, _ := rand2.Int(rand2.Reader, big.NewInt(max-min+1))
		return min + result.Int64(), nil
	}
}

func GetRealIp() string {
	adders, err := net.InterfaceAddrs()
	if err != nil {
		return "0.0.0.0"
	}
	for _, addr := range adders {
		ipAddr, ok := addr.(*net.IPNet)
		if !ok {
			continue
		}
		if ipAddr.IP.IsLoopback() {
			continue
		}
		if !ipAddr.IP.IsGlobalUnicast() {
			continue
		}
		return ipAddr.IP.String()
	}
	return "0.0.0.0"
}

func GetOrDefault(str string, def string) string {
	if len(str) < 1 {
		return def
	}
	return str
}

func GetIntOrDefault(str int, def int) int {
	if str < 1 {
		return def
	}
	return str
}

func Addr() (string, error) {
	ifaces, err := net.Interfaces()
	if err != nil {
		return "", err
	}
	for _, i := range ifaces {
		if i.Flags&net.FlagUp == 0 {
			continue
		}
		addrs, err := i.Addrs()
		if err != nil {
			continue
		}
		for _, addr := range addrs {
			ipnet, ok := addr.(*net.IPNet)
			if !ok || ipnet.IP.IsLoopback() {
				continue
			}
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String(), nil
			}
		}
	}
	return "", nil
}

func AppDir() string {
	homePath := GetHomePath()
	if len(homePath) < 1 {
		return homePath
	}
	dir := getCurrentAbPathByExecutable()
	tmpDir, _ := filepath.EvalSymlinks(os.TempDir())
	if strings.Contains(dir, tmpDir) {
		return getCurrentAbPathByCaller()
	}
	return dir
}

func getCurrentAbPathByExecutable() string {
	exePath, err := os.Executable()
	if err != nil {
		logger.Log().Error(context.Background(), err)
	}
	res, _ := filepath.EvalSymlinks(filepath.Dir(exePath))
	return res
}

func getCurrentAbPathByCaller() string {
	var abPath string
	_, filename, _, ok := runtime.Caller(0)
	if ok {
		abPath = path.Dir(filename)
	}
	return abPath
}
