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
	"syscall"
)

var ErrorNocommand = errors.New("no command")

var ErrorNotEnded = errors.New("")
var ErrorNotConverttoAtoi = errors.New("")

var ErrorReadLargeReceive = errors.New("Cannot Read Large Data")
var ErrorEOF = errors.New("연결 종료")
var ErrorBreakpipe = errors.New("Cannot Sending Data")
var lock = new(sync.Mutex)

func read(c net.Conn) error {

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
				//log.Println("연결 종료")
				return ErrorEOF
			}
			return err
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
								//log.Printf("Read LargeReceive error : %v\n", err)
								return ErrorReadLargeReceive
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

func sending(c net.Conn, sendingerr chan error) {

	sc := bufio.NewScanner(os.Stdin) //init scanner

	sc.Scan() //stdinput으로 들어온 한 줄 그대로 scan
	if sc.Err() != nil {
		//if errors.Is(sc.Err(), syscall.EPIPE) {
		//	sendingerr <- ErrorBreakpipe
		//	return
		//}
		//log.Println(sc.Err())
		sendingerr <- sc.Err()
		return
	} else {
		var comm string  //command
		comm = sc.Text() //읽어온 데이터를 변수에 저장

		if comm == "" {
			log.Println("insert command!")
			sendingerr <- ErrorNocommand
			return
		} else {
			commlen := strconv.Itoa(len(comm))
			_, er := c.Write([]byte(commlen + "\n" + comm)) //server로 전송
			if er != nil {
				if errors.Is(er, syscall.EPIPE) {
					sendingerr <- ErrorBreakpipe
				} else {
					//log.Println(er)
					sendingerr <- er
				}
				return
			}
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

	if err != nil || conn == nil {
		log.Println(err)
		return
	}
	log.Println("read goroutine make")
	go func() {
		for {
			readerr := read(conn)
			switch readerr {
			case nil:
				log.Println("Read Complete")
				return
			case ErrorEOF:
				log.Println(readerr)
				return
			case ErrorReadLargeReceive:
				log.Println(readerr)
				return
			default:
				log.Println(readerr)
				return
			}
		}

	}()

	for {
		defer conn.Close()
		sendingerr := make(chan error)

		log.Println("sending go make")
		go sending(conn, sendingerr)

		e := <-sendingerr

		switch e {
		case ErrorBreakpipe:
			log.Printf("%v\n", e)
			break
		case ErrorNocommand:
			continue
		default:
			if e == nil {
				log.Println("Sending Complete")
			} else {
				log.Printf("%v\n", e)
				break
			}
		}
	}
}
