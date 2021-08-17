package main

import "fmt"

// func main() {
// 	var a, b = 10, 5
// 	var result int

// 	c := make(chan int)

// 	go func() {
// 		c <- a + b //channel 'c'에 a+b한 값을 보냄
// 	}()

// 	result = <-c //channel c에서 값을 받아 result에 대입.
// 	fmt.Printf("두 수의 합은 %d 입니다.\n", result)
// }

// func main() { //channel에 값이 수신될대까지 대기
// 	var str = "Hello Goorm!"
// 	done := make(chan bool)

// 	go func() {
// 		for i := 0; i < 10; i++ {
// 			fmt.Println(str, i)
// 		}

// 		done <- true //channle에 true 전송
// 	}()
// 	fmt.Println("loop finished")
// 	<-done //channel done에서 수신해 대기 끝냄
// }

// func main() {
// 	c := make(chan string, 1)
// 	c <- "Hello goorm!"
// 	fmt.Println(<-c)
// }

// func main() {
// 	done := make(chan bool, 2)

// 	go func() {
// 		for i := 0; i < 6; i++ {
// 			done <- true

// 			fmt.Println("goroutine : ", i)
// 		}
// 	}()

// 	for i := 0; i < 6; i++ {
// 		<-done
// 		fmt.Println("main : ", i)
// 	}
// // }

// func main() {
// 	c := make(chan string, 2) // 버퍼 2개 생성

// 	// 채널(버퍼)에 송신
// 	c <- "Hello"
// 	c <- "goorm"

// 	close(c) // 채널 닫음

// 	// 채널 수신
// 	fmt.Println(<-c)
// 	fmt.Println(<-c)
// 	fmt.Println(<-c) // 무한 대기 상황 발생 x
// 	fmt.Println(<-c)
// }

// func main() { //channel 개폐 여부 확인
// 	c := make(chan string, 2)

// 	c <- "Hello"
// 	c <- "goorm"

// 	close(c)

// 	val, open := <-c
// 	fmt.Println(val, open)
// 	val, open = <-c
// 	fmt.Println(val, open)
// 	val, open = <-c
// 	fmt.Println(val, open)
// 	val, open = <-c
// 	fmt.Println(val, open)
// }

// func main() {
// 	c := make(chan int, 10)

// 	for i := 0; i < 10; i++ {
// 		c <- i
// 	}
// 	close(c)

// 	for val := range c { // <- c를 사용하지 않음
// 		fmt.Println(val)
// 	}
// }

// func main() {
// 	c := make(chan int)

// 	go channel1(c)
// 	go channel2(c)

// 	fmt.Scanln()
// }

// func channel1(ch chan int) {
// 	ch <- 1
// 	ch <- 2

// 	fmt.Println(<-ch)
// 	fmt.Println(<-ch)

// 	fmt.Println("done1")
// }

// func channel2(ch chan int) {
// 	fmt.Println(<-ch)
// 	fmt.Println(<-ch)

// 	ch <- 3
// 	ch <- 4

// 	fmt.Println("done2")
// }

// func main() {
// 	c := make(chan int)

// 	go sendChannel(c)
// 	go receiveChannel(c)

// 	fmt.Scanln()
// }

// func sendChannel(ch chan<- int) {
// 	ch <- 1
// 	ch <- 2
// 	// <-ch 오류 발생
// 	fmt.Println("done1")
// }

// func receiveChannel(ch <-chan int) {
// 	fmt.Println(<-ch)
// 	fmt.Println(<-ch)
// 	//ch <- 3 오류 발생
// 	fmt.Println("done2")
// }

// func main() {
// 	ch := sum(10, 5)

// 	fmt.Println(<-ch)
// }

// func sum(num1, num2 int) <-chan int {
// 	result := make(chan int)

// 	go func() { //고루틴 생성
// 		result <- num1 + num2 //연산 및 채널에 데이터 송신
// 	}()

// 	return result //수신 채널 반환
// }

func main() {
	numsch := num(10, 5)
	result := sum(numsch)

	//channel result는 수신만 가능
	fmt.Println(<-result)
}

func num(num1, num2 int) <-chan int {
	numch := make(chan int)

	go func() {
		numch <- num1
		numch <- num2 //송신 후
		close(numch)
	}()
	return numch //수신 전용 채널로 반환
}

func sum(c <-chan int) <-chan int {
	//channel c는 수신만 할 수 있음.
	sumch := make(chan int)

	go func() {
		r := 0
		for i := range c { //channel c로부터 수신
			r = r + i
		}
		sumch <- r //송신 후
	}()

	return sumch //수신 전용 채널로 반환
}
