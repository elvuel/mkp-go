package mkpgo

/*

Example:

```go
	package main

	import (
		"fmt"

		mkpgo "github.com/elvuel/mkp-go"
		"github.com/elvuel/mkp-go/helper"
	)

	func main() {
		sfport := mkpgo.NewSFSerialPort()
		sfport.Name = "COM5"

		err := sfport.Open()
		if err != nil {
			panic(err)
		}
		defer sfport.Close()
		go sfport.Read()

		sfport.SyncOuputEnabled = true
		fs, err := helper.ListDir(sfport, "/eMMC/applog/mkpdemo")
		fmt.Println(fs, err)
	}
```
*/
