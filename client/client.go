package main

import (
	"bufio"
	"bytes"
	"errors"
	"io"
	"log"
	"net"
	"os"
	"strconv"
)

var ErrorNocommand = errors.New("no command")

var ErrorNotEnded = errors.New("")
var ErrorNotConverttoAtoi = errors.New("")

func read(c net.Conn) error {
	for { //연결이 존재할 때
		var resultall []byte
		resultlen := 0
		data := make([]byte, 4096)
		// datag := make(chan string)

		recieive, err := c.Read(data) //server로부터 data 읽어오면
		if err != nil {
			if io.EOF == err {
				log.Println("연결 종료")
			}
			return err
		} else { //data를 읽어올 때 error 없을 경우
			if recieive > 0 {
				recvSize := data[:bytes.Index(data, []byte("\n"))] //data 첫부분에서 보낸 size check
				size, err := strconv.Atoi(string(recvSize))        //int로 converting

				log.Printf("receive size : %d\n", size)

				sizelen := len(strconv.Itoa(size))
				log.Printf("sizelen : %d\n", sizelen)

				totsize := size + sizelen
				log.Printf("totsize : %d\n", totsize)
				if err != nil {
					log.Println(err)

				} else {
					sizeindex := bytes.Index(data, []byte("\n")) + 1
					log.Printf("sizeindex : %d\n", sizeindex)

					log.Printf("receive - sizeindex : %d\n", recieive-sizeindex)

					resultall = append(resultall, data[sizeindex:recieive]...)
					resultlen += recieive

					if totsize > len(data) {
						a := 0
						for {
							log.Println("loop started")
							LargeReceive, err := c.Read(data)

							if err != nil {
								log.Printf("read LargeReceive error : %v\n", err)
							}
							log.Printf("Read LargeReceive : %v\n", LargeReceive)

							resultlen += LargeReceive

							resultall = append(resultall, data[:LargeReceive]...)

							a += 1
							log.Printf("%d번째 loop\n", a)
							if resultlen == totsize {
								log.Printf("\n%v", string(resultall[:resultlen]))
								log.Printf("resultlen : %v == totsize : %v\n", resultlen, totsize)
								break
							}
							log.Printf("NOT FINISH LOOP resultlen : %v == totsize : %v\n", resultlen, totsize)
						}
					} else {
						log.Printf("\n%v", string(resultall[:recieive-sizeindex]))

						return nil
					}
					return nil
				}
			} else {
				log.Println("No receive data")
				return nil
			}
		}
	}
}

func sending(c net.Conn) error {
	var comm string                  //command
	sc := bufio.NewScanner(os.Stdin) //init scanner

	sc.Scan() //stdinput으로 들어온 한 줄 그대로 scan
	if sc.Err() != nil {
		log.Println(sc.Err())
		// return sc.Err()
	} else {
		comm = sc.Text() //읽어온 데이터를 변수에 저장

		if comm == "" {
			log.Println("insert command!")
			// return ErrorNocommand
		} else {
			commlen := strconv.Itoa(len(comm))
			_, err := c.Write([]byte(commlen + "\n" + comm)) //server로 전송
			if err != nil {
				log.Println(err)
				// return err
			}
		}
	}
	return nil
}

func main() {
	var (
		network = "tcp"
		port    = ":3011"
	)

	conn, err := net.Dial(network, port) //client가 server와 연결할 객체 생성.

	if err != nil || conn == nil {
		log.Println(err)
		return
	}

	for {
		// err := sending(conn)
		go sending(conn)
		// if err == ErrorNocommand {
		// 	continue
		// }

		// if err != nil {
		// 	log.Printf("%v", err)
		// 	break
		// }

		err = read(conn)
		if err == ErrorNotEnded {
			continue
		}

		if err == ErrorNotConverttoAtoi {
			continue
		}

		if err != nil {
			log.Println(err)
			break
		}
	}
}
