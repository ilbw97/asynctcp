package main

import (
	"bytes"
	"errors"
	"io"
	"log"
	"net"
	"os/exec"
	"strconv"
)

type data struct {
}

func main() {
	var (
		network = "tcp"
		port    = ":3011"
	)
	server, err := net.Listen(network, port) //socket 열어준다..

	if nil != err {
		log.Printf("Failed to Listen : %v\n", err)
	}
	defer server.Close()

	for {
		conn, err := server.Accept() //연결 기다림.. 계속 block해서 기다리다가, 연결이 들어왔을 경우 값을 return
		if nil != err {
			log.Printf("Accept Error : %v\n", err)
			continue
		}
		// recvall := make(chan []byte)
		// recvlen := make(chan int, 0)
		// totsize := make(chan int, 0)
		// go CommRead(conn, recvall, recvlen, totsize) //연결을 parameter로 넘겨주고 ConnHandler go routine 실행 (goroutine 실행 이유 : client는 여러개일수도 있으니까.)
		go CommRead(conn)
		// sendingall := <-recvall
		// sendinglen := <-recvlen
		// sendingsize := <-totsize
		// log.Printf("sendingall : %v\n", sendingall)
		// log.Printf("sendinglen : %v\n", sendinglen)
		// log.Printf("sendingsize : %v\n", sendingall)

		// go CommSend(conn, sendingall, sendinglen, sendingsize)

	}
}

var ErrorConfirmData = errors.New("")

// func CommSend(conn net.Conn, sendingall []byte, sendinglen int, sendingsize int) {
// 	defer conn.Close()
// 	// sendingall = sendingall
// 	// resultlarge := execute(string(sendingall[:sendinglen]))
// 	go execute(string(sendingall)[:sendinglen])
// 	_, err := conn.Write(resultlarge)
// 	if err != nil {
// 		log.Printf("write err : %v\n", err)
// 		// return
// 	}
// 	log.Printf("resultlen : %v == totsize : %v\n", sendinglen, sendingsize)
// }

func CommRead(conn net.Conn) {
	log.Printf("serving %s\n", conn.RemoteAddr().String())
	defer conn.Close()
	for {
		var recvall []byte
		// recvall := make(chan []byte)
		// recvlen := make(chan int, 0)
		recvlen := 0
		recvData := make([]byte, 4096)          //값을 읽어와 저장할 버퍼 생성
		recvCommand, err := conn.Read(recvData) //client가 값을 줄 때까지 blocking 되어 대기하다가 값을 주면 읽어들인다.
		resultlarge := make(chan []byte)
		if nil != err { //입력이 종료되면 종료
			if io.EOF == err {
				log.Printf("connection is closed from client : %v\n", conn.RemoteAddr().String())
				return
			}
			log.Printf("Failed to receive data : %v\n", err)
		} else {
			if recvCommand > 0 { //받은 data len이 0보다 클 때 ==> 정상동작
				//보낸 command size 구하고
				recvSize := recvData[:bytes.Index(recvData, []byte("\n"))]
				size, err := strconv.Atoi(string(recvSize))
				if err != nil {
					return
				}
				log.Printf("receive command size : %d\n", size)
				sizelen := len(strconv.Itoa(size))
				log.Printf("sizelen : %d\n", sizelen)

				totsize := size + sizelen
				log.Printf("totsize : %d\n", totsize)
				if err != nil {
					log.Println(err)
				} else {
					sizeindex := bytes.Index(recvData, []byte("\n")) + 1
					log.Printf("sizeindex : %d\n", sizeindex)

					log.Printf("receive - sizeindex : %d\n", recvCommand-sizeindex)

					log.Printf("recvlen : %v\n", recvlen)

					recvall = append(recvall, recvData[sizeindex:recvCommand]...)
					recvlen += recvCommand - sizeindex

					if totsize > len(recvData) {
						a := 0
						for {
							log.Println("loop started")
							LargeReceive, err := conn.Read(recvData)

							if err != nil {
								log.Printf("read LargeReceive error : %v\n", err)
							}
							log.Printf("Read LargeReceive : %v\n", LargeReceive)

							recvlen += LargeReceive
							recvall = append(recvall, recvData[:LargeReceive]...)

							a += 1
							log.Printf("%d번째 loop\n", a)
							if recvlen == totsize {
								// sendingall <- recvall
								// sendinglen <- recvlen
								// sendingsize <- totsize
								go execute(string(recvall[:recvlen]), resultlarge)
								sending := <-resultlarge
								_, err := conn.Write(sending)
								if err != nil {
									log.Printf("write err : %v\n", err)
									return
								}
								log.Printf("resultlen : %v == totsize : %v\n", recvlen, totsize)
							}
							// return
							log.Printf("NOT FINISH LOOP resultlen : %v == totsize : %v\n", recvlen, totsize)
						}

					} else {
						// sendingall <- recvall
						// sendinglen <- recvlen
						// sendingsize <- totsize
						go execute(string(recvall[:recvlen]), resultlarge)
						sending := <-resultlarge
						_, err := conn.Write(sending)
						if err != nil {
							log.Printf("write err : %v\n", err)
							return
						}
						log.Printf("resultlen : %v == totsize : %v\n", recvlen, totsize)
					}

					// return
				}
			} else {
				log.Println("No receive data")
				return
			}

		}
	}
}

func execute(command string, resultlarge chan []byte) {
	cmd := exec.Command("bash", "-c", command)
	log.Printf("Execute Command : %v\n", cmd)

	cmdres, err := cmd.Output()
	cmdreslen := []byte(strconv.Itoa(len(string(cmdres)+"\n")) + "\n")
	if err != nil {
		log.Println(err)
		cmdreslen = []byte(strconv.Itoa(len("Command error : "+err.Error()+"\n")) + "\n")
		resultlarge <- append(cmdreslen, []byte("Command error : "+err.Error())...)
	}
	if string(cmdres) == "" {
		cmdreslen = []byte(strconv.Itoa(len("No output data"+"\n")) + "\n")
		resultlarge <- append(cmdreslen, ([]byte("No output data"))...)
	}

	log.Printf("stdout: %v bytes\n%s", string(bytes.Trim(cmdreslen, "\n")), string(cmdres))

	resultlarge <- append(cmdreslen, cmdres...)
	// return append(cmdreslen, cmdres...)
}
