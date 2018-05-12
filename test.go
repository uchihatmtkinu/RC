package main

import (
	"crypto/sha256"
	"fmt"
	"os"
	"time"

	"github.com/uchihatmtkinu/RC/account"
)

//"crypto/sha512"

func main() {
	/*
		outputFile, outputError := os.OpenFile("output.dat", os.O_WRONLY|os.O_CREATE, 0666)
		if outputError != nil {
			fmt.Printf("An error occurred with file opening or creation\n")
			return
		}
		defer outputFile.Close()
		for i := 1; i < 10; i++ {
			ID := "test" + string(i)
			var acc account.RcAcc
			acc.New(ID, 0)
			var str1 string
			str1 = acc.RetPri().D.String()
			fmt.Fprintf(outputFile, "%s\n", str1)
			str1 = acc.Puk.X.String()
			fmt.Fprintf(outputFile, "%s\n", str1)
			str1 = acc.Puk.Y.String()
			fmt.Fprintf(outputFile, "%s\n", str1)
			fmt.Fprintf(outputFile, "%s\n", acc.Addr)
			fmt.Fprintln(outputFile, acc.AccType)
		}
	*/

	file, _ := os.Open("output.dat")
	defer file.Close()
	var acc [10]account.RcAcc
	for i := 1; i < 10; i++ {
		var str1, str2, str3, str4, str5 string
		_, _ = fmt.Fscanln(file, &str1)
		_, _ = fmt.Fscanln(file, &str2)
		_, _ = fmt.Fscanln(file, &str3)
		_, _ = fmt.Fscanln(file, &str4)
		_, _ = fmt.Fscanln(file, &str5)
		acc[i].Load(str1, str2, str3, str4, str5)
		//fmt.Println(acc[i].AccType)
		//fmt.Println(acc[i].Puk.X)
		//fmt.Println(acc[i].AddrReal)
		//fmt.Println(cryptonew.Verify(acc[i].Puk, acc[i].AddrReal))
		tmp := acc[i].Puk.X.Bytes()
		fmt.Println(len(tmp))
	}

	test := []byte("123456789192387519837591837591375981273501234567890")
	t1 := time.Now()
	for i := 0; i < 2000000; i++ {
		sha256.Sum256(test)

	}
	elapsed := time.Since(t1)

	fmt.Println(elapsed)
}
