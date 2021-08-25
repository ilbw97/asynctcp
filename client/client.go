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

var ErrorNocommand = errors.New("Insert command!")

var ErrorNotEnded = errors.New("")

var ErrorReadLargeReceive = errors.New("Cannot Read Large Data")
var ErrorEOF = errors.New("Connection is closed from server")

// var lock = new(sync.Mutex)

var connerr error

// Server가 보낸 data를 읽어 log로 출력하는 funciton
func read(c net.Conn) error {

	for { //연결이 존재할 때
		var resultall []byte
		resultlen := 0
		data := make([]byte, 4096)

		receive, err := c.Read(data) //server로부터 data 읽어오면

		if err != nil {
			if err == io.EOF {
				connerr = ErrorEOF //conn error 공유
				return ErrorEOF
			}

			return err
		} else { //data를 읽어올 때 error 없을 경우
			if receive > 0 {
				recvSize := data[:bytes.Index(data, []byte("\n"))] //data 첫부분에서 보낸 size check
				size, err := strconv.Atoi(string(recvSize))        //int로 converting

				if err != nil {
					panic(err)
				} else {
					sizelen := len(strconv.Itoa(size)) //server가 명시한 data의 전체 길이(cmd.Output()의 길이)

					totsize := size + sizelen //server가 보낸 data의 전체 길이
					sizeindex := bytes.Index(data, []byte("\n")) + 1

					resultall = append(resultall, data[sizeindex:receive]...)
					resultlen += receive

					if totsize > len(data) {
						a := 0
						for {
							log.Println("loop started")
							LargeReceive, err := c.Read(data)

							if err != nil {
								return err
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
							log.Printf("NOT FINISH LOOP resultlen : %v != totsize : %v\n", resultlen, totsize)
						}
					} else {
						log.Printf("\n%v", string(resultall[:receive-sizeindex]))

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

// Stdin으로 입력받은 data를 Server에게 전송하는 function
func sending(c net.Conn, sendingerr chan error) {

	sc := bufio.NewScanner(os.Stdin) //init scanner

	sc.Scan()
	if sc.Err() != nil {
		sendingerr <- sc.Err()
		return
	} else {
		var comm string
		comm = sc.Text()

		//connection error 존재 시
		if connerr != nil {
			sendingerr <- connerr
		} else {
			if comm == "" {
				sendingerr <- ErrorNocommand
			} else {
				commlen := strconv.Itoa(len(comm))
				_, er := c.Write([]byte(commlen + "\n" + comm)) //server로 전송
				if er != nil {
					sendingerr <- er
				}
			}
		}
		sendingerr <- nil
	}
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

	go func() {
		for {
			readerr := read(conn)
			switch readerr {
			case nil:
				log.Println("Complete reading")
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

		go sending(conn, sendingerr)

		e := <-sendingerr

		switch e {
		case ErrorNocommand:
			log.Println(e)
			continue
		case nil:
			log.Println("Sending Complete")
		default:
			log.Println(e)
			break
		}

	}

}
