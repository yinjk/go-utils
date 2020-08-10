/**
 *
 * @author yinjk
 * @create 2019-05-21 11:17
 */
package mysql

import (
	"database/sql"
	"errors"
	"reflect"
	"strings"
	"sync"
	"time"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	"github.com/prometheus/common/log"

	"github.com/yinjk/go-utils/pkg/utils/collection/maps"
)

var (
	tables = maps.NewSegmentHashMap()
	once   = sync.Once{}
	db     *BaseOrm
)

// Config mysql config.
type Config struct {
	DataBase    string //data base name
	DSN         string // data source name.
	Active      int    // pool
	Idle        int    // pool
	IdleTimeout int    // connect max life time.
	LogMode     bool   // is print log
}

// NewMySQL new db and retry connection when has error.
func NewMySQL(c *Config) (db *BaseOrm) {
	createDataBase(c) //如果数据库不存在，则尝试创建数据库
	orm, err := gorm.Open("mysql", c.DSN)
	if err != nil {
		log.Errorf("db dsn(%v) error: %v", c.DSN, err)
		panic(err)
	}
	db = &BaseOrm{orm}
	log.Info("begin set max idle connection ", c.Idle)
	db.DB.DB().SetMaxIdleConns(c.Idle)
	log.Info("begin set max open connection ", c.Active)
	db.DB.DB().SetMaxOpenConns(c.Active)
	idleTimeout := c.IdleTimeout
	//添加判断防止设置超时时间
	if idleTimeout > 10 || idleTimeout <= 0 {
		idleTimeout = 10
		log.Info("use default idle time 10 min")
	}
	log.Info("begin set connection max life time, ", idleTimeout)
	db.DB.DB().SetConnMaxLifetime(time.Duration(idleTimeout) * time.Minute)
	if err = db.DB.DB().Ping(); err != nil {
		log.Error(err)
		panic(err)
	}
	db.DB.LogMode(c.LogMode)
	return
}

//Get get orm client single instance
func Get(c *Config) *BaseOrm {
	once.Do(func() {
		db = NewMySQL(c)
	})
	return db
}

/**
 * 尝试创建database，如果database不存在的话
 * @param : c 数据库配置
 * @return:
 * @author: yinjk
 * @time  : 2019/6/12 10:03
 */
func createDataBase(c *Config) {
	dsn := strings.Replace(c.DSN, c.DataBase, "mysql", 1) //连接名为mysql的数据库
	db, err := gorm.Open("mysql", dsn)
	if err != nil {
		log.Errorf("db dsn(%v) error: %v", dsn, err)
		panic(err)
	}
	defer func() { _ = db.Close() }() //数据库创建之后直接关闭该连接
	if err = db.Exec("create database if not exists " + c.DataBase + " Character Set UTF8;").Error; err != nil {
		log.Errorf("create database %v error(%v)", c.DataBase, err)
		panic(err)
	}
}

type BaseOrm struct {
	*gorm.DB
}

type Pagination struct {
	Page       int `json:"page"`
	PageSize   int `json:"pageSize"`
	TotalCount int `json:"totalCount"`
}

type PageData struct {
	Pagination
	Data interface{} `json:"data"`
}

/**
 * 开启mysql事物
 * @param :
 * @return: tx 开启的事物
 * @author: yinjk
 * @time  : 2019/6/11 17:06
 */
func (bo BaseOrm) Begin() (tx *Tx) {
	db := bo.DB.Begin()
	return &Tx{db}
}

//Commit 请勿执行该方法，已废弃，若要提交事物，请使用tx.Commit()
//Deprecated
func (bo BaseOrm) Commit() (tx *gorm.DB) {
	panic(gorm.ErrInvalidTransaction)
}

//Rollback 请勿执行该方法，已废弃，若要回滚事物，请使用tx.Rollback()
//Deprecated
func (bo BaseOrm) Rollback() (tx *gorm.DB) {
	panic(gorm.ErrInvalidTransaction)
}

/**
 * 创建表
 * @param : beans 要创建的表的实体
 * @return:
 * @author: yinjk
 * @time  : 2019/6/11 17:10
 */
func (bo BaseOrm) CreateTable(beans interface{}) {
	beansType := reflect.TypeOf(beans)
	if !tables.PutIfAbsent(beansType, true) { //返回false表示已经map中已经创建过该类型的表了，这里直接返回
		return
	}
	//put成功，表示第一次创建，执行创建表语句
	if !bo.HasTable(beans) {
		if tables, ok := beans.(Tabler); ok {
			log.Infof("create table %v", tables.TableName())
		}
		if err := bo.Set("gorm:table_options", "ENGINE=InnoDB DEFAULT CHARSET=utf8").CreateTable(beans).Error; err != nil && !strings.Contains(err.Error(), "already exists") {
			panic(err)
		}
	}
}

//expand method
/**
 * 分页查询
 * @param : result 查询结果
 * @param : deleted 是否查询已逻辑删除的结果
 * @param : pagination 分页器
 * @param : orderBy 排序规则，默认为空
 * @param : query 查询语句
 * @param : args 查询参数
 * @return: page 分页查询结果，包括当前页码，结果总条数等结果
 * @author: yinjk
 * @time  : 2019/6/10 10:46
 */
func (bo BaseOrm) ListWithPage(result interface{}, pagination Pagination, orderBy string, query interface{}, args ...interface{}) (page *PageData, err error) {
	typ := reflect.TypeOf(result).Elem()
	if typ.Kind() != reflect.Slice {
		return nil, errors.New("the result only supports slice type")
	}
	db := bo.DB
	if orderBy != "" {
		db = db.Order(orderBy)
	}
	if query != nil && query != "" {
		db = db.Where(query, args...)
	}
	if err = db.Offset((pagination.Page - 1) * pagination.PageSize).Limit(pagination.PageSize).Find(result).Error; err != nil {
		log.Errorf("BaseOrm.ListWithPage db.Find error(%v):", err)
		return nil, err
	}
	//查询分页数据集的总数
	value := reflect.New(typ.Elem()).Interface()
	if err = db.Model(value).Count(&pagination.TotalCount).Error; err != nil {
		log.Errorf("page query BaseOrm.ListWithPage db.Count error(%v):", err)
		return nil, err
	}
	return &PageData{Pagination: pagination, Data: result}, nil
}

/**
 * 执行sql语句查询
 * @param : query 查询语句
 * @param : args 查询参数
 * @return: rows 查询结果，需要遍历这个rows
 * @author: yinjk
 * @time  : 2019/6/10 10:55
 */
func (bo BaseOrm) Query(query string, args ...interface{}) (rows *sql.Rows, err error) {
	return bo.DB.Raw(query, args...).Rows()
}

type StructFieldMap map[string]reflect.StructField

func (s StructFieldMap) getField(name string) (reflect.StructField, bool) {
	field, ok := s[name]
	if ok {
		return field, ok
	}
	// if not found, try covert mysql field name from snake case to camel case
	name = strings.ReplaceAll(name, "_", "")
	field, ok = s[name]
	if ok {
		return field, ok
	}
	// if still not found, it maybe a join query，try analyze it.
	//if strings.Contains(name, ".") {
	//	strings.Split(name, ".")
	//}
	return field, ok
}

/**
 * 执行sql快速查询,该方法相比Query使用更简单，不需要自己手动去遍历rows了
 * @param : query 查询语句
 * @param : scanner 查询结果映射
 * @param : args 查询参数
 * @return: result 查询结果
 * @author: yinjk
 * @time  : 2019/6/10 11:07
 */
func (bo BaseOrm) FastQuery(query string, value interface{}, args ...interface{}) (err error) {
	var (
		targetTypeMap = make(StructFieldMap)
		targetValue   = reflect.ValueOf(value)
	)
	if targetValue.Kind() != reflect.Ptr {
		return errors.New("FastQuery must accept a ptr type")
	}
	targetValue = reflect.Indirect(targetValue)
	if targetValue.Kind() == reflect.Slice {
		rows, err := bo.Query(query, args...)
		if err != nil {
			return err
		}
		dbTypes, err := rows.ColumnTypes()
		if err != nil {
			return err
		}
		scanValues := make([]interface{}, 0)
		keys := make([]string, 0)
		slice := reflect.MakeSlice(targetValue.Type(), 0, 0)
		switch targetValue.Type().Elem().Kind() {
		case reflect.Map:
			for _, v := range dbTypes {
				if targetValue.Type().Elem().Elem().Kind() == reflect.Interface {
					i := reflect.New(v.ScanType()).Interface()
					scanValues = append(scanValues, i)
				} else {
					i := reflect.New(targetValue.Type().Elem().Elem()).Interface()
					scanValues = append(scanValues, i)
				}
				keys = append(keys, v.Name())
			}
			for rows.Next() {
				refMap := reflect.MakeMap(targetValue.Type().Elem())
				if err := rows.Scan(scanValues...); err != nil {
					return err
				}
				for i, v := range scanValues {
					refMap.SetMapIndex(reflect.ValueOf(keys[i]), reflect.ValueOf(v).Elem())
				}
				slice = reflect.Append(slice, refMap)
			}
		case reflect.Struct:
			for i := 0; i < targetValue.Type().Elem().NumField(); i++ {
				fieldName := targetValue.Type().Field(i).Name
				structField := targetValue.Field(i).Type()
				// If the body contains a nested structure, then need to traverse the fields of the inner structure
				if structField.Kind() == reflect.Struct && structField.Name() == fieldName {
					for j := 0; j < structField.NumField(); j++ {
						targetTypeMap[strings.ToUpper(structField.Field(j).Name)] = structField.Field(j)
					}
					continue
				}
				targetTypeMap[strings.ToUpper(fieldName)] = targetValue.Type().Elem().Field(i)
			}
			for _, v := range dbTypes {
				field, ok := targetTypeMap.getField(strings.ToUpper(v.Name()))
				if !ok { //target中没有这个字段，设置一个虚拟字段用来占位
					var ignoreField interface{}
					scanValues = append(scanValues, &ignoreField)
					keys = append(keys, strings.ToUpper(v.Name()))
					continue
				}
				i := reflect.New(field.Type).Interface()
				//fmt.Println(v.ScanType().Kind() == reflect.Uint32)
				//i := reflect.New(v.ScanType()).Interface()
				scanValues = append(scanValues, i)
				keys = append(keys, strings.ToUpper(v.Name()))
			}
			for rows.Next() {
				row := reflect.New(targetValue.Type().Elem()).Interface()
				refStu := reflect.ValueOf(row).Elem()
				if err = rows.Scan(scanValues...); err != nil {
					return err
				}
				for i, v := range scanValues {
					field, ok := targetTypeMap.getField(keys[i])
					if !ok { //结构体中没有该字段，说明是一个用来占位的虚拟字段，直接跳过
						continue
					}
					// if this field is ptr type
					if refStu.FieldByIndex(field.Index).Kind() == reflect.Ptr {
						refStu.FieldByIndex(field.Index).Set(reflect.ValueOf(&v))
					} else {
						refStu.FieldByIndex(field.Index).Set(reflect.ValueOf(v).Elem())
					}
				}
				slice = reflect.Append(slice, refStu)
			}
		case reflect.Slice:
			for _, v := range dbTypes {
				if targetValue.Type().Elem().Elem().Kind() == reflect.Interface {
					i := reflect.New(v.ScanType()).Interface()
					scanValues = append(scanValues, i)
				} else {
					i := reflect.New(targetValue.Type().Elem().Elem()).Interface()
					scanValues = append(scanValues, i)
				}
			}
			for rows.Next() {
				refSlice := reflect.MakeSlice(targetValue.Type().Elem(), len(scanValues), len(scanValues))
				if err := rows.Scan(scanValues...); err != nil {
					return err
				}
				for i, v := range scanValues {
					refSlice.Index(i).Set(reflect.ValueOf(v).Elem())
				}
				slice = reflect.Append(slice, refSlice)
			}
		default:
			for rows.Next() {
				i := reflect.New(targetValue.Type().Elem()).Interface()
				if err = rows.Scan(i); err != nil {
					return err
				}
				slice = reflect.Append(slice, reflect.ValueOf(i).Elem())
			}
		}
		targetValue.Set(slice)
	} else {
		return errors.New("FastQuery only support a slice type")
	}
	return nil
}

/**
 * 执行sql，不包括返回结果，一般用于增删改
 * @param : sql 查询语句
 * @param : args 查询参数
 * @return: rowsAffected 影响行数
 * @author: yinjk
 * @time  : 2019/6/10 11:11
 */
func (bo BaseOrm) Exec(sql string, args ...interface{}) (rowsAffected int64, err error) {
	exec := bo.DB.Exec(sql, args...)
	return exec.RowsAffected, exec.Error
}

/**
 * 通过id删除
 * @param : id id
 * @return: beans 要删除的对象类型
 * @author: yinjk
 * @time  : 2019/6/10 11:11
 */
func (bo BaseOrm) DeleteById(id interface{}, beans interface{}) error {
	return bo.Delete(beans, "id=?", id).Error
}

/**
 * 创建或更新数据库表记录
 * @param : val 要创建的数据库对象
 * @return:
 * @author: yinjk
 * @time  : 2019/6/10 11:12
 */
func (bo BaseOrm) Save(val interface{}) error {
	return bo.DB.Save(val).Error
}

/**
 * 创建数据库表记录
 * @param : val 要创建的数据库对象
 * @return:
 * @author: yinjk
 * @time  : 2019/6/10 11:12
 */
func (bo BaseOrm) Create(val interface{}) error {
	return bo.DB.Create(val).Error
}

/**
 * 更新数据库表记录，非全量更新，只更新非空值字段（空值：string的空白 bool的false int的0值都是空值）
 * @param : bean 要更新的对象
 * @return:
 * @author: yinjk
 * @time  : 2019/6/10 11:16
 */
func (bo BaseOrm) Update(bean interface{}) error {
	return bo.DB.Model(bean).Update(bean).Error
}

/**
 * 非全量更新，只更新args参数中的字段
 * @param : bean 要更新的表的结构体
 * @return: args 要更新的字段
 * @author: yinjk
 * @time  : 2019/6/10 11:20
 */
func (bo BaseOrm) Updates(bean interface{}, args map[string]interface{}) error {
	return bo.DB.Model(bean).Updates(args).Error
}

/**
 * find in
 * @param : field 查询字段
 * @param : values slice类型参数
 * @param : beans 要查询的实体
 * @return:
 * @author: yinjk
 * @time  : 2019/6/10 11:47
 */
func (bo BaseOrm) FindIn(field string, value interface{}, beans interface{}) error {
	return bo.DB.Where(field+" in(?)", value).Find(beans).Error
}

func (bo BaseOrm) FindById(id interface{}, beans interface{}) error {
	if err := bo.DB.Where("id=?", id).First(beans).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil
		}
		return err
	}
	return nil
}

func (bo BaseOrm) FindOneEq(beans interface{}, fieldName string, value interface{}) (err error) {
	return bo.DB.Where(fieldName+" = ?", value).First(beans).Error
}
func (bo BaseOrm) FindOneCondition(bean interface{}, condition string, params ...interface{}) (err error) {
	return bo.DB.Where(condition, params...).First(bean).Error
}

func (bo BaseOrm) FindCondition(bean interface{}, condition string, params ...interface{}) (err error) {
	return bo.DB.Where(condition, params...).Find(bean).Error
}

func (bo BaseOrm) FindOneLike(fieldName string, value interface{}, beans interface{}) (err error) {
	if err = bo.DB.Where(fieldName+" like ?", value).First(beans).Error; err != nil {
		log.Error(err)
	}
	return
}

/**
 *@describe
 *@param sql 原始sql语句
 *@param params 需要设置的参数
 *@return map数组
 *@author yinchong
 *@create 2019/6/11 17:23
 */
func (bo BaseOrm) ListMap(sql string, params ...interface{}) (data []map[string]interface{}, err error) {
	//defer db.Close()
	rows, err := bo.DB.Raw(sql, params...).Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	cols, _ := rows.Columns()
	for rows.Next() {
		columns := make([]interface{}, len(cols))
		columnPointers := make([]interface{}, len(cols))
		for i := range columns {
			columnPointers[i] = &columns[i]
		}

		if err := rows.Scan(columnPointers...); err != nil {
			return nil, err
		}

		m := make(map[string]interface{})
		for i, colName := range cols {
			val := columnPointers[i].(*interface{})
			switch v := (*val).(type) {
			case nil:
				m[colName] = nil
			case time.Time:
				m[colName] = v.Format("2006-01-02 15:04:05")
			case []byte:
				m[colName] = string(v)
			default:
				m[colName] = v
			}

		}
		data = append(data, m)
	}
	return data, nil
}

func (bo BaseOrm) Count(beans interface{}, query interface{}, args ...interface{}) (count int, err error) {
	err = bo.DB.Model(beans).Where(query, args...).Count(&count).Error
	return
}

func (bo BaseOrm) ListBy(beans interface{}, fields *[]string, query interface{}, args ...interface{}) error {
	if len(*fields) == 0 {
		return bo.DB.Where(query, args...).Find(beans).Error
	} else {
		return bo.DB.Select(*fields).Where(query, args...).Find(beans).Error
	}
}

func (bo BaseOrm) ListEq(fieldName string, value interface{}, beans interface{}) error {
	return bo.DB.Where(fieldName+" = ?", value).Find(beans).Error
}

func (bo BaseOrm) ListNotEq(fieldName string, value interface{}, beans interface{}) error {
	return bo.DB.Where(fieldName+" != ?", value).Find(beans).Error
}

func (bo BaseOrm) ListLike(fieldName string, value interface{}, beans interface{}) error {
	return bo.DB.Where(fieldName+" like ?", value).Find(beans).Error
}

func (bo BaseOrm) ListAll(beans interface{}) error {
	return bo.DB.Find(beans).Error
}

/**
 *@describe   获取单列单条数据
 *@param rowSql  原始sql语句 params 参数
 *@return 单列单条数据记录
 *@author yinchong
 *@create 2019/6/13 19:01
 */
func (bo BaseOrm) FindAloneRecord(rowSql string, params ...interface{}) (value interface{}, err error) {
	var (
		rows *sql.Rows
	)
	if rows, err = bo.Query(rowSql, params...); err != nil {
		log.Error(err)
		return -1, err
	}
	defer rows.Close()
	if rows.Next() {
		if err = rows.Scan(&value); err != nil {
			log.Error(err)
			return nil, err
		}
	}
	return value, err
}

// Tx method

type Tx struct {
	*gorm.DB
}

func (t Tx) ListWithPage(result interface{}, deleted bool, pagination Pagination, orderBy string, query interface{}, args ...interface{}) (page *PageData, err error) {
	typ := reflect.TypeOf(result).Elem()
	if typ.Kind() != reflect.Slice {
		return nil, errors.New("the result only supports slice type")
	}
	var db = t.DB
	if deleted {
		db = t.Unscoped()
	}
	if orderBy != "" {
		db = db.Order(orderBy)
	}
	if query != nil && query != "" {
		db = db.Where(query, args...)
	}
	if err = db.Offset((pagination.Page - 1) * pagination.PageSize).Limit(pagination.PageSize).Find(result).Error; err != nil {
		log.Errorf("BaseOrm.ListWithPage db.Find error(%v):", err)
		return nil, err
	}
	//查询分页数据集的总数
	value := reflect.New(typ.Elem()).Interface()
	if err = db.Model(value).Count(&pagination.TotalCount).Error; err != nil {
		log.Errorf("page query BaseOrm.ListWithPage db.Count error(%v):", err)
		return nil, err
	}
	return &PageData{Pagination: pagination, Data: result}, nil
}

func (t Tx) Query(query string, args ...interface{}) (rows *sql.Rows, err error) {
	return t.DB.Raw(query, args...).Rows()
}

func (t Tx) FastQuery(query string, value interface{}, args ...interface{}) (err error) {
	var (
		targetTypeMap = make(map[string]reflect.StructField)
		targetValue   = reflect.ValueOf(value)
	)
	if targetValue.Kind() != reflect.Ptr {
		return errors.New("FastQuery must accept a ptr type")
	}
	targetValue = reflect.Indirect(targetValue)
	if targetValue.Kind() == reflect.Slice {
		rows, err := t.Query(query, args...)
		if err != nil {
			return err
		}
		dbTypes, err := rows.ColumnTypes()
		if err != nil {
			return err
		}
		scanValues := make([]interface{}, 0)
		keys := make([]string, 0)
		slice := reflect.MakeSlice(targetValue.Type(), 0, 0)
		switch targetValue.Type().Elem().Kind() {
		case reflect.Map:
			for _, v := range dbTypes {
				if targetValue.Type().Elem().Elem().Kind() == reflect.Interface {
					i := reflect.New(v.ScanType()).Interface()
					scanValues = append(scanValues, i)
				} else {
					i := reflect.New(targetValue.Type().Elem().Elem()).Interface()
					scanValues = append(scanValues, i)
				}
				keys = append(keys, v.Name())
			}
			for rows.Next() {
				refMap := reflect.MakeMap(targetValue.Type().Elem())
				if err := rows.Scan(scanValues...); err != nil {
					return err
				}
				for i, v := range scanValues {
					refMap.SetMapIndex(reflect.ValueOf(keys[i]), reflect.ValueOf(v).Elem())
				}
				slice = reflect.Append(slice, refMap)
			}
		case reflect.Struct:
			for i := 0; i < targetValue.Type().Elem().NumField(); i++ {
				fieldName := targetValue.Type().Elem().Field(i).Name
				targetTypeMap[strings.ToUpper(fieldName)] = targetValue.Type().Elem().Field(i)
			}
			for _, v := range dbTypes {
				if _, ok := targetTypeMap[strings.ToUpper(v.Name())]; !ok { //target中没有这个字段
					continue
				}
				i := reflect.New(targetTypeMap[strings.ToUpper(v.Name())].Type).Interface()
				//fmt.Println(v.ScanType().Kind() == reflect.Uint32)
				//i := reflect.New(v.ScanType()).Interface()
				scanValues = append(scanValues, i)
				keys = append(keys, strings.ToUpper(v.Name()))
			}
			for rows.Next() {
				row := reflect.New(targetValue.Type().Elem()).Interface()
				refStu := reflect.ValueOf(row).Elem()
				if err = rows.Scan(scanValues...); err != nil {
					return err
				}
				for i, v := range scanValues {
					refStu.FieldByIndex(targetTypeMap[keys[i]].Index).Set(reflect.ValueOf(v).Elem())
				}
				slice = reflect.Append(slice, refStu)
			}
		case reflect.Slice:
			for _, v := range dbTypes {
				if targetValue.Type().Elem().Elem().Kind() == reflect.Interface {
					i := reflect.New(v.ScanType()).Interface()
					scanValues = append(scanValues, i)
				} else {
					i := reflect.New(targetValue.Type().Elem().Elem()).Interface()
					scanValues = append(scanValues, i)
				}
			}
			for rows.Next() {
				refSlice := reflect.MakeSlice(targetValue.Type().Elem(), len(scanValues), len(scanValues))
				if err := rows.Scan(scanValues...); err != nil {
					return err
				}
				for i, v := range scanValues {
					refSlice.Index(i).Set(reflect.ValueOf(v).Elem())
				}
				slice = reflect.Append(slice, refSlice)
			}
		default:
			for rows.Next() {
				i := reflect.New(targetValue.Type().Elem()).Interface()
				if err = rows.Scan(i); err != nil {
					return err
				}
				slice = reflect.Append(slice, reflect.ValueOf(i).Elem())
			}
		}
		targetValue.Set(slice)
	} else {
		return errors.New("FastQuery only support a slice type")
	}
	return nil
}

func (t Tx) Exec(sql string, args ...interface{}) (rowsAffected int64, err error) {
	exec := t.DB.Exec(sql, args...)
	return exec.RowsAffected, exec.Error
}

func (t Tx) DeleteById(id interface{}, base interface{}) error {
	return t.Delete(base, "id=?", id).Error
}

func (t Tx) Create(val interface{}) error {
	return t.DB.Create(val).Error
}

/**
 * 创建或更新数据库表记录
 * @param : val 要创建的数据库对象
 * @return:
 * @author: yinjk
 * @time  : 2019/6/10 11:12
 */
func (t Tx) Save(val interface{}) error {
	return t.DB.Save(val).Error
}

func (t Tx) Update(bean interface{}) error {
	return t.DB.Model(bean).Update(bean).Error
}

/**
 * 非全量更新，只更新args参数中的字段
 * @param : bean 要更新的表的结构体
 * @return: args 要更新的字段
 * @author: yinjk
 * @time  : 2019/6/10 11:20
 */
func (t Tx) Updates(bean interface{}, args map[string]interface{}) error {
	return t.DB.Model(bean).Updates(args).Error
}

/**
 * find in
 * @param : field 查询字段
 * @param : values slice类型参数
 * @param : beans 要查询的实体
 * @return:
 * @author: yinjk
 * @time  : 2019/6/10 11:47
 */
func (t Tx) FindIn(field string, value interface{}, beans interface{}) error {
	return t.DB.Where(field+" in(?)", value).Find(beans).Error
}

func (t Tx) FindById(id interface{}, beans interface{}) error {
	if err := t.DB.Where("id=?", id).First(beans).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil
		}
		return err
	}
	return nil
}

func (t Tx) FindOneEq(beans interface{}, fieldName string, value interface{}) (err error) {
	if err = t.DB.Where(fieldName+" = ?", value).First(beans).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil
		}
		return err
	}
	return nil
}

func (t Tx) FindCondition(condition string, bean interface{}, params ...interface{}) (err error) {
	return t.DB.Where(condition, params...).Find(bean).Error
}

func (t Tx) FindOneLike(fieldName string, value interface{}, beans interface{}) (err error) {
	if err = t.DB.Where(fieldName+" like ?", value).First(beans).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil
		}
	}
	return
}

func (t Tx) Count(beans interface{}, query interface{}, args ...interface{}) (count int, err error) {
	err = t.DB.Model(beans).Where(query, args...).Count(&count).Error
	return
}

func (t Tx) ListBy(beans interface{}, fields *[]string, query interface{}, args ...interface{}) error {
	if len(*fields) == 0 {
		return t.DB.Where(query, args...).Find(beans).Error
	} else {
		return t.DB.Select(*fields).Where(query, args...).Find(beans).Error
	}
}

func (t Tx) ListEq(fieldName string, value interface{}, beans interface{}) error {
	return t.DB.Where(fieldName+" = ?", value).Find(beans).Error
}

func (t Tx) ListLike(fieldName string, value interface{}, beans interface{}) error {
	return t.DB.Where(fieldName+" like ?", value).Find(beans).Error
}

func (t Tx) ListAll(beans interface{}) error {
	return t.DB.Find(beans).Error
}

/**
 *@describe   获取单列单条数据
 *@param rowSql  原始sql语句 params 参数
 *@return 单列单条数据记录
 *@author yinchong
 *@create 2019/6/13 19:01
 */
func (t Tx) FindAloneRecord(rowSql string, params ...interface{}) (value interface{}, err error) {
	var (
		rows *sql.Rows
	)
	if rows, err = t.Query(rowSql, params...); err != nil {
		log.Error(err)
		return -1, err
	}
	defer rows.Close()
	if rows.Next() {
		if err = rows.Scan(&value); err != nil {
			log.Error(err)
			return nil, err
		}
	}
	return value, err
}

func (t Tx) Commit() *BaseOrm {
	db := t.DB.Commit()
	return &BaseOrm{db}
}

func (t Tx) Rollback() *BaseOrm {
	db := t.DB.Rollback()
	return &BaseOrm{db}
}

type Tabler interface {
	TableName() string
}
