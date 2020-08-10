package httpclient

import (
	"bytes"
	"encoding/json"
	"github.com/yinjk/go-utils/pkg/net/common"
	"github.com/prometheus/common/log"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
)

const (
	ContentTypeKey  = "Content-type"
	ContentTypeJson = "application/json"
	ContentTypeForm = "application/x-www-form-urlencoded"
)

// 发送http post请求，v表示要post的结构体，最终会被解析成json，rv是返回post请求返回的数据，可传nil
// 此时返回的结果会解析成一个map[string] interface{} 放在Data属性中，若是想将返回结果解析成指定结构体，
// 则需要将该结构体的地址通过rv传给方法
// For example:
//
// type Person struct {
//      Name string `json:"name"`
//      Age  string `json:"age"`
// }
//
// 直接将结果解析成map：
// httpclient.PostValue("localhost:8080/hello", nil, nil, nil)
//
// 将结果解析成Person结构体：
// var p Person
// response, e := httpclient.PostValue("localhost", nil, nil, &p)
//
func PostValue(url string, h Header, v interface{}, rv interface{}) (*common.Result, error) {
	jsonBytes, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}
	return PostJson(url, h, string(jsonBytes), rv)
}

//PostJson 通过post方法提交json格式请求体的http请求
func PostJson(url string, h Header, jsonStr string, rv interface{}) (*common.Result, error) {
	client := &http.Client{}
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer([]byte(jsonStr)))
	if err != nil {
		return nil, err
	}
	if h != nil && len(h) > 0 {
		for k, v := range h {
			req.Header[k] = v
		}
	}
	req.Header.Set(ContentTypeKey, ContentTypeJson)
	//执行请求
	response, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	//封装返回结果
	resp := &common.Result{Code: response.StatusCode}
	if rv == nil {
		resp.Data = make(map[string]interface{})
	} else {
		resp.Data = rv
	}
	body, _ := ioutil.ReadAll(response.Body)
	if response.StatusCode == http.StatusOK { //响应码200，表示响应成功
		rvs, ok1 := rv.(*string)
		_, ok2 := rv.(string)
		if ok1 || ok2 { //如果返回类型是字符串类型，直接返回字符串，不解析json
			resp.Data = string(body)
			*rvs = string(body)
			return resp, nil
		}
		err := json.Unmarshal(body, &resp.Data) //将结果解析成结构体或map类型
		if err != nil {
			return resp, err
		}
	} else { //响应异常
		resp.Message = string(body)
		rvs, ok := rv.(*string)
		if ok {
			*rvs = string(body)
		}

	}
	return resp, nil
}

//FormValue http post form values
type FormValue struct {
	url.Values
}

func NewFormValue() *FormValue {
	value := FormValue{Values: make(url.Values)}
	return &value
}

//PostForm 通过post请求提交http表单
func PostForm(urls string, h Header, value *FormValue, rv interface{}) (*common.Result, error) {
	var body io.ReadCloser
	if value != nil { //form表单body
		body = ioutil.NopCloser(strings.NewReader(value.Encode()))
	}
	client := &http.Client{}
	request, err := http.NewRequest(http.MethodPost, urls, body)
	if err != nil {
		return nil, err
	}
	if h != nil && len(h) > 0 {
		for k, v := range h {
			request.Header[k] = v
		}
	}
	request.Header.Set(ContentTypeKey, ContentTypeForm) //设置为 form 请求
	resp, err := client.Do(request)                     //发送请求
	if err != nil {
		return nil, err
	}
	return parseResponse(resp, rv)
}

//GetStream 通过get请求获取http原始返回流
func GetStream(url string) (data []byte, err error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	buf := make([]byte, 3072)
	for {
		n, err := resp.Body.Read(buf)
		if n > 0 {
			temp := buf[:n]
			data = append(data, temp...)
		}

		if err != nil {
			if err == io.EOF {
				break
			}
		}
	}
	return data, nil
}

//Get 通过get方法发送http请求
func Get(url string, rv interface{}) (*common.Result, error) {
	response, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	return parseResponse(response, rv)
}

//GetParam 通过GET方法带url参数的方式提交http请求
func GetParam(url string, args *FormValue, rv interface{}) (*common.Result, error) {
	url = EncodeUrl(url, args)
	//fmt.Println(url)
	return Get(url, rv)
}

//对client.Do的简单封装
func Execute(method, url string, h Header, body io.Reader, rv interface{}) (*common.Result, error) {
	client := &http.Client{}
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, err
	}
	if h != nil && len(h) > 0 {
		for k, v := range h {
			req.Header[k] = v
		}
	}
	//执行请求
	response, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	return parseResponse(response, rv)
}

func EncodeUrl(url string, value *FormValue) string {
	if value != nil {
		encodeArgs := value.Encode()
		if encodeArgs != "" {
			url = url + "?" + encodeArgs
		}
	}
	return url
}

//对原生返回结果进行封装
func parseResponse(response *http.Response, rv interface{}) (*common.Result, error) {
	//封装返回结果
	resp := &common.Result{Code: response.StatusCode}
	if rv == nil {
		resp.Data = make(map[string]interface{})
	} else {
		resp.Data = rv
	}
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		log.Errorf("parse response failed, err = %v", err)
	}
	defer func() { _ = response.Body.Close() }() //关闭resp.Body
	if response.StatusCode == http.StatusOK {    //响应码200，表示响应成功
		//if body != nil {
		//	utils.Debugf("http exec return message:%v", string(body))
		//}
		rvs, ok1 := rv.(*string)
		_, ok2 := rv.(string)
		if ok1 || ok2 { //如果返回类型是字符串类型，直接返回字符串，不解析json
			resp.Data = string(body)
			if ok1 {
				*rvs = string(body)
			}
			return resp, nil
		}
		err := json.Unmarshal(body, &resp.Data) //将结果解析成结构体或map类型
		if err != nil {
			return resp, err
		}
	} else if response.StatusCode == http.StatusNoContent {
		// 该状态码表示没有返回内容，doing nothing...
	} else { //响应异常
		if body != nil {
			log.Errorf("http exec return error message:%v", string(body))
		}
		resp.Message = string(body)
		rvs, ok := rv.(*string)
		if ok {
			*rvs = string(body)
		}
	}
	return resp, nil
}
