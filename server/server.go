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
	Command string
	Totsize int
	Result  []byte
	Lock    sync.Mutex
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
		defer conn.Close()
	}
}

func (d *Data) CommRead(conn net.Conn) {
	log.Printf("serving %s\n", conn.RemoteAddr().String())
	for {
		var recvall []byte
		recvlen := 0
		recvData := make([]byte, 4096)

		recvCommand, err := conn.Read(recvData)
		if err != nil {

			if err == io.EOF {
				log.Printf("connection is closed from client : %v\n", conn.RemoteAddr().String())
				return
			}

			log.Printf("Failed to receive data : %v\n", err)

		} else {

			if recvCommand > 0 {
				recvSize := recvData[:bytes.Index(recvData, []byte("\n"))]
				size, err := strconv.Atoi(string(recvSize))
				if err != nil {
					return
				}
				log.Printf("receive command size : %d\n", size)

				sizelen := len(strconv.Itoa(size))
				log.Printf("sizelen : %d\n", sizelen)

				totsize := size + sizelen
				d.Totsize = totsize
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

								d.Command = string(recvall[:recvlen])
								go d.Execute(conn)

							}

							log.Printf("NOT FINISH LOOP resultlen : %v == totsize : %v\n", recvlen, totsize)
						}
					} else {

						d.Command = string(recvall[:recvlen])
						go d.Execute(conn)

					}
				}
			} else {
				log.Println("No receive data")
				return
			}
		}
	}
}

func (d *Data) Execute(conn net.Conn) {
	// Result := make(chan []byte, 4096)
	// defer close(Result)

	// log.Printf("Result before exex.Command : %v, len of Result : %v\n", Result, len(Result))

	cmd := exec.Command("bash", "-c", d.Command)
	log.Printf("Execute Command : %v\n", cmd)

	cmdres, err := cmd.Output()
	cmdreslen := []byte(strconv.Itoa(len(string(cmdres)+"\n")) + "\n")

	switch err {
	case nil:
		if string(cmdres) == "" {
			cmdreslen = []byte(strconv.Itoa(len("No output data"+"\n")) + "\n")

			d.Lock.Lock()
			d.Result = append(cmdreslen, ([]byte("No output data"))...)
			d.Lock.Unlock()

			go d.CommSend(conn)
			log.Println("close Result!")

		} else {
			log.Println("#############stable case#############")

			d.Lock.Lock()
			d.Result = append(cmdreslen, cmdres...)
			log.Printf("len of totres : %v, totres : %v\n", len(d.Result), string(d.Result))
			d.Lock.Unlock()

			go d.CommSend(conn)
			log.Println("close Result!")
		}
	default:
		log.Println("#############error case#############")
		log.Println(err)

		cmdreslen = []byte(strconv.Itoa(len("Command error : "+err.Error()+"\n")) + "\n")

		d.Lock.Lock()
		d.Result = append(cmdreslen, []byte("Command error : "+err.Error())...)
		d.Lock.Unlock()

		log.Printf("len of totres : %v, totres : %v\n", len(d.Result), string(d.Result))

		go d.CommSend(conn)
		log.Println("close Result!")
	}
}

func (d *Data) CommSend(conn net.Conn) {
	// var sendingres []byte
	// sendingres = append(sendingres, <-Result...)
	log.Println("Ready for sending!")

	d.Lock.Lock()
	_, err := conn.Write(d.Result)
	if err != nil {
		log.Printf("write err : %v\n", err)
		return
	}
	log.Println("sending ")
	d.Lock.Unlock()
}
