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
	"sync"
)

var ErrorNocommand = errors.New("no command")

var ErrorNotEnded = errors.New("")
var ErrorNotConverttoAtoi = errors.New("")
var lock = new(sync.Mutex)

func read(c net.Conn) {

	for { //연결이 존재할 때
		var resultall []byte
		resultlen := 0
		data := make([]byte, 4096)

		lock.Lock()
		log.Println("lock")
		receive, err := c.Read(data) //server로부터 data 읽어오면
		lock.Unlock()
		log.Println("unlock")
		if err != nil {
			if err == io.EOF {
				log.Println("연결 종료")
			}
			break
		} else { //data를 읽어올 때 error 없을 경우
			if receive > 0 {
				recvSize := data[:bytes.Index(data, []byte("\n"))] //data 첫부분에서 보낸 size check
				size, err := strconv.Atoi(string(recvSize))        //int로 converting

				log.Printf("receive size : %d\n", size)

				lock.Lock()
				sizelen := len(strconv.Itoa(size)) //server가 명시한 data의 전체 길이(cmd.Output()의 길이)
				log.Printf("sizelen : %d\n", sizelen)
				lock.Unlock()

				totsize := size + sizelen //server가 보낸 data의 전체 길이
				log.Printf("totsize : %d\n", totsize)
				if err != nil {
					panic(err)
				} else {

					sizeindex := bytes.Index(data, []byte("\n")) + 1
					log.Printf("sizeindex : %d\n", sizeindex)

					log.Printf("receive - sizeindex : %d\n", receive-sizeindex)

					lock.Lock()
					resultall = append(resultall, data[sizeindex:receive]...)
					resultlen += receive
					lock.Unlock()
					log.Println("unlock")

					if totsize > len(data) {
						a := 0
						for {
							log.Println("loop started")
							lock.Lock()
							LargeReceive, err := c.Read(data)
							lock.Unlock()

							if err != nil {
								log.Printf("Read LargeReceive error : %v\n", err)
								return
							}
							log.Printf("Read LargeReceive : %v\n", LargeReceive)

							lock.Lock()
							log.Println("lock")
							resultlen += LargeReceive
							resultall = append(resultall, data[:LargeReceive]...)
							lock.Unlock()
							log.Println("unlock")

							a += 1
							log.Printf("%d번째 loop\n", a)
							if resultlen == totsize {
								lock.Lock()
								log.Println("lock")
								log.Printf("\n%v", string(resultall[:resultlen]))
								log.Printf("resultlen : %v == totsize : %v\n", resultlen, totsize)
								lock.Unlock()
								log.Println("unlock")
								break
							}
							log.Printf("NOT FINISH LOOP resultlen : %v != totsize : %v\n", resultlen, totsize)
						}
					} else {
						lock.Lock()
						log.Println("lock")
						log.Printf("\n%v", string(resultall[:receive-sizeindex]))
						lock.Unlock()
						log.Println("unlock")
						return
					}
					return
				}
			} else {
				log.Println("No receive data")
				return
			}
		}
	}
}

func sending(c net.Conn, sendingerr chan error) {
	sc := bufio.NewScanner(os.Stdin) //init scanner

	sc.Scan() //stdinput으로 들어온 한 줄 그대로 scan
	if sc.Err() != nil {
		log.Println(sc.Err())
		sendingerr <- sc.Err()
	} else {
		var comm string  //command
		comm = sc.Text() //읽어온 데이터를 변수에 저장

		if comm == "" {
			log.Println("insert command!")
			sendingerr <- ErrorNocommand

		} else {
			commlen := strconv.Itoa(len(comm))
			_, er := c.Write([]byte(commlen + "\n" + comm)) //server로 전송
			if er != nil {
				log.Println(er)
				sendingerr <- er
			}
			// // log.Println("sending complete")

			// go read(c)

		}
	}
	sendingerr <- nil
}

func main() {
	var (
		network = "tcp"
		port    = ":3011"
	)

	conn, err := net.Dial(network, port) //client가 server와 연결할 객체 생성.
	// conn.SetDeadline()

	if err != nil || conn == nil {
		log.Println(err)
		return
	}

	go func() {
		for {
			read(conn)
		}
	}()

	for {
		defer conn.Close()
		sendingerr := make(chan error)
		// wg := sync.WaitGroup{}
		// wg.Add(1)
		log.Printf("go make")
		go sending(conn, sendingerr)
		// err := sending(conn)
		e := <-sendingerr
		if e == ErrorNocommand {
			continue
		}

		if e != nil {
			log.Printf("%v", e)
			break
		}
		if e == nil {
			log.Println("sending complete")
		}
		// go read(conn)
		//
	}

}
