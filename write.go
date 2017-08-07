package main

import (
	"fmt"
	"log"
	"math/rand"
	"time"

	"github.com/influxdata/influxdb/client/v2"
)

const (
	url           = "http://192.168.0.22:8086" //"http://localhost:8086"
	MyDB          = "testDB"                   //eb0f38b2
	username      = "paul"
	password      = "123"
	forCount      = 1
	sentenseCount = 100
)

//select * from cba WHERE t1 = '10.1' order by time desc limit 1;
//SELECT * FROM oknum,ngnum,total WHERE uuid='9D4BDB05-0970-45D2-BF0B-AE169F6A35E7' ORDER BY time DESC LIMIT 1;
//SELECT * FROM red,yellow WHERE uuid='uuid_1' ORDER BY time DESC LIMIT 1;
//SELECT * FROM oknum,ngnum WHERE uuid='uuid_1' ORDER BY time DESC LIMIT 1;
var (
	inCount         int64 = 0
	requestCount    int64 = 0
	startTime       int64 = 0
	qTime           time.Time
	selectSentence  string = "select * from cba order by time desc limit 1;" //
	selectSentences string
)

func queryDB(clnt client.Client, cmd string) (res []client.Result, err error) {
	q := client.Query{
		Command:  cmd,
		Database: MyDB,
	}

	if response, err := clnt.Query(q); err == nil {
		if response.Error() != nil {
			return res, response.Error()
		}
		res = response.Results
	} else {
		return res, err
	}
	return res, nil
}

func writePoints(clnt client.Client, num int) {
	sampleSize := 1 * 100
	rand.Seed(42)
	t := num
	bp, _ := client.NewBatchPoints(client.BatchPointsConfig{
		Database:  MyDB,
		Precision: "us",
	})

	for i := 0; i < sampleSize; i++ {
		t += 1
		tags := map[string]string{
			"system_name":    fmt.Sprintf("sys_%d", i%10),
			"site_name":      fmt.Sprintf("s_%d", (t+i)%10),
			"equipment_name": fmt.Sprintf("e_%d", t%10),
		}
		fields := map[string]interface{}{
			"value": fmt.Sprintf("%d", rand.Int()),
		}
		pt, err := client.NewPoint("monitorStatus", tags, fields, time.Now())
		if err != nil {
			log.Fatalln("Error: ", err)
		}
		bp.AddPoint(pt)
	}

	//	start1 := time.Now().Nanosecond()
	err := clnt.Write(bp)
	//	fmt.Printf("timxxe is %d\n", (time.Now().Nanosecond()-start1)/1e6)

	if err != nil {
		log.Fatal(err)
	}

	//fmt.Printf("%d task done\n",num)
}

func ouput(i int) {
	requestCount = requestCount + 1
	inCount = inCount + 1

	c, err := client.NewHTTPClient(client.HTTPConfig{
		Addr:     url,
		Username: username,
		Password: password,
	})

	if err != nil {
		log.Fatalln("Error: ", err)
	}

	writePoints(c, i)

	requestCount = requestCount - 1
	if requestCount == 0 {
		fmt.Printf("requestCount is %d\n", inCount)
	}

}

func query() {
	requestCount = requestCount + 1
	inCount = inCount + 1

	c, err := client.NewHTTPClient(client.HTTPConfig{
		Addr:     url,
		Username: username,
		Password: password,
	})
	if err != nil {
		log.Fatalln("Error: ", err)
	}

	res, err2 := queryDB(c, selectSentences)
	if err2 != nil {
		log.Fatal(err)
	}
	_ = len(res[0].Series[0].Values)
	//		count := len(res[0].Series[0].Values)
	//		log.Printf("Found a total of %v records\n", count)

	requestCount = requestCount - 1
	if requestCount == 0 {
		fmt.Printf("time2 is %15d\n", time.Now().Sub(qTime).Nanoseconds()/1e6)
		fmt.Printf("requestCount is %d\n", inCount)
	}
}

func demo() {
	t1 := time.NewTimer(time.Second * 3)
	//	t2 := time.NewTimer(time.Second * 10)

	for {
		select {

		case <-t1.C:
			println("5s timer:", time.Now().UnixNano())
			inCount = 0
			qTime = time.Now()
			for i := 0; i < forCount; i++ {
				go query()
				//				go ouput(i)
			}
			t1.Reset(time.Second * 1)

			//		case <-t2.C:
			//			println("10s timer")
			//			t2.Reset(time.Second * 10)
		}
	}
}

func main() {
	// Make client
	//	c, err := client.NewHTTPClient(client.HTTPConfig{
	//		Addr:     "http://localhost:8086",
	//		Username: username,
	//		Password: password,
	//	})
	//	if err != nil {
	//		log.Fatalln("Error: ", err)
	//	}
	//	_, err = queryDB(c, fmt.Sprintf("CREATE DATABASE %s", MyDB))
	//	if err != nil {
	//		log.Fatal(err)
	//	}

	for i := 0; i < sentenseCount; i++ {
		selectSentences += selectSentence
	}
	fmt.Printf("task done : i=%s \n", selectSentences)

	i := 1
	start1 := time.Now()
	for i <= forCount {

		c, err := client.NewHTTPClient(client.HTTPConfig{
			Addr:     url,
			Username: username,
			Password: password,
		})
		if err != nil {
			log.Fatalln("Error: ", err)
		}

		//		writePoints(c, i)

		res, err2 := queryDB(c, selectSentences)
		if err2 != nil {
			log.Fatal(err)
		}
		//		_ = len(res[0].Series[0].Values)
		count := len(res[0].Series[0].Values)
		log.Printf("Found a total of %v records\n", count)

		//		fmt.Printf("i=%d\n", i)
		i += 1

	}
	fmt.Printf("time2 is %15d\n", time.Now().Sub(start1).Nanoseconds()/1e6)
	fmt.Printf("task done : i=%d \n", i)

	demo()

}
