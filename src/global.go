package register

import (
	"log"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
)

type Global struct {
	logger *log.Logger
}
var G *Global

type Participant struct {
	RegisteredTime string `json:"time"`
	Name string `json:"name"`
	Email string `json:"email"`
	Mobile string `json:"mobile"`
	Org string `json:"org"`
	Place string `json:"place"`
	QrCode []byte `json:"qrcode"`
}

func StructToMap(v any) map[string]string {
	vmap := make(map[string]string)
	valueof := reflect.ValueOf(v)
	if valueof.Kind() == reflect.Ptr {
		valueof = valueof.Elem()
	}
	typeof := valueof.Type()
	for index := 0; index < valueof.NumField(); index++ {
		typef := typeof.Field(index)
		valuef := valueof.Field(index)
		switch valuef.Kind() {
		case reflect.Int64: vmap[typef.Name] = strconv.FormatInt(valuef.Int(), 10)
		case reflect.Uint64: vmap[typef.Name] = strconv.FormatUint(valuef.Uint(), 10)
		case reflect.Bool: vmap[typef.Name] = strconv.FormatBool(valuef.Bool())
		case reflect.String: vmap[typef.Name] = valuef.String()
		}
	}
	return vmap
}

func StructSetFromMap(v any, m map[string]any) {
	valueof := reflect.ValueOf(v)
	for k0, v0 := range m {
		valuef := reflect.Indirect(valueof).FieldByName(k0)
		if valuef.IsValid() == false {
			G.logger.Println("Invalid value for key ", k0)
			continue
		}
		switch valuef.Kind() {
		case reflect.Int64: valuef.SetInt(int64(v0.(float64)))
		case reflect.Uint64: valuef.SetUint(uint64(v0.(float64)))
		case reflect.Bool: valuef.SetBool(v0.(bool))
		case reflect.String: valuef.SetString(v0.(string))
		}
	}
}

func GetStaticPath() (string, error) {
	realpath, err := filepath.EvalSymlinks("/proc/self/exe")
	if err != nil {
		G.logger.Println(err)
		return "", err
	}
	staticpath := filepath.Clean(filepath.Dir(realpath) + "/../../static")
	_, err = os.Stat(staticpath)
	if err != nil {
		G.logger.Println(err)
		return "", err
	}
	return staticpath, nil
}

func init() {
	if G != nil {
		return
	}
	G = &Global{}
	G.logger = log.Default()
	G.logger.SetFlags(log.Ldate |log.Ltime | log.Lmicroseconds | log.Lshortfile)
}

