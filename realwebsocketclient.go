/*
 * @Author: your name
 * @Date: 2022-01-10 17:09:47
 * @LastEditTime: 2022-01-26 17:44:14
 * @LastEditors: Please set LastEditors
 * @Description: 打开koroFileHeader查看配置 进行设置: https://github.com/OBKoro1/koro1FileHeader/wiki/%E9%85%8D%E7%BD%AE
 * @FilePath: /go_code/src/new_main2.go
 */
// client.go
package main
 
import (
    "log"
    "os"
    // "os/signal"
    "time"
    "github.com/gorilla/websocket"
    "fmt"
    "github.com/Jeffail/gabs/v2"
    "net/http"
    "sync"
    "encoding/json"
    "io/ioutil"
    "strings"
    "io"
    "strconv"
)
 
var done chan interface{}
var interrupt chan os.Signal

type Data struct{
  id string `json:"id"`
  timestamp string `json:"timestamp"`
  amount string `json:"amount"`
  str string `json:"str"`
  price string `json:price`
  price_str string `json:price_str`
  Type string `json:type`
  mircrotimestamp string `json:mircrotimestamp`
  buy_order_id string `json:buy_order_id`
  sell_order_id string `json:sell_order_id`
}

type Message struct{
  event string `json:"event"`
  channel string `json:"channel"`
  data string `json:"data"`
}

type Fst struct{
  total int `json:total`
  entries []Entries `json:entries`
  limit int `json:limit`
  offset int `json:offset`
}
type Entries struct{
  tradeID string `json:tradeId`
  makerAsset string `json:makerAsset`
  takerAsset string `json:takerAsset`
  makerAmount string `json:makerAmount`
  takerAmount string `json:takerAmount`
  makerAddress string `json:makeAddress`
  takerAddress string `json:takerAddress`
  senderAddress string `json:senderAddress`
  settleAt int `json:senderAddress`
  txHash string `json:txHash`
}


type fstPostBody struct{
  assetBase string `json:assetBase`
  assetQuote string `json:assetQuote`
  timestamp int64 `json:timestamp`
}


var JsonStringStorage Message;
var fstbody Fst;


func getTime() int64{
  now := time.Now()
  UnixNano := now.UnixNano()
  millisec := UnixNano/1000000
  // t := time.Millisecond
  return millisec
}
// var fstbody Fst;
func req(){
  resp, err := http.Get("https://markets.cryptosx.io/api/v1/external/trade?asset0=FST&asset1=USDT&entryLimit=1&entryOffset=0&orderDESC=settleAt")
  if err != nil{
    fmt.Println("fst" ,err)
  }
  defer resp.Body.Close()
  body, err := io.ReadAll(resp.Body)
  if err != nil{
    fmt.Println("too many request" , err)
  }
  msgString := string(body)
  jsonParsed, err := gabs.ParseJSON([]byte(msgString))

  stringerr := jsonParsed.Search("err").Data()
  if stringerr != nil{
    return 
  }
  if err != nil{
    fmt.Println("too many request" , err)
  }
  // dt := time.Now()
  // var entries []Entries
  taker := jsonParsed.Search("entries","0", "takerAsset").Data().(string)
  takerAmount,err := strconv.ParseFloat(jsonParsed.Search("entries","0","takerAmount").Data().(string),64)
  makerAmount,err := strconv.ParseFloat(jsonParsed.Search("entries","0","makerAmount").Data().(string),64)

  // var price float64
  if(taker == "FST"){
    s := makerAmount/takerAmount
    fmt.Println(s)
  }else{
    s := takerAmount/makerAmount
    fmt.Println(s)
  }
  // mill := getTime()/ 1000 - 60 * 60 * 48
  // fmt.Println(mill)

  // assetBase := "FST"
  // assetQuote := "USDT"
  // timestamp := "1642393208"
  // post := "{\"assetBase\":\"" + assetBase + "\",\"assetQuote\":\"" + assetQuote +"\",\"timestamp\":\"" + timestamp +"\"}"
  // var jsonStr = []byte(post)
  // req, err := http.NewRequest("POST","https://markets.cryptosx.io/api/v1/external/trade?asset0=FST&asset1=USDT&entryLimit=1&entryOffset=0&orderDESC=settleAt", bytes.NewBuffer(jsonStr))
  // req.Header.Set("Content-Type", "application/json")
  
  // client := &http.Client{}
  // resp, err = client.Do(req)
  // if err != nil {
  //     panic(err)
  // }
  // defer resp.Body.Close()
  // body, _ = ioutil.ReadAll(resp.Body)
  // fmt.Println("response Body:", string(body))
}


func fst(){
  for{
    req()
    time.Sleep(time.Duration(5000*time.Millisecond))
  }
}

var wg sync.WaitGroup
// var loop chan bool

var upgrader = websocket.Upgrader{
  ReadBufferSize: 1024,
  WriteBufferSize: 1024,
}

func reader(conn *websocket.Conn){
  for{
    messageType, p , err := conn.ReadMessage()
    fmt.Println(messageType, p)
    if err != nil{
      fmt.Println(err)
      return
    }
    for{
      if err := conn.WriteMessage(websocket.TextMessage , []byte(prices_str)); err != nil{
        fmt.Println(err)
        return
      }
      time.Sleep(time.Duration(5000*time.Millisecond))
    }
  }
}

func wsEndpoint(w http.ResponseWriter, r *http.Request){
  upgrader.CheckOrigin = func(r *http.Request) bool {return true}

  ws, err := upgrader.Upgrade(w,r,nil)
  if err !=nil{
    log.Println(err)
  }
  defer ws.Close()
  log.Println("Client Successfully Connected")
  reader(ws)
}

func setupRoutes(){
  http.HandleFunc("/ws",wsEndpoint)
}


func getOpenPrice(pairs string) (float64, float64) {
  url := fmt.Sprintf("https://www.bitstamp.net/api/v2/ticker/%s/" , pairs)
  // fmt.Println(price)
  resp, err := http.Get(url)
  if err != nil {
  log.Fatalln(err)
  }
  body, err := ioutil.ReadAll(resp.Body)
  if err != nil {
    log.Fatalln(err)
  }
  //Convert the body to type string
  sb := string(body)

  jsonParsed, err := gabs.ParseJSON([]byte(sb))

  var last = stringToFloat64(jsonParsed.Search("last").Data().(string))
  var open = stringToFloat64(jsonParsed.Search("open").Data().(string))

  var percentage = (open-last)/last *100

  // fmt.Println(percentage)
  return open,percentage
}


func stringToFloat64(numInString string)(floatNum float64){
  const bitSize = 64 // Don't think about it to much. It's just 64 bits.
  floatNum, err := strconv.ParseFloat(numInString, bitSize)
  if err != nil{
    fmt.Println(err)
  }
  return floatNum
}


func ConnectionHearing(connections [11]*websocket.Conn ,channels [11]string ,prices [11]float64 ,percentage [11]float64){
  for i := 0; i < 5; i++ {
  err := connections[i].WriteMessage(websocket.TextMessage , []byte(channels[i]))
  if err != nil{
    fmt.Println(err ,"236")
    CloseConnections(connections)
    initConnection()
  }
    _, msg, err := connections[i].ReadMessage()
    if err != nil{
      fmt.Println(err, "240")
      CloseConnections(connections)
      initConnection()
    }
    // log.Println(msg)
    if len(msg) == 0{
      return
    }
    msgString := string(msg)
    jsonParsed, err := gabs.ParseJSON([]byte(msgString))
    if err != nil{
      fmt.Println(msg)
      fmt.Println(err,"245")
    }
    // fmt.Println(msgString)
    strings1 := string(jsonParsed.Search("channel").Data().(string))

    arr1 := strings.Split(strings1,"")

    pairs := arr1[12]+arr1[13]+arr1[14]+arr1[15]+arr1[16]+arr1[17]
    price,percentage := getOpenPrice(pairs)

    prices[i] = price 
    percentages[i] = percentage

    for key,child := range jsonParsed.Search("data").ChildrenMap(){
      if(key == "price"){
        prices[i] = child.Data().(float64)
      }
    }
  }
  for i := 5; i < 11; i++ {
    err := connections[i].WriteMessage(websocket.TextMessage , []byte(channels[i]))
    if err != nil{
      fmt.Println(err , "267")
    }
    _, msg, err := connections[i].ReadMessage()
    if err != nil{
      fmt.Println(err ,"271")
    }
    msgString := string(msg)
    // fmt.Println(msgString)

    jsonParsed, err := gabs.ParseJSON([]byte(msgString))
    if err != nil{
      fmt.Println(err,"278")
    }
    var o = jsonParsed.Search("o").Data().(string)
    jsonParsed2, err := gabs.ParseJSON([]byte(o))
    if err != nil{
      fmt.Println(err,"283")
    }
    lastPrice := jsonParsed2.Search("LastTradedPx")
    percentage := jsonParsed2.Search("Rolling24HrPxChange").Data().(float64)
    percentages[i] = percentage

    prices[i] = lastPrice.Data().(float64)
    if err != nil {
        log.Fatal(err)
    }
  }
  // go fst()
  data := map[string]float64{
		coins[0]:prices[0],
    coins[0]+"percentage":percentages[0],
		coins[1]:prices[1],
    coins[1]+"percentage":percentages[1],
		coins[2]:prices[2],
    coins[2]+"percentage":percentages[2],
		coins[3]:prices[3],
    coins[3]+"percentage":percentages[3],
		coins[4]:prices[4],
    coins[4]+"percentage":percentages[4],
		coins[5]:prices[5],
    coins[5]+"percentage":percentages[5],
		coins[6]:prices[6],
    coins[6]+"percentage":percentages[6],
		coins[7]:prices[7],
    coins[7]+"percentage":percentages[7],
		coins[8]:prices[8],
    coins[8]+"percentage":percentages[8],
		coins[9]:prices[9],
    coins[9]+"percentage":percentages[9],
		coins[10]:prices[10],
    coins[10]+"percentage":percentages[10],
	}
  b, err := json.MarshalIndent(data, "", "")
  if err != nil{
    fmt.Println(err ,"321")
  }
  prices_str = string(b)
}

func CloseConnections(connections [11]*websocket.Conn){
  for i := 0; i < 11; i++ {
    err := connections[i].WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
    if err != nil{
      fmt.Println(err,"330")
    }
  }
}


var prices [11]float64
var percentages [11]float64
var channels [11]string
var prices_str = ""
var connections [11]*websocket.Conn
var coins [11]string


func client(){
  for{
    ConnectionHearing(connections , channels ,prices , percentages)
    time.Sleep(time.Duration(time.Millisecond*1000))
  }
}
func initConnection(){
  connections[0], _, _ = websocket.DefaultDialer.Dial("wss://ws.bitstamp.net", nil)
  connections[1], _, _ = websocket.DefaultDialer.Dial("wss://ws.bitstamp.net", nil)
  connections[2], _, _ = websocket.DefaultDialer.Dial("wss://ws.bitstamp.net", nil)
  connections[3], _, _ = websocket.DefaultDialer.Dial("wss://ws.bitstamp.net", nil)
  connections[4], _, _ = websocket.DefaultDialer.Dial("wss://ws.bitstamp.net", nil)
  connections[5], _, _ = websocket.DefaultDialer.Dial("wss://api.cryptosx.io/WSGateway/", nil)
  connections[6], _, _ = websocket.DefaultDialer.Dial("wss://api.cryptosx.io/WSGateway/", nil)
  connections[7], _, _ = websocket.DefaultDialer.Dial("wss://api.cryptosx.io/WSGateway/", nil)
  connections[8], _, _ = websocket.DefaultDialer.Dial("wss://api.cryptosx.io/WSGateway/", nil)
  connections[9], _, _ = websocket.DefaultDialer.Dial("wss://api.cryptosx.io/WSGateway/", nil)
  connections[10], _, _ = websocket.DefaultDialer.Dial("wss://api.cryptosx.io/WSGateway/", nil) 
}



func main() {

channels[0] = `{"event": "bts:subscribe","data": {"channel": "live_trades_btcusd"}}`;
channels[1] = `{"event": "bts:subscribe","data": {"channel": "live_trades_ethusd"}}`;
channels[2] = `{"event": "bts:subscribe","data": {"channel": "live_trades_bchusd"}}`;
channels[3] = `{"event": "bts:subscribe","data": {"channel": "live_trades_ltcusd"}}`;
channels[4] = `{"event": "bts:subscribe","data": {"channel": "live_trades_xrpusd"}}`;
channels[5] = "{\"m\":0,\"n\":\"GetLevel1\",\"i\":0,\"o\":\"{\\\"OMSId\\\":1,\\\"InstrumentId\\\":24}\"}"
channels[6] = "{\"m\":0,\"n\":\"GetLevel1\",\"i\":0,\"o\":\"{\\\"OMSId\\\":1,\\\"InstrumentId\\\":25}\"}"
channels[7] = "{\"m\":0,\"n\":\"GetLevel1\",\"i\":0,\"o\":\"{\\\"OMSId\\\":1,\\\"InstrumentId\\\":21}\"}"
channels[8] = "{\"m\":0,\"n\":\"GetLevel1\",\"i\":0,\"o\":\"{\\\"OMSId\\\":1,\\\"InstrumentId\\\":26}\"}"
channels[9] = "{\"m\":0,\"n\":\"GetLevel1\",\"i\":0,\"o\":\"{\\\"OMSId\\\":1,\\\"InstrumentId\\\":22}\"}"
channels[10] = "{\"m\":0,\"n\":\"GetLevel1\",\"i\":0,\"o\":\"{\\\"OMSId\\\":1,\\\"InstrumentId\\\":28}\"}"
connections[0], _, _ = websocket.DefaultDialer.Dial("wss://ws.bitstamp.net", nil)
connections[1], _, _ = websocket.DefaultDialer.Dial("wss://ws.bitstamp.net", nil)
connections[2], _, _ = websocket.DefaultDialer.Dial("wss://ws.bitstamp.net", nil)
connections[3], _, _ = websocket.DefaultDialer.Dial("wss://ws.bitstamp.net", nil)
connections[4], _, _ = websocket.DefaultDialer.Dial("wss://ws.bitstamp.net", nil)
connections[5], _, _ = websocket.DefaultDialer.Dial("wss://api.cryptosx.io/WSGateway/", nil)
connections[6], _, _ = websocket.DefaultDialer.Dial("wss://api.cryptosx.io/WSGateway/", nil)
connections[7], _, _ = websocket.DefaultDialer.Dial("wss://api.cryptosx.io/WSGateway/", nil)
connections[8], _, _ = websocket.DefaultDialer.Dial("wss://api.cryptosx.io/WSGateway/", nil)
connections[9], _, _ = websocket.DefaultDialer.Dial("wss://api.cryptosx.io/WSGateway/", nil)
connections[10], _, _ = websocket.DefaultDialer.Dial("wss://api.cryptosx.io/WSGateway/", nil) 
coins[0] = `btcusd`
coins[1] = `ethusd`
coins[2] = `bchusd`
coins[3] = `ltcusd`
coins[4] = `xrpusd`
coins[5] = `Tether`
coins[6] = `RavenCoin`
coins[7] = `CCT`
coins[8] = `WhiskyFund`
coins[9] = `AGWD`
coins[10] = `TIRC`
// go setupRoutes()
http.HandleFunc("/ws",wsEndpoint)
// stringToNum()
go client()
// for{
//   fmt.Println("running")
//   time.Sleep(time.Duration(10000*time.Millisecond))
// }
log.Fatal(http.ListenAndServe(":8090", nil))

}


