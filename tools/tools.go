package tools

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gofrs/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

func init() {
	// seed random number generator
	rand.Seed(time.Now().UnixNano())
}

func GetTimeUnix() int64 {
	return time.Now().UTC().Unix()
}
func GetTimeUnixMilli() int64 {
	return time.Now().UTC().UnixNano() / int64(time.Millisecond)
}
func GetTimeSql() string {
	// 2006-01-02 15:04:05 has to be used to recognize format!
	return time.Now().UTC().Format("2006-01-02 15:04:05")
}
func GetTimeFromSql(sqlTime string) (time.Time, error) {
	t, err := time.Parse("2006-01-02 15:04:05", sqlTime)
	if err != nil {
		return t, err
	}
	return t, nil
}

func Substring(s string, start, end int) string {
	ctr, index0 := 0, 0
	for index1 := range s {
		if ctr == start {
			index0 = index1
		}
		if ctr == end {
			return s[index0:index1]
		}
		ctr++
	}
	return s[index0:]
}

func CheckCreateDir(dir string) error {
	exists, err := Exists(dir)

	if err != nil {
		return err
	}

	if !exists {
		if err := os.MkdirAll(dir, os.FileMode(0770)); err != nil {
			return err
		}
	}
	return nil
}
func CheckCreateFile(file string, templateFile string) error {
	exists, err := Exists(file)

	if err != nil {
		return err
	}

	if !exists {
		if err := FileCopy(templateFile, file, false); err != nil {
			return err
		}
	}
	return nil
}

func GetFileContents(filePath string, removeUtf8Bom bool) ([]byte, error) {

	output, err := ioutil.ReadFile(filePath)
	if err != nil {
		return []byte("{}"), err
	}
	if removeUtf8Bom {
		output = RemoveUtf8Bom(output)
	}
	return output, nil
}

func RemoveUtf8Bom(input []byte) []byte {
	return bytes.TrimPrefix(input, []byte("\xEF\xBB\xBF"))
}

func StringListToUInt64Array(input string) ([]uint64, error) {
	var output []uint64 = make([]uint64, 0)

	if len([]rune(input)) == 0 {
		return output, nil
	}

	list := strings.Split(input, ",")

	for _, value := range list {
		uint64Value, err := strconv.ParseUint(value, 10, 64)
		if err != nil {
			return output, err
		}
		output = append(output, uint64Value)
	}
	return output, nil
}

func StringInSlice(needle string, haystack []string) bool {
	for _, value := range haystack {
		if value == needle {
			return true
		}
	}
	return false
}

func IntInSlice(needle int, haystack []int) bool {
	for _, value := range haystack {
		if value == needle {
			return true
		}
	}
	return false
}

func Int64InSlice(needle int64, haystack []int64) bool {
	for _, value := range haystack {
		if value == needle {
			return true
		}
	}
	return false
}

func Uint64InSlice(needle uint64, haystack []uint64) bool {
	for _, value := range haystack {
		if value == needle {
			return true
		}
	}
	return false
}

func UuidInSlice(needle uuid.UUID, haystack []uuid.UUID) bool {
	for _, value := range haystack {
		if value == needle {
			return true
		}
	}
	return false
}

func UuidStringToNullUuid(input string) pgtype.UUID {
	id, err := uuid.FromString(input)
	return pgtype.UUID{
		Bytes: id,
		Valid: err == nil,
	}
}

func PgxNumericToString(value pgtype.Numeric) string {
	s := value.Int.String()
	l := len(s)
	e := int(value.Exp)

	// zero exponent, as in 12 (int=12, len=2, exp=0)
	if e == 0 {
		return s
	}

	// positive exponent, as in 2500 (int=25, len=2, exp=2)
	if e > 0 {
		return fmt.Sprintf("%s%s", s, strings.Repeat("0", e))
	}

	// negative exponents
	// equals out length, as in 0.12 (int=12, len=2, exp=-2)
	if l+e == 0 {
		return fmt.Sprintf("0.%s", s)
	}

	// below zero, as in 0.012 (int=12, len=2, exp=-3)
	if l+e < 0 {
		return fmt.Sprintf("0.%s%s", strings.Repeat("0", (l+e)-((l+e)*2)), s)
	}

	// above zero, as in 11.1 (int=111, len=3, exp=-1)
	return fmt.Sprintf("%s.%s", s[0:l+e], s[l+e:])
}

func RandStringRunes(n int) string {
	var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")

	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}
