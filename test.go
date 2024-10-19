package main

import (
	"fmt"
)


func main() {
	fmt.Println("这个文件是本次作业中遇到的Go 语言问题学习记录，无参考价值，留个纪念。")
	// b, err := ioutil.ReadFile(path.Join("./chandy_lamport/test_data", "8nodes-concurrent-snapshots.events"))
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// count = 5
	// if count%5 == 0 {
	// 	fmt.Println("count/5 =0 ")
	// }
	// //将文件内容按行拆分成不同字符串
	// lines := strings.FieldsFunc(string(b), func(r rune) bool { return r == '\n' })
	// snapshots := make([]int, 0)
	// getSnapshots := make(chan int, 100)
	// numSnapshots := 0
	// for _, line := range lines {
	// 	// Ignore comments忽略注释
	// 	if strings.HasPrefix("#", line) {
	// 		continue
	// 	}
	// 	parts := strings.Fields(line)
	// 	switch parts[0] {
	// 	case "send":
	// 		//send就读取源server号，目的server号
	// 		src := parts[1]
	// 		dest := parts[2]
	// 		//token号
	// 		tokens, err := strconv.Atoi(parts[3])
	// 		fmt.Println("send:", tokens, "src:", src, "dest:", dest)
	// 		if err != nil {
	// 			log.Fatal(err)
	// 		}
	// 	case "snapshot":
	// 		numSnapshots++
	// 		serverId := parts[1]
	// 		fmt.Println("snapshot", serverId)
	// 		fmt.Println("snapshot count", count)
	// 		go func(id int) {
	// 			fmt.Println("go func count", count)
	// 			if count < 50000 {
	// 				if count%5 == 0 {
	// 					fmt.Println("getSnap", serverId, " ", count)
	// 					getSnapshots <- i
	// 					i++
	// 				}
	// 			}
	// 		}(0)
	// 	case "tick":
	// 		numTicks := 1       //默认没有第二个参数，就tick 1次
	// 		if len(parts) > 1 { //如果有第二个参数  tick 4 10等
	// 			//字符串转换为整数
	// 			numTicks, err = strconv.Atoi(parts[1])
	// 			if err != nil {
	// 				log.Fatal(err)
	// 			}
	// 		}
	// 		//至少tick 1次
	// 		for i := 0; i < numTicks; i++ {
	// 			fmt.Println("ticking")
	// 		}
	// 	default:
	// 	}
	// }
	// for numSnapshots > 0 {
	// 	select {
	// 	//算法结束，将所有server的状态进行保存
	// 	case snap := <-getSnapshots:
	// 		snapshots = append(snapshots, snap)
	// 		numSnapshots--
	// 		//算法没结束 继续tick
	// 	default:
	// 		count++
	// 		//fmt.Println("select ticking")
	// 	}
	// }
	// fmt.Println(snapshots)

	//var a map[string]string
	// 在使用 map 前，需要先 make,make 的作用就是给 map 分配数据空间
	//a = make(map[string]string, 10)
	// a = make(map[string]string)
	// a["no1"] = "宋江"
	// a["no2"] = "吴用"
	// a["no1"] = "武松"
	// a["no3"] = "吴用"
	// rand.Seed(time.Now().UnixNano())
	// fmt.Println(m.GetSortedKeys(a))
	// for _, serverId := range m.GetSortedKeys(a) {
	// 	fmt.Println(a[serverId])
	// 	fmt.Println(rand.Intn(9))
	// }

	// b := [2][2]int{{1, 2}, {3, 4}}
	// fmt.Println(len(b))
	// c := b[1]
	// fmt.Println(c)

	// var a int
	// var b bstruct
	// typeOfA := reflect.TypeOf(a)
	// typeOfB := reflect.TypeOf(b)
	// fmt.Println(typeOfA.Name(), typeOfA.Kind())
	// fmt.Println(typeOfB.Name(), typeOfB.Kind())

	// var x float64 = 7.1
	// v := reflect.ValueOf(x)
	// y := v.Interface().(float64) // y will have type float64.
	// fmt.Println(y)

	// var n = [10]int{1, 2, 3, 4, 5}
	// m := make([]int, 5)
	// m = n[:3]
	// fmt.Println(m)
	// fmt.Println(n[:4])
	// fmt.Println("123")
	// var a *int
	// fmt.Println(a)
	// kvs := map[int]string{1: "apple", 2: "banana"}
	// for k, v := range kvs {
	// 	fmt.Printf("%d -> %s\n", k, v)
	// }
	// var ab []int
	// var ac map[string]int
	// var ad chan int
	// var ae func(string) int
	// var af error // error 是接口
}

// package main

// import "fmt"

// func sum(s []int, c chan int) {
// 	sum := 0
// 	for _, v := range s {
// 		sum += v
// 	}
// 	c <- sum // 把 sum 发送到通道 c
// }

// func main() {
// 	s := []int{7, 2, 8, -9, 4, 0}

// 	c := make(chan int)
// 	fmt.Println(len(s) / 2)
// 	fmt.Println(s[:len(s)/2])
// 	fmt.Println(s[len(s)/2:])
// 	go sum(s[:len(s)/2], c)
// 	go sum(s[len(s)/2:], c)
// 	x, y := <-c, <-c // 从通道 c 中接收

// 	fmt.Println(x, y, x+y)
// }
