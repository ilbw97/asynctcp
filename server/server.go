package main

import (
	"bytes"
	"io"
	"log"
	"net"
	"os/exec"
	"strconv"
	"sync"
)

type Data struct {
	Recvall []byte
	Totsize int
	Lock    sync.Mutex
	Connerr error
	Command string
	Result  []CommResult
}
type CommResult struct {
	Sendingsize int
	Result      string
}

func main() {
	var (
		network = "tcp"
		port    = ":3011"
	)
	server, err := net.Listen(network, port)
	if nil != err {
		log.Printf("Failed to Listen : %v\n", err)
	}
	defer server.Close()
	for {
		d := Data{}
		conn, err := server.Accept()
		if nil != err {
			log.Printf("Accept Error : %v\n", err)
			continue
		}

		go d.CommRead(conn)
	}
}

//Client로부터 data를 읽어와, linux command를 추출해 Execute function에 전달하는 함수.
func (d *Data) CommRead(conn net.Conn) {

	defer conn.Close()

	log.Printf("serving %s\n", conn.RemoteAddr().String())
	for {
		var recvall []byte
		recvlen := 0
		recvData := make([]byte, 4096)

		recvCommand, err := conn.Read(recvData)
		if err != nil {
			if err == io.EOF {
				log.Printf("Connection is closed from Client : %v\n", conn.RemoteAddr().String())
				d.Connerr = err
				return
			}

			log.Printf("Failed to receive data : %v\n", err)

		} else {

			if recvCommand > 0 {
				recvSize := recvData[:bytes.Index(recvData, []byte("\n"))]
				size, err := strconv.Atoi(string(recvSize))
				if err != nil {
					return
				} else {
					sizelen := len(strconv.Itoa(size))

					totsize := size + sizelen

					d.Totsize = totsize
					sizeindex := bytes.Index(recvData, []byte("\n")) + 1

					recvall = append(recvall, recvData[sizeindex:recvCommand]...)

					recvlen += recvCommand - sizeindex

					if totsize > len(recvData) {
						a := 0
						for {
							log.Println("Loop started")

							LargeReceive, err := conn.Read(recvData)
							if err != nil {
								log.Printf("Read LargeReceive error : %v\n", err)
							}

							log.Printf("Read LargeReceive : %v\n", LargeReceive)

							recvlen += LargeReceive
							recvall = append(recvall, recvData[:LargeReceive]...)
							a += 1
							log.Printf("%d번째 loop\n", a)

							if recvlen == totsize {
								d.Lock.Lock()
								d.Command = string(recvall[:recvlen])
								log.Println(d.Command)

								go d.Execute(conn)
								d.Lock.Unlock()

							}

							log.Printf("NOT FINISH LOOP resultlen : %v == totsize : %v\n", recvlen, totsize)
						}
					} else {
						d.Lock.Lock()
						d.Command = string(recvall[:recvlen])
						log.Println(d.Command)
						go d.Execute(conn)
						d.Lock.Unlock()
					}
				}
			} else {
				log.Println("No receive data")
				return
			}
		}
	}
}

// Client로부터 받은 data 중, 추출한 linux command를 실행하는 function.
func (d *Data) Execute(conn net.Conn) {
	var commres CommResult

	cmd := exec.Command("bash", "-c", d.Command)
	log.Printf("Execute Command : %v\n", cmd)

	cmdres, err := cmd.Output()
	strcmdres := string(cmdres)
	cmdreslen := strconv.Itoa(len(string(cmdres)+"\n")) + "\n"

	switch err {
	case nil:
		if string(cmdres) == "" {
			cmdreslen = strconv.Itoa(len("No output data"+"\n")) + "\n"
			d.Lock.Lock()

			commres.Result = cmdreslen + "No output data"
			commres.Sendingsize = len(commres.Result)
			d.Result = append(d.Result, commres)

			d.CommSend(conn)
			d.Lock.Unlock()
		} else {
			log.Println("#############stable case#############")
			d.Lock.Lock()

			commres.Result = cmdreslen + strcmdres
			commres.Sendingsize = len(commres.Result)
			d.Result = append(d.Result, commres)

			d.CommSend(conn)
			d.Lock.Unlock()
		}

	default:
		log.Println("#############error case#############")
		log.Println(err)

		cmdreslen = strconv.Itoa(len("Command error : "+err.Error()+"\n")) + "\n"

		d.Lock.Lock()

		commres.Result = cmdreslen + "Command error : " + err.Error()
		commres.Sendingsize = len(commres.Result)
		d.Result = append(d.Result, commres)

		d.CommSend(conn)
		d.Lock.Unlock()
	}
}

// Execute에서 실행한 결과를 Client에게 전송하는 함수
func (d *Data) CommSend(conn net.Conn) {
	if d.Connerr == nil {
		log.Printf("#############Ready for sending to %v!#############", conn.RemoteAddr().String())

		for n, res := range d.Result {
			log.Printf("%d : 번째 %v\n", n, string(res.Result))
			_, err := conn.Write([]byte(res.Result))
			if err != nil {
				log.Printf("write err : %v\n", err)
			}
			d.Result = append(d.Result[:n], d.Result[n+1:]...)
		}
		log.Println("sending complete")
	} else {
		log.Printf("Cannot sending result : %v\n", d.Connerr)
		log.Printf("connection is closed from client : %v\n", conn.RemoteAddr().String())
	}
}
