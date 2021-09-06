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

type data struct {
	conn    net.Conn
	sync    sync.Mutex
	connerr error
	result  []*commResult
}
type commResult struct {
	sendingsize int
	result      string
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
		conn, err := server.Accept()
		d := data{}
		d.conn = conn
		if nil != err {
			log.Printf("Accept Error : %v\n", err)
			continue
		}
		go d.CommRead()
		go d.CommSend()
	}
}

//Client로부터 data를 읽어와, linux command를 추출해 Execute function에 전달하는 함수.
func (d *data) CommRead() {
	defer d.conn.Close()
	log.Printf("serving %s\n", d.conn.RemoteAddr().String())

	for {
		var recvall []byte
		recvlen := 0
		recvData := make([]byte, 4096)

		recvCommand, err := d.conn.Read(recvData)
		if err != nil {
			if err == io.EOF {
				log.Printf("Connection is closed from Client : %v\n", d.conn.RemoteAddr().String())
				d.connerr = err
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

					sizeindex := bytes.Index(recvData, []byte("\n")) + 1
					recvall = append(recvall, recvData[sizeindex:recvCommand]...)
					recvlen += recvCommand - sizeindex

					if totsize > len(recvData) {
						a := 0
						for {
							log.Println("Loop started")

							LargeReceive, err := d.conn.Read(recvData)
							if err != nil {
								log.Printf("Read LargeReceive error : %v\n", err)
							}

							log.Printf("Read LargeReceive : %v\n", LargeReceive)

							recvlen += LargeReceive
							recvall = append(recvall, recvData[:LargeReceive]...)
							a += 1
							log.Printf("%d번째 loop\n", a)

							if recvlen == totsize {
								command := string(recvall[:recvlen])
								log.Println(command)
								go d.Execute(command)
							}
							log.Printf("NOT FINISH LOOP resultlen : %v == totsize : %v\n", recvlen, totsize)
						}
					} else {
						command := string(recvall[:recvlen])
						log.Println(command)
						go d.Execute(command)
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
func (d *data) Execute(command string) {
	var commres commResult

	cmd := exec.Command("bash", "-c", command)
	log.Printf("Execute Command : %v\n", cmd)

	cmdres, err := cmd.Output()
	strcmdres := string(cmdres)
	cmdreslen := strconv.Itoa(len(string(cmdres)+"\n")) + "\n"

	switch err {
	case nil:
		if string(cmdres) == "" {
			cmdreslen = strconv.Itoa(len("No output data"+"\n")) + "\n"
			commres.result = cmdreslen + "No output data"
			commres.sendingsize = len(commres.result)
		} else {
			log.Println("#############stable case#############")
			commres.result = cmdreslen + strcmdres
			commres.sendingsize = len(commres.result)
		}

	default:
		log.Println("#############error case#############")
		log.Println(err)
		cmdreslen = strconv.Itoa(len("Command error : "+err.Error()+"\n")) + "\n"
		commres.result = cmdreslen + "Command error : " + err.Error()
		commres.sendingsize = len(commres.result)
	}
	//queue에 data push
	d.SetResult(commres)
}

//execute result를 push 하는 함수
func (d *data) SetResult(commres commResult) {
	d.sync.Lock()
	defer d.sync.Unlock()
	d.result = append(d.result, &commres)
	log.Printf("len of result : %v\n", len(d.result))
}

// sending할 data를 pop 하는 fuction
func (d *data) GetResult() *commResult {
	d.sync.Lock()
	defer d.sync.Unlock()
	l := len(d.result)

	if l == 0 {
		return nil
	}
	//d.Result의 첫번째 값 가져오고, 가져온 data 제거.
	result, results := d.result[0], d.result[1:]
	d.result = results

	return result
}

// Execute에서 실행한 결과를 Client에게 전송하는 함수
func (d *data) CommSend() {
	log.Println("commsend started")
	//connerr 가 존재하는 경우

	log.Printf("#############Ready for sending to %v!#############", d.conn.RemoteAddr().String())
	for {
		//첫번째 data get
		result := d.GetResult()
		if d.connerr != nil {
			log.Printf("Cannot sending result : %v\n", d.connerr)
			log.Printf("connection is closed from client : %v\n", d.conn.RemoteAddr().String())
			return
		}
		if result == nil {
			continue
		} else {
			log.Printf("sending %v\n", result.result)
			_, err := d.conn.Write([]byte(result.result))
			if err != nil {
				log.Printf("Write Error : %v\n", err)
				return
			}
		}
	}
}
