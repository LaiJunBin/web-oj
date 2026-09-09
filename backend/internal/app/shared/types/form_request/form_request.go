package formrequest

import (
	"encoding/json"
	"errors"
	fileheader "judge-backend/internal/app/shared/types/file_header"
	"mime/multipart"
	"reflect"
	"strconv"
	"strings"

	"github.com/go-playground/validator/v10"
)

var (
	ErrorUnsupportedType = errors.New("unsupported type")
)

type FormRequest struct {
	reader *multipart.Reader
}

func New(reader *multipart.Reader) *FormRequest {
	return &FormRequest{
		reader: reader,
	}
}

func (p *FormRequest) Bind(payload interface{}) error {
	validate := validator.New(validator.WithRequiredStructEnabled())

	form, err := p.reader.ReadForm(0)
	if err != nil {
		return err
	}
	// ReadForm(0) spools every file part to a temp file; without RemoveAll those
	// accumulate in the OS temp dir under sustained uploads until it fills.
	// fileheader.New copies each file's bytes into an in-memory buffer during
	// writeFile below, so the temp files are unneeded once Bind returns.
	defer form.RemoveAll()

	tagFields := make(map[string]string)
	for i := 0; i < reflect.TypeOf(payload).Elem().NumField(); i++ {
		field := reflect.TypeOf(payload).Elem().Field(i)
		tag := field.Tag.Get("form")
		tagFields[tag] = field.Name
	}

	for key, value := range form.Value {
		field, ok := tagFields[key]
		if !ok {
			continue
		}
		if err := writeValue(payload, field, value); err != nil {
			return err
		}
	}

	for key, value := range form.File {
		field, ok := tagFields[key]
		if !ok {
			continue
		}

		if err := writeFile(payload, field, value); err != nil {
			return err
		}
	}

	return validate.Struct(payload)
}

func writeValue(payload interface{}, key string, value interface{}) error {
	field, err := getField(payload, key)
	if err != nil {
		return err
	}

	switch field.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		val := value.([]string)[0]
		intVal, err := strconv.Atoi(val)
		if err != nil {
			return err
		}

		field.SetInt(int64(intVal))
	case reflect.Float32, reflect.Float64:
		val := value.([]string)[0]
		floatVal, err := strconv.ParseFloat(val, 64)
		if err != nil {
			return err
		}

		field.SetFloat(floatVal)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		val := value.([]string)[0]
		uintVal, err := strconv.ParseUint(val, 10, 64)
		if err != nil {
			return err
		}

		field.SetUint(uintVal)
	case reflect.Bool:
		val := value.([]string)[0]
		boolVal, err := strconv.ParseBool(val)

		if err != nil {
			return err
		}

		field.SetBool(boolVal)
	case reflect.String:
		field.SetString(value.([]string)[0])
	case reflect.Slice:
		strVals := strings.Split(strings.Trim(value.([]string)[0], "[]"), ",")
		for i := 0; i < len(strVals); i++ {
			if strVals[i] == "" {
				strVals = append(strVals[:i], strVals[i+1:]...)
				i--
			}
		}
		switch field.Type().Elem().Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			switch field.Type().Elem().Kind() {
			case reflect.Int:

				intVal := make([]int, len(strVals))
				for i, val := range strVals {
					intVal[i], err = strconv.Atoi(val)
					if err != nil {
						return err
					}
				}

				field.Set(reflect.ValueOf(intVal))
			case reflect.Int32:
				int32Val := make([]int32, len(strVals))
				for i, val := range strVals {
					ival, err := strconv.Atoi(val)
					if err != nil {
						return err
					}
					int32Val[i] = int32(ival)
				}

				field.Set(reflect.ValueOf(int32Val))
			default:
				return ErrorUnsupportedType
			}
		case reflect.Float32, reflect.Float64:
			floatVal := make([]float64, len(strVals))
			for i, val := range strVals {
				floatVal[i], err = strconv.ParseFloat(val, 64)
				if err != nil {
					return err
				}
			}

			field.Set(reflect.ValueOf(floatVal))
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			uintVal := make([]uint64, len(strVals))
			for i, val := range strVals {
				uintVal[i], err = strconv.ParseUint(val, 10, 64)
				if err != nil {
					return err
				}
			}

			field.Set(reflect.ValueOf(uintVal))
		case reflect.Bool:
			boolVal := make([]bool, len(strVals))
			for i, val := range strVals {
				boolVal[i], err = strconv.ParseBool(val)
				if err != nil {
					return err
				}
			}

			field.Set(reflect.ValueOf(boolVal))
		case reflect.String:
			field.Set(reflect.ValueOf(strVals))
		default:
			return ErrorUnsupportedType
		}
	case reflect.Map:
		mapStr := value.([]string)[0]
		mapValue := make(map[string]interface{})
		if err := json.Unmarshal([]byte(mapStr), &mapValue); err != nil {
			return err
		}

		switch field.Type().Elem().Kind() {
		case reflect.Int:
			m := make(map[string]int)
			for k, v := range mapValue {
				intVal, err := strconv.Atoi(v.(string))
				if err != nil {
					return err
				}
				m[k] = intVal
			}
			field.Set(reflect.ValueOf(m))
		case reflect.Int32:
			m := make(map[string]int32)
			for k, v := range mapValue {
				intVal, err := strconv.Atoi(v.(string))
				if err != nil {
					return err
				}
				m[k] = int32(intVal)
			}
			field.Set(reflect.ValueOf(m))
		case reflect.String:
			m := make(map[string]string)
			for k, v := range mapValue {
				m[k] = v.(string)
			}
			field.Set(reflect.ValueOf(m))
		default:
			return ErrorUnsupportedType
		}
	case reflect.Ptr:
		switch field.Type().Elem().Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			val := value.([]string)[0]
			intVal, err := strconv.Atoi(val)
			if err != nil {
				return err
			}

			switch field.Type().Elem().Kind() {
			case reflect.Int:
				field.Set(reflect.ValueOf(&intVal))
			case reflect.Int8:
				int8Val := int8(intVal)
				field.Set(reflect.ValueOf(&int8Val))
			case reflect.Int16:
				int16Val := int16(intVal)
				field.Set(reflect.ValueOf(&int16Val))
			case reflect.Int32:
				int32Val := int32(intVal)
				field.Set(reflect.ValueOf(&int32Val))
			case reflect.Int64:
				int64Val := int64(intVal)
				field.Set(reflect.ValueOf(&int64Val))
			default:
				return ErrorUnsupportedType
			}
		case reflect.Float32, reflect.Float64:
			val := value.([]string)[0]
			floatVal, err := strconv.ParseFloat(val, 64)
			if err != nil {
				return err
			}

			field.Set(reflect.ValueOf(&floatVal))
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			val := value.([]string)[0]
			uintVal, err := strconv.ParseUint(val, 10, 64)
			if err != nil {
				return err
			}

			field.Set(reflect.ValueOf(&uintVal))
		case reflect.Bool:
			val := value.([]string)[0]
			boolVal, err := strconv.ParseBool(val)

			if err != nil {
				return err
			}

			field.Set(reflect.ValueOf(&boolVal))
		case reflect.String:
			field.Set(reflect.ValueOf(&value.([]string)[0]))
		case reflect.Slice:
			field.Set(reflect.ValueOf(&value))
		case reflect.Map:
			mapStr := value.([]string)[0]
			mapValue := make(map[string]interface{})
			if err := json.Unmarshal([]byte(mapStr), &mapValue); err != nil {
				return err
			}

			field.Set(reflect.ValueOf(&mapValue))
		default:
			return ErrorUnsupportedType
		}
	default:
		return ErrorUnsupportedType
	}

	return nil
}

func writeFile(payload interface{}, key string, value []*multipart.FileHeader) error {
	field, err := getField(payload, key)
	if err != nil {
		return err
	}

	switch field.Kind() {
	case reflect.Ptr:
		v, err := fileheader.New(value[0])
		if err != nil {
			return err
		}
		field.Set(reflect.ValueOf(v))
	case reflect.Slice:
		v := make([]*fileheader.FileHeader, len(value))
		for i, file := range value {
			f, err := fileheader.New(file)
			if err != nil {
				return err
			}
			v[i] = f
		}
		field.Set(reflect.ValueOf(v))
	default:
		return ErrorUnsupportedType
	}

	return nil
}

func getField(payload interface{}, key string) (reflect.Value, error) {
	elem := reflect.ValueOf(payload).Elem()

	field := elem.FieldByName(key)
	if !field.IsValid() {
		return reflect.Value{}, errors.New("field not found")
	}

	if !field.CanSet() {
		return reflect.Value{}, errors.New("field can not set")
	}

	return field, nil
}
