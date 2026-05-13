package mcping

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"os"
	"strconv"
)

type ServerListReturn struct {
	Version struct {
		Name     string `json:"name"`
		Protocol int    `json:"protocol"`
	} `json:"version"`
	Players struct {
		Max    int `json:"max"`
		Online int `json:"online"`
		Sample []struct {
			Name string `json:"name"`
			Id   string `json:"id"`
		} `json:"sample"`
	}
	Description string `json:"description"`
	Favion      string `json:"favicon"`
}

func main() {
	client, err := NewClient("8.148.66.53", 40100)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
	client.Hello()
	Data := client.Recvive()
	fmt.Printf("%v\n", Data.Data)
}

type Client struct {
	conn net.Conn
	host string
	port int
	buff []byte
}

type McData struct {
	Length   int64
	PacketID uint64
	Data     ServerListReturn
}

func NewClient(host string, port int) (Client, error) {
	address := net.JoinHostPort(host, strconv.Itoa(port))
	conn, e := net.Dial("tcp", address)
	var result = Client{
		conn: conn,
		host: host,
		port: port,
		buff: make([]byte, 1024),
	}
	return result, e
}

func (client *Client) Hello() {
	buffer := bytes.NewBuffer(make([]byte, 0))
	//buff := make([]byte, 5)
	buffer.Write(client.PutVarint(-1))
	buffer.Write(String(client.host))
	binary.Write(buffer, binary.BigEndian, uint16(client.port))
	//buffer.Write(buff[:binary.PutUvarint(buff, 1)])
	buffer.Write(client.PutVarint(1))
	client.Send_Packet(0x0, buffer.Bytes())
	client.Write([]byte{0x00, 0x00})
}

func (client *Client) PutVarint(i int64) []byte {
	return client.buff[:binary.PutVarint(client.buff, i)]
}

func PutVarint(i int64) []byte {
	buff := make([]byte, 9)
	return buff[:binary.PutVarint(buff, i)]
}

func (client *Client) PutUVarint(i uint64) []byte {
	return client.buff[:binary.PutUvarint(client.buff, i)]
}

func PutUVarint(i uint64) []byte {
	buff := make([]byte, 9)
	return buff[:binary.PutUvarint(buff, i)]
}

func (client *Client) Write(bit []byte) {
	client.conn.Write(bit)
}

func String(str string) []byte {
	l := len(str)
	buff := make([]byte, l+8)
	buff = append(buff, PutUVarint(uint64(l))...)
	buff = append(buff, []byte(str)...)
	return buff
}

func (client *Client) Send_Packet(packet_id uint64, packet []byte) {
	buffer := bytes.NewBuffer([]byte{})
	Len := make([]byte, 5)
	id := make([]byte, 5)
	id_len := binary.PutUvarint(id, packet_id)
	length := binary.PutUvarint(Len, uint64(len(packet)+id_len))
	buffer.Write(Len[:length])
	buffer.Write(id[:id_len])
	buffer.Write(packet)
	ret := buffer.Bytes()
	client.conn.Write(ret)
	client.conn.Write([]byte{0x01, 0x00})

}

func (client *Client) Recvive() McData {
	data := McData{}
	reader := bufio.NewReader(client.conn)
	data.Length = Read_Varint(reader)
	data.PacketID = uint64(Read_Varint(reader))
	_ = Read_Varint(reader)
	buf := make([]byte, 32767*4+3)
	n, _ := reader.Read(buf)
	json.Unmarshal(buf[:n], &data.Data)
	return data
}

func Read_Varint(r io.ByteReader) int64 {
	var ret int64 = 0
	number := 0
	for {
		if byt, _ := r.ReadByte(); byt&0x80 != 0 {
			value := byt & 0x7F
			ret = ret | int64(value<<(7*number))
			number += 1
		} else {
			value := byt & 0x7F
			ret = ret | int64(value<<(7*number))
			return ret
		}
	}
}
