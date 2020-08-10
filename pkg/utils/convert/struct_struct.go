/**
 *
 * @author yinjk
 * @create 2019-06-21 18:38
 */
package convert

import (
	"github.com/yinjk/go-utils/pkg/utils/collection/collections"
	"reflect"
)

//CopyProperties 将source结构体所有字段的值拷贝到target结构中
func CopyProperties(source, target interface{}, ignore ...string) {
	var (
		targetTypeMap  = make(map[string]reflect.StructField)
		targetValueMap = make(map[string]reflect.Value)
		sourceType     = reflect.TypeOf(source)
		sourceValue    = reflect.ValueOf(source)
		targetType     = reflect.TypeOf(target).Elem()
		targetValue    = reflect.ValueOf(target)
	)

	if targetValue.IsNil() {
		targetValue = reflect.New(targetType)
	}
	if sourceValue.Kind() == reflect.Ptr { //如果source是指针，取其值递归
		//fmt.Println("ok")
		CopyProperties(sourceValue.Elem().Interface(), target, ignore...)
		return
	}
	targetValue = targetValue.Elem()
	//将target的所有字段缓存到targetFieldMap中
	for i := 0; i < targetType.NumField(); i++ {
		fieldName := targetType.Field(i).Name
		targetTypeMap[fieldName] = targetType.Field(i)
		targetValueMap[fieldName] = targetValue.Field(i)
	}
	//遍历source的所有字段
	for i := 0; i < sourceType.NumField(); i++ {
		var (
			targetType reflect.StructField
			ok         bool
		)
		sourceType := sourceType.Field(i)
		if targetType, ok = targetTypeMap[sourceType.Name]; !ok { //target中没有这个字段
			continue
		}
		if sourceType.Type != targetType.Type { //类型不同的两个字段忽略
			continue
		}
		if collections.IsStringIn(sourceType.Name, ignore...) { //显示指定忽略的，忽略掉
			continue
		}
		//fmt.Println(reflect.ValueOf(sourceValue.Field(i).Interface()))
		targetValue.FieldByName(sourceType.Name).Set(reflect.ValueOf(sourceValue.Field(i).Interface()))
	}
}
