package main
import (
	"net/http"
	"strings"
	"io"
    "io/ioutil"
	"log"
	"os"
	"strconv"
	"fmt"
	"errors"
	"encoding/json"
)
var (
	ErrMissingParameters error = errors.New("Missing parameters")
	ErrBadBlockNum error = errors.New("Bad block number, must be: integer > -1")
	ErrBadTransactionHash error = errors.New("Bad transaction hash specified as parameter")
	ErrNoRows error = errors.New("No rows found")
	ErrInvalidOffset error = errors.New("Invalid offset parameter")
	ErrInvalidLimit error = errors.New("Invalid limit parameter")
	ErrBlockNotFound error = errors.New("Block (by number or hash) not found")

	Trace   *log.Logger
	Info    *log.Logger
	Warning *log.Logger
	Error   *log.Logger
)
func Init(traceHandle io.Writer,infoHandle io.Writer,warningHandle io.Writer,errorHandle io.Writer) {

	Trace = log.New(traceHandle,"TRACE: ",log.Ldate|log.Ltime|log.Lshortfile)
	Info = log.New(infoHandle,"INFO: ",log.Ldate|log.Ltime|log.Lshortfile)
	Warning = log.New(warningHandle,"WARNING: ",log.Ldate|log.Ltime|log.Lshortfile)
	Error = log.New(errorHandle,"ERROR: ",log.Ldate|log.Ltime|log.Lshortfile)
}
func reject_request_with_error(w http.ResponseWriter,err_text string) {
	w.Write([]byte(`{"error:"`+err_text+`","result":{}}`));
}
func deliver_stats_difficulty(w http.ResponseWriter, r *http.Request) {
	stats,cmd_err:=Stats_difficulty()

	stats_json,err:=json.Marshal(stats)
	if (err!=nil) {
		Error.Println(fmt.Sprintf("Error in marshaling value_transfers at deliver_stats_difficulty(): %v",stats))
		os.Exit(2)
	}
	var err_text string
	if (cmd_err!=nil) {
		err_text=cmd_err.Error()
	}
	w.Write([]byte(`{"error":"`+err_text+`","result":`))
	w.Write([]byte(stats_json))
	w.Write([]byte(`}`))
}
func deliver_new_blocks_exist(w http.ResponseWriter, r *http.Request) {

	uri_arr:=strings.Split(r.URL.Path,"/")
	uri_arr_len:=len(uri_arr)
	if (uri_arr_len<3) {
		reject_request_with_error(w,ErrMissingParameters.Error())
		return
	}
	block_num,err:=strconv.Atoi(uri_arr[2])
	if (err!=nil) {
		reject_request_with_error(w,ErrBadBlockNum.Error())
		return
	}
	exist,cmd_err:=New_blocks_exist(block_num)
	exist_json,err:=json.Marshal(exist)
	if (err!=nil) {
		Error.Println(fmt.Sprintf("Error in marshaling new_blocks_exist data: %v",block_num))
		os.Exit(2)
	}
	var err_text string
	if (cmd_err!=nil) {
		err_text=cmd_err.Error()
	}
	w.Write([]byte(`{"error":"`+err_text+`","result":`))
	w.Write([]byte(exist_json))
	w.Write([]byte(`}`))
}
func deliver_uncles(w http.ResponseWriter, r *http.Request) {

	uri_arr:=strings.Split(r.URL.Path,"/")
	uri_arr_len:=len(uri_arr)
	if (uri_arr_len<3) {
		reject_request_with_error(w,ErrMissingParameters.Error())
		return
	}
	block_num,err:=strconv.Atoi(uri_arr[2])
	if (err!=nil) {
		reject_request_with_error(w,ErrBadBlockNum.Error())
		return
	}
	uncles,cmd_err:=Get_uncles(block_num)
	uncles_json,err:=json.Marshal(uncles)
	if (err!=nil) {
		Error.Println(fmt.Sprintf("Error in marshaling uncle data: %v",uncles))
		os.Exit(2)
	}
	var err_text string
	if (cmd_err!=nil) {
		err_text=cmd_err.Error()
	}
	w.Write([]byte(`{"error":"`+err_text+`","result":`))
	w.Write([]byte(uncles_json))
	w.Write([]byte(`}`))
}
func deliver_block_list(w http.ResponseWriter, r *http.Request) {

	uri_arr:=strings.Split(r.URL.Path,"/")
	uri_arr_len:=len(uri_arr)
	if (uri_arr_len<3) {
		reject_request_with_error(w,ErrMissingParameters.Error())
		return
	}
	block_num,err:=strconv.Atoi(uri_arr[2])
	if (err!=nil) {
		reject_request_with_error(w,ErrBadBlockNum.Error())
		return
	}
	block_list,cmd_err:=Get_block_list(block_num)
	block_list_json,err:=json.Marshal(block_list)
	if (err!=nil) {
		Error.Println(fmt.Sprintf("Error in marshaling block list data: %v",block_list))
		os.Exit(2)
	}
	var err_text string
	if (cmd_err!=nil) {
		err_text=cmd_err.Error()
	}
	w.Write([]byte(`{"error":"`+err_text+`","result":`))
	w.Write([]byte(block_list_json))
	w.Write([]byte(`}`))
}
func deliver_search(w http.ResponseWriter, r *http.Request) {

	uri_arr:=strings.Split(r.URL.Path,"/")
	uri_arr_len:=len(uri_arr)
	if (uri_arr_len<3) {
		return
	}
	search_text:=uri_arr[2]
	search_result,cmd_err:=Search(search_text)
	search_result_json,err:=json.Marshal(search_result)
	if (err!=nil) {
		Error.Println(fmt.Sprintf("Error in marshaling transaction data: %v",search_result))
		os.Exit(2)
	}
	var err_text string
	if (cmd_err!=nil) {
		err_text=cmd_err.Error()
	}
	w.Write([]byte(`{"error":"`+err_text+`","result":`))
	w.Write([]byte(search_result_json))
	w.Write([]byte(`}`))
}
func deliver_main_stats(w http.ResponseWriter, r *http.Request) {
	stats,cmd_err:=Main_stats()

	stats_json,err:=json.Marshal(stats)
	if (err!=nil) {
		Error.Println(fmt.Sprintf("Error in marshaling value_transfers at deliver_main_stats(): %v",stats))
		os.Exit(2)
	}
	var err_text string
	if (cmd_err!=nil) {
		err_text=cmd_err.Error()
	}
	w.Write([]byte(`{"error":"`+err_text+`","result":`))
	w.Write([]byte(stats_json))
	w.Write([]byte(`}`))
}
func deliver_transaction_value_transfers(w http.ResponseWriter, r *http.Request) {

	uri_arr:=strings.Split(r.URL.Path,"/")
	uri_arr_len:=len(uri_arr)
	if (uri_arr_len<3) {
		return
	}
	tx_hash:=uri_arr[2]
	value_transfers,cmd_err:=Get_transaction_value_transfers(tx_hash)
	value_transfers_json,err:=json.Marshal(value_transfers)
	if (err!=nil) {
		Error.Println(fmt.Sprintf("Error in marshaling value_transfers at deliver_transaction_value_transfers(): %v",value_transfers))
		os.Exit(2)
	}
	var err_text string
	if (cmd_err!=nil) {
		err_text=cmd_err.Error()
	}
	w.Write([]byte(`{"error":"`+err_text+`","result":`))
	w.Write([]byte(value_transfers_json))
	w.Write([]byte(`}`))
}
func deliver_account_value_transfers(w http.ResponseWriter, r *http.Request) {

	uri_arr:=strings.Split(r.URL.Path,"/")
	uri_arr_len:=len(uri_arr)
	if (uri_arr_len<5) {
		return
	}
	acct_addr:=uri_arr[2]
	offset,err:=strconv.Atoi(uri_arr[3])
	if (err!=nil) {
		reject_request_with_error(w,ErrInvalidOffset.Error())
		return
	}
	limit,err:=strconv.Atoi(uri_arr[4])
	if (err!=nil) {
		reject_request_with_error(w,ErrInvalidLimit.Error())
		return
	}
	value_transfers,cmd_err:=Get_account_value_transfers(acct_addr,offset,limit)
	value_transfers_json,err:=json.Marshal(value_transfers)
	if (err!=nil) {
		Error.Println(fmt.Sprintf("Error in marshaling value_transfers at deliver_account_value_transfers(): %v",value_transfers))
		os.Exit(2)
	}
	var err_text string
	if (cmd_err!=nil) {
		err_text=cmd_err.Error()
	}
	w.Write([]byte(`{"error":"`+err_text+`","result":`))
	w.Write([]byte(value_transfers_json))
	w.Write([]byte(`}`))
}
func deliver_account_transactions(w http.ResponseWriter, r *http.Request) {

	uri_arr:=strings.Split(r.URL.Path,"/")
	uri_arr_len:=len(uri_arr)
	if (uri_arr_len<5) {
		return
	}
	acct_addr:=uri_arr[2]
	offset,err:=strconv.Atoi(uri_arr[3])
	if (err!=nil) {
		reject_request_with_error(w,ErrInvalidOffset.Error())
		return
	}
	limit,err:=strconv.Atoi(uri_arr[4])
	if (err!=nil) {
		reject_request_with_error(w,ErrInvalidLimit.Error())
		return
	}
	transactions,cmd_err:=Get_account_transactions(acct_addr,offset,limit)
	transactions_json,err:=json.Marshal(transactions)
	if (err!=nil) {
		Error.Println(fmt.Sprintf("Error in marshaling transaction data: %v",transactions))
		os.Exit(2)
	}
	var err_text string
	if (cmd_err!=nil) {
		err_text=cmd_err.Error()
	}
	w.Write([]byte(`{"error":"`+err_text+`","result":`))
	w.Write([]byte(transactions_json))
	w.Write([]byte(`}`))
}
func deliver_block_value_transfers(w http.ResponseWriter, r *http.Request) {


	uri_arr:=strings.Split(r.URL.Path,"/")
	uri_arr_len:=len(uri_arr)
	if (uri_arr_len<3) {
		return
	}
	block_num,err:=strconv.Atoi(uri_arr[2])
	if (err!=nil) {
		reject_request_with_error(w,ErrBadBlockNum.Error())
		return;
	}
	value_transfers,cmd_err:=Get_block_value_transfers(block_num)
	value_transfers_json,err:=json.Marshal(value_transfers)
	if (err!=nil) {
		Error.Println(fmt.Sprintf("Error in marshaling value_transfers of the block: %v",value_transfers))
		os.Exit(2)
	}
	var err_text string
	if (cmd_err!=nil) {
		err_text=cmd_err.Error()
	}
	w.Write([]byte(`{"error":"`+err_text+`","result":`))
	w.Write([]byte(value_transfers_json))
	w.Write([]byte(`}`))
}
func deliver_block_transactions(w http.ResponseWriter, r *http.Request) {

	uri_arr:=strings.Split(r.URL.Path,"/")
	uri_arr_len:=len(uri_arr)
	if (uri_arr_len<3) {
		return
	}
	block_num,err:=strconv.Atoi(uri_arr[2])
	if (err!=nil) {
		reject_request_with_error(w,ErrBadBlockNum.Error())
		return;
	}
	transactions,cmd_err:=Get_block_transactions(block_num)
	transactions_json,err:=json.Marshal(transactions)
	if (err!=nil) {
		Error.Println(fmt.Sprintf("Error in marshaling transactions of the block: %v",transactions))
		os.Exit(2)
	}
	var err_text string
	if (cmd_err!=nil) {
		err_text=cmd_err.Error()
	}
	w.Write([]byte(`{"error":"`+err_text+`","result":`))
	w.Write([]byte(transactions_json))
	w.Write([]byte(`}`))
}
func deliver_transaction(w http.ResponseWriter, r *http.Request) {

	uri_arr:=strings.Split(r.URL.Path,"/")
	uri_arr_len:=len(uri_arr)
	if (uri_arr_len<3) {
		return
	}

	tx_hash:=uri_arr[2]
	tx_data,cmd_err:=Get_transaction(tx_hash)
	tx_data_json,err:=json.Marshal(tx_data)
	if (err!=nil) {
		Error.Println(fmt.Sprintf("Error in marshaling transaction data: %v",tx_data))
		os.Exit(2)
	}
	var err_text string
	if (cmd_err!=nil) {
		err_text=cmd_err.Error()
	}
	w.Write([]byte(`{"error":"`+err_text+`","result":`))
	w.Write([]byte(tx_data_json))
	w.Write([]byte(`}`))
}
func deliver_block(w http.ResponseWriter, r *http.Request) {
	var block_data Json_block_t
	var found bool
	var err error
	var cmd_err error
	uri_arr:=strings.Split(r.URL.Path,"/")
	uri_arr_len:=len(uri_arr)
	if (uri_arr_len<3) {
		reject_request_with_error(w,ErrMissingParameters.Error())
		return
	}
	block_num,err:=strconv.Atoi(uri_arr[2])
	if (err!=nil) {
		block_data,found,cmd_err=Get_block(-1,uri_arr[2])
	} else {
		block_data,found,cmd_err=Get_block(block_num,"")
	}
	if !found {
		reject_request_with_error(w,ErrBlockNotFound.Error())
	}
	block_data_json,err:=json.Marshal(block_data)
	if (err!=nil) {
		Error.Println(fmt.Sprintf("Error in marshaling block data: %v",block_data))
		os.Exit(2)
	}
	var err_text string
	if (cmd_err!=nil) {
		err_text=cmd_err.Error()
	}
	w.Write([]byte(`{"error":"`+err_text+`","result":`))
	w.Write([]byte(block_data_json))
	w.Write([]byte(`}`))
}
func main() {

	Init(ioutil.Discard, os.Stdout, os.Stdout, os.Stderr)
	init_postgres()
	http.HandleFunc("/nbe/", deliver_new_blocks_exist)
	http.HandleFunc("/blist/", deliver_block_list)
	http.HandleFunc("/search/", deliver_search)
	http.HandleFunc("/mainstats", deliver_main_stats)
	http.HandleFunc("/mainstats/", deliver_main_stats)
	http.HandleFunc("/block/", deliver_block)
	http.HandleFunc("/tx/", deliver_transaction)
	http.HandleFunc("/btx/", deliver_block_transactions)
	http.HandleFunc("/bvt/", deliver_block_value_transfers)
	http.HandleFunc("/atx/", deliver_account_transactions)
	http.HandleFunc("/avt/", deliver_account_value_transfers)
	http.HandleFunc("/tvt/", deliver_transaction_value_transfers)
	http.HandleFunc("/uncles/", deliver_uncles)
	http.HandleFunc("/stats_difficulty", deliver_stats_difficulty)
	if err := http.ListenAndServe(":8080", nil); err != nil {
		panic(err)
	}
}
