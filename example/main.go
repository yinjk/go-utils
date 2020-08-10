/*
 @Desc

 @Date 2020-04-17 18:35
 @Author yinjk
*/
package main

import (
	"fmt"
	"github.com/yinjk/go-utils/pkg/net/solace"
)

func main() {
	context := solace.NewContext()
	fmt.Println(context)
}
