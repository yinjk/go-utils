package convert

import (
	"github.com/ghodss/yaml"
	"reflect"
	"strings"
)

/**
 * 结构体转map
 * @param : obj 要转换的结构体
 * @return: 转换后的map
 * @author: yinjk
 * @time  : 2019/2/12 19:29
 */
func StructToMap(obj interface{}) map[string]interface{} {
	t := reflect.TypeOf(obj)
	v := reflect.ValueOf(obj)
	var data = make(map[string]interface{})
	for i := 0; i < t.NumField(); i++ {
		data[t.Field(i).Name] = v.Field(i).Interface()
	}
	return data
}

// 该实现为一个简陋版本，只能对基础类型的字段转换，高级版后续再添加
// 相同的类型、名称的属性可以通过反射转换，其他字段忽略
func MapToStruct(maps map[string]interface{}, obj interface{}) {
	v := reflect.ValueOf(obj).Elem()
	for i := 0; i < v.NumField(); i++ {
		mapVal := maps[v.Type().Field(i).Name]
		field := v.Field(i)
		switch mapVal.(type) {
		case string:
			if field.Type().Name() == "string" {
				field.SetString(mapVal.(string))
			}
		case int:
			if field.Type().Name() == "int" {
				field.SetInt(int64(mapVal.(int)))
			}
		case int32:
			if field.Type().Name() == "int32" {
				field.SetInt(int64(mapVal.(int32)))
			}
		case int64:
			if field.Type().Name() == "int64" {
				field.SetInt(int64(mapVal.(int64)))
			}
		case float64:
			if field.Type().Name() == "float64" {
				field.SetFloat(mapVal.(float64))
			}
		case float32:
			if field.Type().Name() == "float32" {
				field.SetFloat(float64(mapVal.(float32)))
			}
		case bool:
			if field.Type().Name() == "bool" {
				field.SetBool(mapVal.(bool))
			}
			field.SetBool(mapVal.(bool))
		case []byte:
			if field.Type().Name() == "[]byte" {
				field.SetBytes(mapVal.([]byte))
			}
		case uint64:
			if field.Type().Name() == "uint64" {
				field.SetUint(mapVal.(uint64))
			}
		}
	}
}

/**
 * Take a {@code String} that is a delimited list and convert it into
 * a {@code String} array.
 * <p>A single {@code delimiter} may consist of more than one character,
 * but it will still be considered as a single delimiter string, rather
 * than as bunch of potential delimiter characters, in contrast to
 * {@link #tokenizeToStringArray}.
 * @param str the input {@code String} (potentially {@code null} or empty)
 * @param delimiter the delimiter between elements (this is a single delimiter,
 * rather than a bunch individual delimiter characters)
 * @return an array of the tokens in the list
 * @see #tokenizeToStringArray
 */
func DelimitedListToStringArray(str string, delimiter string) []string {
	return strings.Split(str, delimiter)
}

/**
 * Convert a comma delimited list (e.g., a row from a CSV file) into an
 * array of strings.
 * @param str the input {@code String} (potentially {@code null} or empty)
 * @return an array of strings, or the empty array in case of empty input
 */
func CommaDelimitedListToStringArray(str string) []string {
	return DelimitedListToStringArray(str, ",")
}

func CollectionToDelimitedString(param []string) (s string) {
	for _, v := range param {
		s = s + v + ","
	}
	if s != "" {
		s = s[:len(s)-1]
	}
	return
}

func RemoveFLComma(source string) string {
	if source == "" {
		return source
	}
	first, last := -1, 0
	for index, v := range source {
		if v != ',' {
			first = index
			break
		}
	}
	if first == -1 {
		//全是逗号，直接返回""
		return ""
	}
	for index := len(source) - 1; index >= 0; index-- {
		if source[index] != ',' {
			last = index
			break
		}
	}
	return source[first : last+1]
}

//YamlToJson Yaml转json
func YamlToJson(content string) (result string, err error) {
	jsonBytes, err := yaml.YAMLToJSON([]byte(content))
	if err != nil {
		return "", err
	}
	return string(jsonBytes), nil
}

//JsonToYaml Yaml转Json
func JsonToYaml(content string) (result string, err error) {
	toYAML, err := yaml.JSONToYAML([]byte(content))
	if err != nil {
		return "", err
	}
	return string(toYAML), nil
}
