package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"strings"
	"time"

	fixedwidth "github.com/ianlopshire/go-fixedwidth"

	"github.com/egorka-gh/zbazar/zsync/pkg/endpoint"
	"github.com/egorka-gh/zbazar/zsync/pkg/service"
)

func main() {
	// listen to incoming udp packets
	pc, err := net.ListenPacket("udp", ":21845")
	if err != nil {
		log.Fatal(err)
	}
	defer pc.Close()

	for {
		buf := make([]byte, 1024)
		n, addr, err := pc.ReadFrom(buf)
		if err != nil {
			continue
		}
		go serve(pc, addr, buf[:n])
	}

}

func serve(pc net.PacketConn, addr net.Addr, buf []byte) {
	m := parseMsg(buf)
	if m.Mode == 35 { //|| m.Mode == 31
		//fmt.Println("Get from ", addr.String(), ". packet:", string(buf))
		//fmt.Println(m)
		srvURL := "http://127.0.0.1:8091"
		src := "s1"
		if m.Number > 15 {
			src = "s2"
			srvURL = "http://127.0.0.1:8092"
		}
		dt := time.Date(2000+m.Year, time.Month(m.Month), m.Day, m.Hour, m.Min, m.Sec, 0, time.Now().Location())
		var k float32 = 1

		if strings.Trim(m.CardNumber, " ") == "01230001700044" && m.Sec%2 == 1 {
			//random +/- k
			k = -1
		}
		a := service.Activity{
			Source:    src, //not need it
			Doc:       fmt.Sprintf("%v.%v", m.Number, m.CKNumber),
			DocDate:   dt.Format("2006-01-02 15:04:05"),
			Card:      strings.Trim(m.CardNumber, " "),
			DocSum:    m.Sum * k,
			BonuceSum: m.Sum * k,
		}
		buff, err := json.Marshal(endpoint.AddActivityRequest{Activity: a})
		if err != nil {
			fmt.Println("Err ", err)
			return
		}
		//fmt.Println("Request", string(buff))
		req, err := http.NewRequest("POST", srvURL+"/add-activity", bytes.NewReader(buff))
		if err != nil {
			fmt.Println("Err ", err)
			return
		}
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			fmt.Println("Err ", err)
			return
		}
		_, err = ioutil.ReadAll(resp.Body)
		if err != nil {
			fmt.Println("Err ", err)
			return
		}
		//fmt.Println("Responce ", string(body))

	}

}

type message struct {
	Prefix string `fixed:"1,3"`   //3 array[0..2] of char Символьный Префикс объекта (ККМ)
	Number int    `fixed:"4,9"`   //6  array[0..5] of char Целый числовой Номер ККМ
	Mode   int    `fixed:"10,13"` //4 array[0..3] of char Целый числовой Код события
	//CassirItem string `fixed:"14,33"`   //20 array[0..19] of char Символьный Код кассира (таб.		номер)
	//Cassir     string `fixed:"34,63"`   //30 array[0..29] of char Символьный Имя кассира
	CKNumber int `fixed:"64,73"` //10 array[0..9] of char Целый числовой Номер чека
	//Count      string `fixed:"74,76"`   //3 array[0..2] of char Целый числовой Номер позиции в		чеке
	//BarCode    string `fixed:"77,89"`   //13 array[0..12] of char Символьный Штрих-код		товара
	//GoodsItem  string `fixed:"90,119"`  //30 array[0..29] of char Символьный Код товара
	//GoodsName  string `fixed:"120,149"` //30 array[0..29] of char Символьный Наименование		товара
	//GoodsPrice string `fixed:"150,164"` //15 array[0..14] of char Числовой Цена товара
	//GoodsQuant string `fixed:"165,179"` //15 array[0..14] of char Числовой Количество		товара
	//GoodsSum   string `fixed:"180,194"` //15 array[0..14] of char Числовой Сумма по товарной позиции
	Sum        float32 `fixed:"195,209"` //15 array[0..14] of char Числовой Сумма по чеку
	CardType   string  `fixed:"210,212"` //3 array[0..2] of char Символьный Тип карты(дисконтная,кредитная)
	CardNumber string  `fixed:"213,232"` //20 array[0..19] of char Символьный Номер карты
	//DiscStr    string `fixed:"233,247"` //15 array[0..14] of char Числовой Скидка по строкечека
	//DiscSum    string `fixed:"248,262"` //15 array[0..14] of char Числовой Скидка по чеку
	Day    int `fixed:"263,264"` //2 array[0..1] of char Числовой День
	Month  int `fixed:"265,266"` //2 array[0..1] of char Числовой Месяц
	Year   int `fixed:"267,268"` //2 array[0..1] of char Числовой Год
	Sec100 int `fixed:"269,271"` //3 array[0..2] of char Числовой Милисекунды
	Sec    int `fixed:"272,273"` //2 array[0..1] of char Числовой Секунды
	Min    int `fixed:"274,275"` //2 array[0..1] of char Числовой Минуты
	Hour   int `fixed:"276,277"` //2 array[0..1] of char Числовой Часы
}

/**/
func parseMsg(buf []byte) (m message) {

	err := fixedwidth.Unmarshal(buf, &m)
	if err != nil {
		log.Fatal(err)
	}
	return
}

/**/
