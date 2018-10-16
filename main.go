/*
	Copyright 2018 The AfterEther Team
	This file is part of AEX, Ethereum Blockchain Explorer.
		
	AEX is free software: you can redistribute it and/or modify
	it under the terms of the GNU Lesser General Public License as published by
	the Free Software Foundation, either version 3 of the License, or
	(at your option) any later version.
	
	AEX is distributed in the hope that it will be useful,
	but WITHOUT ANY WARRANTY; without even the implied warranty of
	MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
	GNU Lesser General Public License for more details.
	
	You should have received a copy of the GNU Lesser General Public License
	along with AEX. If not, see <http://www.gnu.org/licenses/>.
*/
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
	ErrInvalidParamNum error = errors.New("Invalid number of parameters for this request")
	ErrInvalidAddrList error = errors.New("Address list contains invalid character")

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
	w.Write([]byte(`{"error":"`+err_text+`","result":{}}`));
}
func invalid_address_list(w http.ResponseWriter,address_list string) bool {
	for _,r:=range address_list { // validates string, this is required to prevent SQL injection, since libpq doesn't accept slices as parameter
		if (r>=48 && r<=57) || (r>=97 && r<=102) || (r==44) || (r==39) {
			// character is valid
		} else {
			reject_request_with_error(w,ErrInvalidAddrList.Error())
			return true
		}
	}
	return false
}
func deliver_balances(w http.ResponseWriter, r *http.Request) {

	uri_arr:=strings.Split(r.URL.Path,"/")
	uri_arr_len:=len(uri_arr)
	if (uri_arr_len<3) {
		reject_request_with_error(w,ErrInvalidParamNum.Error())
		return
	}
	address_list:=strings.ToLower(uri_arr[2])
	if invalid_address_list(w,address_list) {
		return
	}

	balances_list,cmd_err:=Get_balances(address_list)
	balances_list_json,err:=json.Marshal(balances_list)
	if (err!=nil) {
		Error.Println(fmt.Sprintf("Error in marshaling balances list data: %v",balances_list))
		os.Exit(2)
	}
	var err_text string
	if (cmd_err!=nil) {
		err_text=cmd_err.Error()
	}
	w.Write([]byte(`{"error":"`+err_text+`","result":`))
	w.Write([]byte(balances_list_json))
	w.Write([]byte(`}`))
}
func deliver_balances_sum(w http.ResponseWriter, r *http.Request) {

	uri_arr:=strings.Split(r.URL.Path,"/")
	uri_arr_len:=len(uri_arr)
	if (uri_arr_len<3) {
		reject_request_with_error(w,ErrInvalidParamNum.Error())
		return
	}
	address_list:=strings.ToLower(uri_arr[2])
	if invalid_address_list(w,address_list) {
		return
	}

	balance_sum:=Get_balances_sum(address_list)
	balance_sum_json,err:=json.Marshal(balance_sum)
	if (err!=nil) {
		Error.Println(fmt.Sprintf("Error in marshaling balances list data: %v",balance_sum))
		os.Exit(2)
	}
	w.Write([]byte(`{"error":"","result":`))
	w.Write([]byte(balance_sum_json))
	w.Write([]byte(`}`))
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
func deliver_eth_aet_prices(w http.ResponseWriter, r *http.Request) {

	prices:=Get_eth_aet_prices()
	prices_json,err:=json.Marshal(prices)
	if (err!=nil) {
		Error.Println(fmt.Sprintf("Error in marshaling eth/aet price data"))
		os.Exit(2)
	}
	var err_text string
	w.Write([]byte(`{"error":"`+err_text+`","result":`))
	w.Write([]byte(prices_json))
	w.Write([]byte(`}`))
}
func deliver_last_block_number(w http.ResponseWriter, r *http.Request) {

	block_num:=get_last_block()
	block_num_json,err:=json.Marshal(block_num)
	if (err!=nil) {
		Error.Println(fmt.Sprintf("Error in marshaling last bock number"))
		os.Exit(2)
	}
	var err_text string
	w.Write([]byte(`{"error":"`+err_text+`","result":`))
	w.Write([]byte(block_num_json))
	w.Write([]byte(`}`))
}
func deliver_search(w http.ResponseWriter, r *http.Request) {

	uri_arr:=strings.Split(r.URL.Path,"/")
	uri_arr_len:=len(uri_arr)
	if (uri_arr_len<3) {
		reject_request_with_error(w,ErrInvalidParamNum.Error())
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
		reject_request_with_error(w,ErrInvalidParamNum.Error())
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
func deliver_account_ft_approvals(w http.ResponseWriter, r *http.Request) {

	uri_arr:=strings.Split(r.URL.Path,"/")
	uri_arr_len:=len(uri_arr)
	if (uri_arr_len<4) {
		reject_request_with_error(w,ErrInvalidParamNum.Error())
		return
	}
	contract_addr:=uri_arr[2]
	account_addr:=uri_arr[3]

	approvals,cmd_err:=Get_account_ft_approvals(contract_addr,account_addr)
	approvals_json,err:=json.Marshal(approvals)
	if (err!=nil) {
		Error.Println(fmt.Sprintf("Error in marshaling approval data: %v",approvals))
		os.Exit(2)
	}
	var err_text string
	if (cmd_err!=nil) {
		err_text=cmd_err.Error()
	}
	w.Write([]byte(`{"error":"`+err_text+`","result":`))
	w.Write([]byte(approvals_json))
	w.Write([]byte(`}`))
}
func deliver_account_fungible_tokens(w http.ResponseWriter, r *http.Request) {

	uri_arr:=strings.Split(r.URL.Path,"/")
	uri_arr_len:=len(uri_arr)
	if (uri_arr_len<5) {
		reject_request_with_error(w,ErrInvalidParamNum.Error())
		return
	}
	address_list:=strings.ToLower(uri_arr[2])
	if invalid_address_list(w,address_list) {
		return
	}
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
	tokens,cmd_err:=Get_account_fungible_tokens(address_list,offset,limit)
	tokens.Address_list=strings.Replace(address_list,`'`,`%27`,-1)
	tokens_json,err:=json.Marshal(tokens)
	if (err!=nil) {
		Error.Println(fmt.Sprintf("Error in marshaling tokens data: %v",tokens))
		os.Exit(2)
	}
	var err_text string
	if (cmd_err!=nil) {
		err_text=cmd_err.Error()
	}
	w.Write([]byte(`{"error":"`+err_text+`","result":`))
	w.Write([]byte(tokens_json))
	w.Write([]byte(`}`))
}
func deliver_fungible_token_sum(w http.ResponseWriter, r *http.Request) {

	uri_arr:=strings.Split(r.URL.Path,"/")
	uri_arr_len:=len(uri_arr)
	if (uri_arr_len<3) {
		reject_request_with_error(w,ErrInvalidParamNum.Error())
		return
	}
	address_list:=strings.ToLower(uri_arr[2])
	if invalid_address_list(w,address_list) {
		return
	}
	tokens,cmd_err:=Get_fungible_token_sum(address_list)
	tokens_json,err:=json.Marshal(tokens)
	if (err!=nil) {
		Error.Println(fmt.Sprintf("Error in marshaling tokens data: %v",tokens))
		os.Exit(2)
	}
	var err_text string
	if (cmd_err!=nil) {
		err_text=cmd_err.Error()
	}
	w.Write([]byte(`{"error":"`+err_text+`","result":`))
	w.Write([]byte(tokens_json))
	w.Write([]byte(`}`))
}
func deliver_account_nonfungible_tokens(w http.ResponseWriter, r *http.Request) {

	uri_arr:=strings.Split(r.URL.Path,"/")
	uri_arr_len:=len(uri_arr)
	if (uri_arr_len<5) {
		reject_request_with_error(w,ErrInvalidParamNum.Error())
		return
	}
	address_list:=strings.ToLower(uri_arr[2])
	if invalid_address_list(w,address_list) {
		return
	}
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
	tokens,cmd_err:=Get_account_nonfungible_tokens(address_list,offset,limit)
	tokens.Address_list=address_list
	tokens_json,err:=json.Marshal(tokens)
	if (err!=nil) {
		Error.Println(fmt.Sprintf("Error in marshaling tokens data: %v",tokens))
		os.Exit(2)
	}
	var err_text string
	if (cmd_err!=nil) {
		err_text=cmd_err.Error()
	}
	w.Write([]byte(`{"error":"`+err_text+`","result":`))
	w.Write([]byte(tokens_json))
	w.Write([]byte(`}`))
}
func deliver_account_value_transfers(w http.ResponseWriter, r *http.Request) {

	uri_arr:=strings.Split(r.URL.Path,"/")
	uri_arr_len:=len(uri_arr)
	if (uri_arr_len<5) {
		reject_request_with_error(w,ErrInvalidParamNum.Error())
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
		reject_request_with_error(w,ErrInvalidParamNum.Error())
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
		reject_request_with_error(w,ErrInvalidParamNum.Error())
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
		reject_request_with_error(w,ErrInvalidParamNum.Error())
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
		reject_request_with_error(w,ErrInvalidParamNum.Error())
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
func deliver_pending_tx_watch(w http.ResponseWriter, r *http.Request) {

	uri_arr:=strings.Split(r.URL.Path,"/")
	uri_arr_len:=len(uri_arr)
	if (uri_arr_len<3) {
		reject_request_with_error(w,ErrInvalidParamNum.Error())
		return
	}

	tx_hash:=uri_arr[2]
	watch_data,cmd_err:=Pending_transaction_watch(tx_hash)
	watch_data_json,err:=json.Marshal(watch_data)
	if (err!=nil) {
		Error.Println(fmt.Sprintf("Error in marshaling transaction data: %v",watch_data))
		os.Exit(2)
	}
	var err_text string
	if (cmd_err!=nil) {
		err_text=cmd_err.Error()
	}
	w.Write([]byte(`{"error":"`+err_text+`","result":`))
	w.Write([]byte(watch_data_json))
	w.Write([]byte(`}`))
}
func deliver_account(w http.ResponseWriter, r *http.Request) {

	uri_arr:=strings.Split(r.URL.Path,"/")
	uri_arr_len:=len(uri_arr)
	if (uri_arr_len<3) {
		reject_request_with_error(w,ErrInvalidParamNum.Error())
		return
	}
	acct_addr:=uri_arr[2]
	account,_,cmd_err:=Get_account(acct_addr)
	account_json,err:=json.Marshal(account)
	if (err!=nil) {
		Error.Println(fmt.Sprintf("Error in marshaling account data: %v",account))
		os.Exit(2)
	}
	var err_text string
	if (cmd_err!=nil) {
		err_text=cmd_err.Error()
	}
	w.Write([]byte(`{"error":"`+err_text+`","result":`))
	w.Write([]byte(account_json))
	w.Write([]byte(`}`))
}
func deliver_account_token_operations(w http.ResponseWriter, r *http.Request) {

	uri_arr:=strings.Split(r.URL.Path,"/")
	uri_arr_len:=len(uri_arr)
	if (uri_arr_len<6) {
		reject_request_with_error(w,ErrInvalidParamNum.Error())
		return
	}
	contract_addr:=uri_arr[2]
	account_addr:=uri_arr[3]
	offset,err:=strconv.Atoi(uri_arr[4])
	if (err!=nil) {
		reject_request_with_error(w,ErrInvalidOffset.Error())
		return
	}
	limit,err:=strconv.Atoi(uri_arr[5])
	if (err!=nil) {
		reject_request_with_error(w,ErrInvalidLimit.Error())
		return
	}
	tokops,cmd_err:=Get_tokops(contract_addr,account_addr,offset,limit)
	tokops_json,err:=json.Marshal(tokops)
	if (err!=nil) {
		Error.Println(fmt.Sprintf("Error in marshaling tokops at deliver_account_token_operations(): %v",tokops))
		os.Exit(2)
	}
	var err_text string
	if (cmd_err!=nil) {
		err_text=cmd_err.Error()
	}
	w.Write([]byte(`{"error":"`+err_text+`","result":`))
	w.Write([]byte(tokops_json))
	w.Write([]byte(`}`))
}
func deliver_account_ftoken_balance(w http.ResponseWriter, r *http.Request) {

	uri_arr:=strings.Split(r.URL.Path,"/")
	uri_arr_len:=len(uri_arr)
	if (uri_arr_len<4) {
		reject_request_with_error(w,ErrInvalidParamNum.Error())
		return
	}
	contract_addr:=uri_arr[2]
	account_addr:=uri_arr[3]
	account_balance,cmd_err:=Get_fungible_token_account_balance(contract_addr,account_addr)
	account_balance_json,err:=json.Marshal(account_balance)
	if (err!=nil) {
		Error.Println(fmt.Sprintf("Error in marshaling token balance at deliver_account_ftoken_balance(): %v",account_balance))
		os.Exit(2)
	}
	var err_text string
	if (cmd_err!=nil) {
		err_text=cmd_err.Error()
	}
	w.Write([]byte(`{"error":"`+err_text+`","result":`))
	w.Write([]byte(account_balance_json))
	w.Write([]byte(`}`))
}
func deliver_account_full_info(w http.ResponseWriter, r *http.Request) {
	var full_info Json_account_full_info_t
	uri_arr:=strings.Split(r.URL.Path,"/")
	uri_arr_len:=len(uri_arr)
	if (uri_arr_len<3) {
		reject_request_with_error(w,ErrInvalidParamNum.Error())
		return
	}
	acct_addr:=uri_arr[2]
	var cmd_err error
	var err error
	full_info.BAccount,_,cmd_err=Get_account(acct_addr)
	Info.Println(fmt.Sprintf("Get_account(%v): err=%v",acct_addr,cmd_err))
	if cmd_err!=nil {
		reject_request_with_error(w,cmd_err.Error())
		return
	}
	Info.Println(fmt.Sprintf("Get_token_account(): err=%v",cmd_err))
	full_info.TAccount,_,cmd_err=Get_token_account(acct_addr)
	if cmd_err!=nil {
		reject_request_with_error(w,cmd_err.Error())
		return
	}
	full_info.Value_transfers,err=Query_value_transfers_by_acct_addr(acct_addr,0,0)
	Info.Println(fmt.Sprintf("Get_value_transfers_by_acct_addr(): err=%v",cmd_err))
	if err!=nil {
		reject_request_with_error(w,err.Error())
		return
	}
	full_info.Transactions,err=Query_transactions_by_acct_addr(acct_addr,0,0)
	Info.Println(fmt.Sprintf("Get_transactions_by_acct_addr(): err=%v",cmd_err))
	if err!=nil {
		reject_request_with_error(w,err.Error())
		return
	}
	full_info_json,err:=json.Marshal(full_info)
	if (err!=nil) {
		Error.Println(fmt.Sprintf("Error in marshaling account full info: %v",full_info))
		os.Exit(2)
	}
	var err_text string
	if (cmd_err!=nil) {
		err_text=cmd_err.Error()
	}
	w.Write([]byte(`{"error":"`+err_text+`","result":`))
	w.Write([]byte(full_info_json))
	w.Write([]byte(`}`))
}
func deliver_token_info(w http.ResponseWriter, r *http.Request) {
	uri_arr:=strings.Split(r.URL.Path,"/")
	uri_arr_len:=len(uri_arr)
	if (uri_arr_len<3) {
		reject_request_with_error(w,ErrInvalidParamNum.Error())
		return
	}
	contract_address:=uri_arr[2]
	var cmd_err error
	var err error
	tok_info,cmd_err:=Get_token_info(contract_address)
	if cmd_err!=nil {
		reject_request_with_error(w,cmd_err.Error())
		return
	}
	tok_info_json,err:=json.Marshal(tok_info)
	if (err!=nil) {
		Error.Println(fmt.Sprintf("Error in marshaling token info: %v",tok_info))
		os.Exit(2)
	}
	var err_text string
	if (cmd_err!=nil) {
		err_text=cmd_err.Error()
	}
	w.Write([]byte(`{"error":"`+err_text+`","result":`))
	w.Write([]byte(tok_info_json))
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
func deliver_token_holders(w http.ResponseWriter, r *http.Request) {

	uri_arr:=strings.Split(r.URL.Path,"/")
	uri_arr_len:=len(uri_arr)
	if (uri_arr_len<5) {
		reject_request_with_error(w,ErrInvalidParamNum.Error())
		return
	}
	contract_addr:=uri_arr[2]
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
	holders,cmd_err:=Get_token_holders(contract_addr,offset,limit)
	holders_json,err:=json.Marshal(holders)
	if (err!=nil) {
		Error.Println(fmt.Sprintf("Error in marshaling token holders in deliver_token_holders(): %v",holders))
		os.Exit(2)
	}
	var err_text string
	if (cmd_err!=nil) {
		err_text=cmd_err.Error()
	}
	w.Write([]byte(`{"error":"`+err_text+`","result":`))
	w.Write([]byte(holders_json))
	w.Write([]byte(`}`))
}
func deliver_token_transfers(w http.ResponseWriter, r *http.Request) {

	uri_arr:=strings.Split(r.URL.Path,"/")
	uri_arr_len:=len(uri_arr)
	if (uri_arr_len<5) {
		reject_request_with_error(w,ErrInvalidParamNum.Error())
		return
	}
	contract_addr:=uri_arr[2]
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
	transfers,cmd_err:=Get_token_transfers(contract_addr,offset,limit)
	transfers_json,err:=json.Marshal(transfers)
	if (err!=nil) {
		Error.Println(fmt.Sprintf("Error in marshaling token transfers deliver_token_transfers(): %v",transfers))
		os.Exit(2)
	}
	var err_text string
	if (cmd_err!=nil) {
		err_text=cmd_err.Error()
	}
	w.Write([]byte(`{"error":"`+err_text+`","result":`))
	w.Write([]byte(transfers_json))
	w.Write([]byte(`}`))
}
func deliver_token_approvals(w http.ResponseWriter, r *http.Request) {

	uri_arr:=strings.Split(r.URL.Path,"/")
	uri_arr_len:=len(uri_arr)
	if (uri_arr_len<5) {
		reject_request_with_error(w,ErrInvalidParamNum.Error())
		return
	}
	contract_addr:=uri_arr[2]
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
	approvals,cmd_err:=Get_token_approvals(contract_addr,offset,limit)
	approvals_json,err:=json.Marshal(approvals)
	if (err!=nil) {
		Error.Println(fmt.Sprintf("Error in marshaling token approvals deliver_token_approvals(): %v",approvals))
		os.Exit(2)
	}
	var err_text string
	if (cmd_err!=nil) {
		err_text=cmd_err.Error()
	}
	w.Write([]byte(`{"error":"`+err_text+`","result":`))
	w.Write([]byte(approvals_json))
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
	http.HandleFunc("/lbn/", deliver_last_block_number)
	http.HandleFunc("/account/", deliver_account)
	http.HandleFunc("/balances/", deliver_balances)
	http.HandleFunc("/balsum/", deliver_balances_sum)
	http.HandleFunc("/ftoks/", deliver_account_fungible_tokens)
	http.HandleFunc("/ftsum/", deliver_fungible_token_sum)
	http.HandleFunc("/aftappr/", deliver_account_ft_approvals)
	http.HandleFunc("/nftoks/", deliver_account_nonfungible_tokens)
	http.HandleFunc("/afinfo/", deliver_account_full_info)
	http.HandleFunc("/tokinfo/", deliver_token_info)
	http.HandleFunc("/atokops/",deliver_account_token_operations)
	http.HandleFunc("/aftokbal/",deliver_account_ftoken_balance)
	http.HandleFunc("/tokhold/",deliver_token_holders)
	http.HandleFunc("/toktransf/",deliver_token_transfers)
	http.HandleFunc("/tokappr/",deliver_token_approvals)
	http.HandleFunc("/eap/",deliver_eth_aet_prices)
	http.HandleFunc("/tx/", deliver_transaction)
	http.HandleFunc("/btx/", deliver_block_transactions)
	http.HandleFunc("/bvt/", deliver_block_value_transfers)
	http.HandleFunc("/atx/", deliver_account_transactions)
	http.HandleFunc("/avt/", deliver_account_value_transfers)
	http.HandleFunc("/tvt/", deliver_transaction_value_transfers)
	http.HandleFunc("/uncles/", deliver_uncles)
	http.HandleFunc("/stats_difficulty", deliver_stats_difficulty)
	http.HandleFunc("/ptxw/", deliver_pending_tx_watch)
	if err := http.ListenAndServe(":8080", nil); err != nil {
		panic(err)
	}
}
