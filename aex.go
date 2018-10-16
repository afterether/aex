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
	"fmt"
	"strconv"
	"os"
	"math/big"
	"strings"
	"net"
	"errors"

	"database/sql"
	_ "github.com/lib/pq"
)
var default_limit int=20
var db *sql.DB
func init_postgres() {
	var err error
	Info.Println(fmt.Sprintf("Connecting to PostgreSQL database: %v@%v/%v",os.Getenv("ETHBOT_USERNAME"),os.Getenv("ETHBOT_HOST"),os.Getenv("ETHBOT_DATABASE")))
	host,port,err:=net.SplitHostPort(os.Getenv("ETHBOT_HOST"))
	if (err!=nil) {
		host=os.Getenv("ETHBOT_HOST")
		port="5432"
	}
	conn_str:="user='"+os.Getenv("ETHBOT_USERNAME")+"' dbname='"+os.Getenv("ETHBOT_DATABASE")+"' password='"+os.Getenv("ETHBOT_PASSWORD")+"' host='"+host+"' port='"+port+"'";
	db,err=sql.Open("postgres",conn_str);
	if (err!=nil) {
		Error.Println("Can't connect to PostgreSQL database. Check that you have set ETHBOT_USERNAME,ETHBOT_PASSWORD,ETHBOT_DATABASE and ETHBOT_HOST environment variables");
	} else {
	}
	row := db.QueryRow("SELECT now()")
	var now string
	err=row.Scan(&now);
	if (err!=nil) {
		Error.Println("Can't connect to PostgreSQL database. Check that you have set ETHBOT_USERNAME,ETHBOT_PASSWORD,ETHBOT_DATABASE and ETHBOT_HOST environment variables");
		Error.Println(fmt.Sprintf("error: %v",err));
		os.Exit(2)
	} else {
		Info.Println("Connected to Postgres successfuly");
	}
	block_num:=get_last_block();
	if (block_num==-2) {
		Error.Println("can't get block_num from `last_block` table")
		os.Exit(2)
	} else {
		Info.Println(fmt.Sprintf("Last block is: %v",block_num))
	}
}
func Get_balances(address_list string) ([]Json_balance_t,error) {
	var output_array []Json_balance_t
	var query string
	query=`SELECT address,last_balance FROM account WHERE address IN(`+address_list+`)`

	rows,err:=db.Query(query);
	if (err!=nil) {
		Error.Println(fmt.Sprintf("failed to execute query: %v, address list: %v, error=%v",query,address_list,err))
		return output_array,errors.New("Bad query for getting balances:"+err.Error())
	}
	defer rows.Close()
	for rows.Next() {
		var b Json_balance_t
		err:=rows.Scan(&b.Address,&b.Balance)
		if (err!=nil) {
			if err==sql.ErrNoRows {
				return output_array,ErrNoRows;
			} else {
				Error.Println(fmt.Sprintf("Scan failed at Get_balances(): %v",err))
				os.Exit(2)
			}
		}
		output_array=append(output_array,b)
	}
	return output_array,nil
}
func Get_balances_sum(address_list string) string {
	var sum string
	var query string
	query=`SELECT sum(last_balance) as sum FROM account WHERE address IN(`+address_list+`)`

	row:=db.QueryRow(query);
	var aux_sum_str sql.NullString
	err:=row.Scan(&aux_sum_str)
	if (err!=nil) {
		Error.Println(fmt.Sprintf("failed to execute query: %v, address list: %v, error=%v",query,address_list,err))
		os.Exit(2)
	}
	if aux_sum_str.Valid {
		sum=aux_sum_str.String
	}
	return sum
}
func Get_block(block_num int,hash string) (Json_block_t,bool,error) {
	var output Json_block_t
	var query string

	var where_condition string = "block_num=$1"
	query=`
		SELECT 
			b.block_num,
			b.block_hash,
			b.block_ts,
			m.address,
			b.num_tx,
			b.num_vt,
			b.num_uncles,
			b.difficulty,
			b.total_dif,
			b.gas_used,
			b.gas_limit,
			b.size,
			b.nonce,
			b.parent_id,
			b.uncle_hash,
			b.extra,
			b.val_transferred,
			b.miner_reward
		FROM block AS b
		LEFT JOIN account AS m ON b.miner_id=m.account_id
		WHERE `
	var row *sql.Row
	if (block_num==-1) {
		where_condition="block_hash=$1"
		row=db.QueryRow(query+where_condition,hash);
	} else {
		where_condition="block_num=$1"
		row=db.QueryRow(query+where_condition,block_num);
	}
	var parent_id int64
	err:=row.Scan(&output.Number,&output.Hash,&output.Timestamp,&output.Miner,&output.Num_transactions,&output.Num_value_transfers,&output.Num_uncles,&output.Difficulty,&output.Total_difficulty,&output.Gas_used,&output.Gas_limit,&output.Size,&output.Nonce,&parent_id,&output.Sha3uncles,&output.Extra_data,&output.Val_transferred,&output.Miner_reward)
	if (err!=nil) {
		if (err==sql.ErrNoRows) {
			return output,false,errors.New(fmt.Sprintf("Block number %v not found",block_num))
		} else {
			return output,false,errors.New(fmt.Sprintf("SQL error: %v",err))
		}
	}
	last_block_num:=get_last_block()
	if (last_block_num<0) {
		return output,true,errors.New(fmt.Sprintf("Probably the database is empty, last block number is < 0"))
	}
	output.Confirmations=last_block_num-int(output.Number)
	if output.Number>0 {
		parent_num:=output.Number-1
		query="SELECT block_num,block_hash FROM block WHERE block_num=$1"
		row=db.QueryRow(query,parent_num)
		if (err==sql.ErrNoRows) {
			return output,true,errors.New(fmt.Sprintf("Parent block (id=%v) not found",parent_num))
		} else {
			err:=row.Scan(&output.Parent_num,&output.Parent_hash)
			if err!=nil {
				Error.Println(fmt.Sprintf("Failed getting parent block data for block_num=%v: %v",parent_num,err))
			}
		}
	}
	return output,true,nil
}
func Get_block_transactions(block_num int) ([]Json_transaction_t,error) {
	var output_array []Json_transaction_t

	var query string
	query=`
		SELECT 
			tx.tx_id,
			tx.from_id,
			tx.to_id,
			src.address AS from_addr,
			dst.address AS to_addr,
			tx.tx_value::text,
			tx.tx_hash,
			tx.gas_limit,
			tx.gas_used,
			tx.gas_price,
			tx.nonce::text,
			tx.block_num,
			tx.tx_index,
			tx.tx_status,
			tx.v,
			tx.r,
			tx.s,
			tx.tx_error,
			tx.vm_error
		FROM transaction AS tx
		LEFT JOIN account as src ON tx.from_id=src.account_id
		LEFT JOIN account as dst ON tx.to_id=dst.account_id
		WHERE block_num=$1 
		ORDER BY tx.tx_index
		`
	rows,err:=db.Query(query,block_num);
	if (err!=nil) {
		Error.Println(fmt.Sprintf("failed to execute query: %v, block_num=%v, error=%v",query,block_num,err))
		os.Exit(2)
	}
	defer rows.Close()
	for rows.Next() {
		var tx Json_transaction_t
		err:=rows.Scan(&tx.Tx_id,&tx.From_id,&tx.To_id,&tx.From_addr,&tx.To_addr,&tx.Value,&tx.Tx_hash,&tx.Gas_limit,&tx.Gas_used,&tx.Gas_price,&tx.Nonce,&tx.Block_num,&tx.Tx_index,&tx.Tx_status,&tx.V,&tx.R,&tx.S,&tx.Tx_error,&tx.Vm_error)
		if (err!=nil) {
			if err==sql.ErrNoRows {
				return output_array,ErrNoRows;
			} else {
				Error.Println(fmt.Sprintf("Scan failed at Getblocktransactions(): %v",err))
				os.Exit(2)
			}
		}
		output_array=append(output_array,tx)
	}
	return output_array,nil
}
func Get_block_value_transfers(block_num int) ([]Json_value_transfer_t,error) {
	var output_array []Json_value_transfer_t
	var query string
	query=`
		SELECT 
			v.valtr_id,
			v.tx_id,
			t.tx_hash::text,
			t.tx_index,
			v.block_num,
			v.from_id,
			v.to_id,
			src.address AS from_addr,
			dst.address AS to_addr,
			v.value::text,
			v.from_balance::text,
			v.to_balance::text,
			v.kind,
			v.error
		FROM value_transfer AS v
		LEFT JOIN account as src ON v.from_id=src.account_id
		LEFT JOIN account as dst ON v.to_id=dst.account_id
		LEFT JOIN transaction as t ON v.tx_id=t.tx_id
		WHERE v.block_num=$1 
		ORDER BY bnumvt
		`
	rows,err:=db.Query(query,block_num);
	if (err!=nil) {
		Error.Println(fmt.Sprintf("failed to execute query: %v, block_num=%v, error=%v",query,block_num,err))
		os.Exit(2)
	}
	defer rows.Close()
	for rows.Next() {
		var vt Json_value_transfer_t
		var tx_id sql.NullInt64
		var tx_hash sql.NullString
		var tx_index sql.NullInt64
		err:=rows.Scan(&vt.Valtr_id,&tx_id,&tx_hash,&tx_index,&vt.Block_num,&vt.To_addr,&vt.To_id,&vt.From_addr,&vt.To_addr,&vt.Value,&vt.From_balance,&vt.To_balance,&vt.Kind,&vt.Error)
		if (err!=nil) {
			if err==sql.ErrNoRows {
				return output_array,nil;
			} else {
				Error.Println(fmt.Sprintf("Scan failed at Getblocktransactions(): %v",err))
				os.Exit(2)
			}
		}
		if (tx_id.Valid) {
			vt.Tx_id=tx_id.Int64
		} else {
			vt.Tx_id=-1
		}
		if (tx_hash.Valid) {
			vt.Tx_hash=tx_hash.String
		} else {
			vt.Tx_hash=""
		}
		if (tx_index.Valid) {
			vt.Tx_index=int(tx_index.Int64)+1
		}
		output_array=append(output_array,vt)
	}
	return output_array,nil
}
func Get_block_token_operations() {


}
func Get_account_fungible_tokens(address_list string,offset int,limit int) (Tok_acct_fungible_holdings_t,error) {
	var output Tok_acct_fungible_holdings_t

	if limit<default_limit {
		limit=default_limit
	}
	output.Holdings=make([]Account_fungible_holding_t,0,8)
	var query string
	query=`
		SELECT 
			t.contract_id,
			c.address AS contract_address,
			t.symbol,
			t.name,
			t.decimals,
			t.total_supply,
			s.sum_amount AS amount
		FROM token AS t
		LEFT JOIN account AS c ON t.contract_id=c.account_id,
		(
			SELECT h.contract_id,Sum(h.amount) AS sum_amount
			FROM ft_hold AS h,tokacct AS a
			WHERE (h.tokacct_id=a.account_id) AND (a.address  IN(`+address_list+`))
			GROUP BY h.contract_id
		) AS s
		WHERE s.contract_id=t.contract_id
		LIMIT $1
		OFFSET $2
		`
	rows,err:=db.Query(query,limit,offset);
	if (err!=nil) {
		if err!=sql.ErrNoRows {
			return output,err
		}
	}
	defer rows.Close()
	for rows.Next() {
		var holding Account_fungible_holding_t 
		err:=rows.Scan(
			&holding.Token.Contract_id,
			&holding.Token.Contract_addr,
			&holding.Token.Symbol,
			&holding.Token.Name,
			&holding.Token.Decimals,
			&holding.Token.Total_supply,
			&holding.Value)
		if (err!=nil) {
			if err==sql.ErrNoRows {
				break;
			} else {
				Error.Println(fmt.Sprintf("Scan failed at Getblocktransactions(): %v",err))
				os.Exit(2)
			}
		}
		output.Holdings=append(output.Holdings,holding)
	}
	output.Offset=offset;
	output.Limit=limit;
	return output,nil
}
func Get_fungible_token_sum(address_list string) ([]Json_ftoken_sum_t,error) {
	var output []Json_ftoken_sum_t
	output=make([]Json_ftoken_sum_t,0,8)
	var query string
	query=`
		SELECT 
			c.address AS contract_address,
			t.symbol,
			t.decimals,
			s.sum_amount AS amount
		FROM token AS t
		LEFT JOIN account AS c ON t.contract_id=c.account_id,
		(
				SELECT h.contract_id,Sum(h.amount) AS sum_amount
				FROM ft_hold AS h,tokacct AS a
				WHERE (h.tokacct_id=a.account_id) AND (a.address  IN(`+address_list+`))
				GROUP BY h.contract_id
		) AS s
		WHERE s.contract_id=t.contract_id
		`
	rows,err:=db.Query(query);
	if (err!=nil) {
		if err!=sql.ErrNoRows {
			return output,err
		}
	}
	defer rows.Close()
	for rows.Next() {
		var holding Json_ftoken_sum_t
		err:=rows.Scan(
			&holding.Contract_addr,
			&holding.Symbol,
			&holding.Decimals,
			&holding.Balance)
		if (err!=nil) {
			if err==sql.ErrNoRows {
				break;
			} else {
				Error.Println(fmt.Sprintf("Scan failed at Getblocktransactions(): %v",err))
				os.Exit(2)
			}
		}
		output=append(output,holding)
	}
	return output,nil
}
func Get_account_nonfungible_tokens(acct_addr string,offset int,limit int) (Tok_acct_nonfungible_holdings_t ,error) {
	var output Tok_acct_nonfungible_holdings_t

	if limit<default_limit {
		limit=default_limit
	}
	output.Holdings=make(map[int]*Json_nonfungible_holding_t)

	if (limit==0) {
		limit=default_limit
	}
	var query string
	query=`
		SELECT 
			h.contract_id,
			a.address AS account_address,
			h.token_id
		FROM nft_hold AS h,tokacct AS a
		WHERE h.tokacct_id IN(`+acct_addr+`) AND (h.tokacct_id=tokacct.account_id)
		LIMIT $2
		OFFSET $3
		`
	rows,err:=db.Query(query,limit,offset);
	if err!=nil {
		if err!=sql.ErrNoRows {
			return output,err
		}
	}
	defer rows.Close()
	var IN_str string
	for rows.Next() {
		var token_id,account_address string
		var contract_id int
		err:=rows.Scan(&contract_id,&account_address,&token_id)
		if (err!=nil) {
			if err==sql.ErrNoRows {
				break;
			} else {
				Error.Println(fmt.Sprintf("Scan failed at getting nonfungible token ids: %v",err))
				os.Exit(2)
			}
		}
		map_entry,exists:=output.Holdings[contract_id]
		if exists {
			map_entry.Token_IDs=append(map_entry.Token_IDs,token_id)
		} else {
			map_entry=new(Json_nonfungible_holding_t)
			output.Holdings[contract_id]=map_entry
		}
		if len(IN_str)>0 {
			IN_str=IN_str+","
		}
		IN_str=IN_str+strconv.Itoa(contract_id)
	}
	if len(IN_str)>0 {
		query=`
			SELECT 
				t.contract_id,
				c.address AS contract_address,
				t.symbol,
				t.name,
				t.total_supply 
			FROM token AS t
			LEFT JOIN account AS c ON t.contract_id=c.account_id
			WHERE t.contract_id IN(`+IN_str+`)`
		rows,err=db.Query(query);
		if (err!=nil) {
			Error.Println(fmt.Sprintf("failed to execute query: %v, contract_ids=%v, error=%v",query,IN_str,err))
			os.Exit(2)
		}
		defer rows.Close()
		for rows.Next() {
			var contract_id int
			var contract_address,symbol,name,total_supply string
			err:=rows.Scan(&contract_id,&contract_address,&symbol,&name,&total_supply)
			if (err!=nil) {
				if err==sql.ErrNoRows {
					break;
				} else {
					return output,err
				}
			}
			map_entry,exists:=output.Holdings[contract_id]
			if exists {
				map_entry.Token.Contract_id=contract_id
				map_entry.Token.Contract_addr=contract_address
				map_entry.Token.Symbol=symbol
				map_entry.Token.Name=name
				map_entry.Token.Total_supply=total_supply
			} else {
				Error.Println(fmt.Sprintf("Internal bug: contract_id=%v not found in the token map",contract_id))
				os.Exit(2)
			}
		}
	}
	output.Offset=offset;
	output.Limit=limit;
	return output,nil
}
func Get_account_ft_approvals(contract_addr,account_addr string) (Tok_acct_approvals_t,error) {
	var output Tok_acct_approvals_t
	var query string

	var err error
	output.Account_address=account_addr
	output.Token,err=Get_token_info(contract_addr)
	if err!=nil {
		return output,err
	}
	account_id,err:=lookup_token_account(account_addr)
	if err!=nil {
		return output,errors.New(fmt.Sprintf("Can't find account %v in token accounts",account_addr))
	}
	query=`
		SELECT 
			ap.block_num,
			ap.block_ts,
			ap.value,
			ap.value_consumed,
			ap.value-ap.value_consumed AS value_remaining,
			src.address,
			tx_hash
		FROM approval AS ap,tokacct AS src,transaction AS tx
		WHERE
			ap.contract_id=$1 AND
			ap.to_id=$2 AND
			ap.expired=FALSE AND
			ap.from_id=src.account_id AND
			ap.tx_id=tx.tx_id
		ORDER BY block_num DESC
		`
	rows,err:=db.Query(query,output.Token.Contract_id,account_id)
	if err!=nil {
		if err==sql.ErrNoRows {
			return output,nil
		}
		return output,err
	}
	defer rows.Close()
	output.Approvals=make([]Tok_approval_t,0,4)
	for rows.Next() {
		var approval Tok_approval_t
		approval.To=account_addr
		err=rows.Scan(&approval.Block_num,&approval.Timestamp,&approval.Amount_approved,&approval.Amount_transferred,&approval.Amount_remaining,&approval.From,&approval.Tx_hash)
		if err!=nil {
			Error.Println(fmt.Sprintf("Error at Scan() while retrieving approval: %v",err))
		}
		output.Approvals=append(output.Approvals,approval)
	}
	return output,nil
}
func Query_transactions_by_acct_addr(acct_addr string,offset,limit int) ([]Json_transaction_t,error) {
	var transactions []Json_transaction_t=make([]Json_transaction_t,0,32)
	acct_addr=strings.ToLower(acct_addr)
	account_id,_,_,err:=lookup_account(acct_addr)
	if err!=nil {
		return transactions,err
	}
	var query string
	query=`
		SELECT 
			tx.tx_id,
			tx.from_id,
			tx.to_id,
			src.address AS from_addr,
			dst.address AS to_addr,
			tx.tx_value::text,
			tx.tx_hash,
			tx.gas_limit,
			tx.gas_used,
			tx.gas_price,
			tx.nonce::text,
			tx.block_num,
			tx.tx_index,
			tx.tx_status,
			tx.v,
			tx.r,
			tx.s,
			tx.tx_error
		FROM transaction AS tx
		LEFT JOIN account as src ON tx.from_id=src.account_id
		LEFT JOIN account as dst ON tx.to_id=dst.account_id
		WHERE (
			(from_id=$1) OR
			(to_id=$1) 
		)
		ORDER BY bnumtx DESC
		LIMIT $2
		OFFSET $3
		`
	rows,err:=db.Query(query,account_id,limit,offset);
	if (err!=nil) {
		Error.Println(fmt.Sprintf("failed to execute query: %v, account_id=%v, error=%v",query,account_id,err))
		os.Exit(2)
	}
	defer rows.Close()
	for rows.Next() {
		var tx Json_transaction_t
		err:=rows.Scan(&tx.Tx_id,&tx.From_id,&tx.To_id,&tx.From_addr,&tx.To_addr,&tx.Value,&tx.Tx_hash,&tx.Gas_limit,&tx.Gas_used,&tx.Gas_price,&tx.Nonce,&tx.Block_num,&tx.Tx_index,&tx.Tx_status,&tx.V,&tx.R,&tx.S,&tx.Tx_error)
		if (err!=nil) {
			if err==sql.ErrNoRows {
				break;
			} else {
				Error.Println(fmt.Sprintf("Scan failed at Getblocktransactions(): %v",err))
				os.Exit(2)
			}
		}
		transactions=append(transactions,tx)
	}
	return transactions,err
}
func Get_account_transactions(acct_addr string,offset int,limit int) (Json_tx_set_t,error) {
	var output Json_tx_set_t

	if limit==0 {
		limit=default_limit
	}
	output.Offset=offset;
	output.Limit=limit;
	output.Account,_,_=Get_account(acct_addr)
	var err error
	output.Transactions,err=Query_transactions_by_acct_addr(acct_addr,offset,limit)
	return output,err
}
func Query_value_transfers_by_acct_addr(acct_addr string,offset,limit int) ([]Json_value_transfer_t,error) {
	var value_transfers	[]Json_value_transfer_t=make([]Json_value_transfer_t,0,32)

	account_id,_,_,err:=lookup_account(acct_addr)
	if err!=nil {
		return value_transfers,err
	}
	var query string
	rows_to_get:=offset+limit
	query=`
		SELECT
		v.valtr_id,v.block_num,v.from_id,v.to_id,src.address,dst.address,v.from_balance,v.to_balance,v.value,v.kind,v.tx_id,t.tx_hash,v.error
		FROM (
			(
				SELECT valtr_id,block_num,bnumvt,from_id,to_id,from_balance::text,to_balance::text,value,kind,tx_id,error
				FROM value_transfer
				WHERE from_id=$1
				ORDER BY bnumvt DESC LIMIT $2
			) UNION ALL (
				SELECT valtr_id,block_num,bnumvt,from_id,to_id,from_balance::text,to_balance::text,value,kind,tx_id,error
				FROM value_transfer
				WHERE to_id=$1
				ORDER BY bnumvt DESC LIMIT $2
			)
		) AS v
		LEFT JOIN account AS src ON v.from_id=src.account_id
		LEFT JOIN account AS dst ON v.to_id=dst.account_id
		LEFT JOIN transaction AS t ON v.tx_id=t.tx_id
		ORDER BY v.bnumvt DESC
		LIMIT $3
		OFFSET $4
	`
	dquery:=strings.Replace(query,`$1`,strconv.Itoa(account_id),-1)
	dquery=strings.Replace(dquery,`$2`,strconv.Itoa(rows_to_get),-1)
	dquery=strings.Replace(dquery,`$3`,strconv.Itoa(limit),-1)
	dquery=strings.Replace(dquery,`$4`,strconv.Itoa(offset),-1)
	rows,err:=db.Query(query,account_id,rows_to_get,limit,offset);
	if (err!=nil) {
		Error.Println(fmt.Sprintf("failed to execute query: %v, account_id=%v, error=%v",query,account_id,err))
		os.Exit(2)
	}
	defer rows.Close()
	for rows.Next() {
		var vt Json_value_transfer_t
		var tx_id sql.NullInt64
		var tx_hash sql.NullString
		var from_id sql.NullInt64
		var from_addr sql.NullString
		err:=rows.Scan(&vt.Valtr_id,&vt.Block_num,&from_id,&vt.To_id,&from_addr,&vt.To_addr,&vt.From_balance,&vt.To_balance,&vt.Value,&vt.Kind,&tx_id,&tx_hash,&vt.Error)
		if (err!=nil) {
			if err==sql.ErrNoRows {
				break;
			} else {
				Error.Println(fmt.Sprintf("Scan failed at Getaccountvaluetransfers(): %v",err))
				os.Exit(2)
			}
		}
		if (tx_id.Valid) {
			vt.Tx_id=tx_id.Int64
		} else {
			vt.Tx_id=-1
		}
		if (tx_hash.Valid) {
			vt.Tx_hash=tx_hash.String
		} else {
			vt.Tx_hash=""
		}
		if (from_id.Valid) {
			vt.From_id=from_id.Int64
		} else {
			vt.From_id=-1
		}
		if (from_addr.Valid) {
			vt.From_addr=from_addr.String
		} else {
			vt.From_addr=""
		}
		if vt.From_id==vt.To_id { // selftransfer
			// let the value as it is, i.e. the 0
		} else {
			if vt.From_id==int64(account_id) {
				vt.Direction=-1
			} else if vt.To_id==int64(account_id) {
				vt.Direction=1
			}
		}
		value_transfers=append(value_transfers,vt)
	}
	return value_transfers,err
}
func Get_account_value_transfers(acct_addr string,offset int,limit int) (Json_vt_set_t,error) {
	var output	Json_vt_set_t

	if (limit==0) {
		limit=default_limit
	}
	if strings.ToUpper(acct_addr)=="BLOCKCHAIN" {
		acct_addr="0"
	} else {
		acct_addr=strings.ToLower(acct_addr)
	}
	output.Offset=offset;
	output.Limit=limit;
	output.Account,_,_=Get_account(acct_addr)
	var err error
	output.Value_transfers,err=Query_value_transfers_by_acct_addr(acct_addr,offset,limit)
	return output,err
}
func Get_transaction_value_transfers(transaction_hash string) (Json_vt_set_t,error) {
	var output	Json_vt_set_t

	var query string
	query=`
		SELECT 
			v.valtr_id,
			v.block_num,
			v.from_id,
			v.to_id,
			src.address,
			dst.address,
			v.from_balance::text,
			v.to_balance::text,
			v.value::text,
			v.kind,
			v.tx_id,
			tx.tx_hash,
			v.error
		FROM value_transfer v
			LEFT JOIN account AS src ON v.from_id=src.account_id
			LEFT JOIN account AS dst ON v.to_id=dst.account_id,
			transaction tx
		WHERE (tx.tx_hash=$1) AND (v.tx_id=tx.tx_id)
		ORDER BY v.bnumvt ASC
		`
	rows,err:=db.Query(query,transaction_hash);
	if (err!=nil) {
		Error.Println(fmt.Sprintf("failed to execute query: %v, tx_hash=%vv, error=%v",query,transaction_hash,err))
		os.Exit(2)
	}
	defer rows.Close()
	for rows.Next() {
		var vt Json_value_transfer_t
		var tx_id sql.NullInt64
		var tx_hash sql.NullString
		var from_id sql.NullInt64
		var from_addr sql.NullString
		err:=rows.Scan(&vt.Valtr_id,&vt.Block_num,&from_id,&vt.To_id,&from_addr,&vt.To_addr,&vt.From_balance,&vt.To_balance,&vt.Value,&vt.Kind,&tx_id,&tx_hash,&vt.Error)
		if (err!=nil) {
			if err==sql.ErrNoRows {
				break;
			} else {
				Error.Println(fmt.Sprintf("Scan failed at Gettransactionvaluetransfers(): %v",err))
				os.Exit(2)
			}
		}
		if (tx_id.Valid) {
			vt.Tx_id=tx_id.Int64
		} else {
			vt.Tx_id=-1
		}
		if (tx_hash.Valid) {
			vt.Tx_hash=tx_hash.String
		} else {
			vt.Tx_hash=""
		}
		if (from_id.Valid) {
			vt.From_id=from_id.Int64
		} else {
			vt.From_id=-1
		}
		if (from_addr.Valid) {
			vt.From_addr=from_addr.String
		} else {
			vt.From_addr=""
		}
		output.Value_transfers=append(output.Value_transfers,vt)

	}
	output.Offset=0;
	output.Limit=0;
	return output,nil
}
func Get_token_info(contract_address string) (Token_info_t,error) {
	var output Token_info_t

	output.Search_string=contract_address
	var query string
	query=`
		SELECT contract_id,num_transfers,tk.block_created,tx.tx_hash,symbol,name,decimals,total_supply,non_fungible 
		FROM token AS tk 
		LEFT JOIN transaction AS tx ON tk.created_tx_id=tx.tx_id,
		account AS a
		WHERE tk.contract_id=a.account_id AND a.address=$1
	`
	row:=db.QueryRow(query,contract_address)
	output.Contract_addr=contract_address
	err:=row.Scan(&output.Contract_id,
		&output.Num_transfers,
		&output.Block_created,
		&output.Tx_hash,
		&output.Symbol,
		&output.Name,
		&output.Decimals,
		&output.Total_supply,
		&output.Non_fungible)
	if err!=nil {
		if err==sql.ErrNoRows {
			return output,nil
		} else {
			return output,err
		}
	}
	Info.Println(fmt.Sprintf("Token info: returning contract_id=%v",output.Contract_id))
	return output,err
}
func Get_token_holders(contract_address string,offset int,limit int) (Token_holders_t,error) {
	var output Token_holders_t
	var query string

	if limit<default_limit {
		limit=default_limit
	}
	var err error
	output.Token,err=Get_token_info(contract_address)
	if err!=nil {
		Error.Println(fmt.Sprintf("Error getting token info %v",err))
		os.Exit(2)
	}
	query=`SELECT a.address,h.amount FROM ft_hold AS h LEFT JOIN tokacct as a ON h.tokacct_id=a.account_id WHERE contract_id=$1  ORDER by amount DESC limit $2 OFFSET $3`
	Info.Println(fmt.Sprintf(`SELECT a.address,h.amount FROM ft_hold WHERE contract_id=%v ORDER by amount DESC limit %v OFSSET %v`,output.Token.Contract_id,limit,offset))
	rows,err:=db.Query(query,output.Token.Contract_id,limit,offset)
	if err!=nil {
		Error.Println(fmt.Sprintf("Error at query %v: %v",query,err))
		os.Exit(2)
	}
	defer rows.Close()
	holders:=make([]Tok_acct_holder_t,0,limit)
	for rows.Next() {
		var holder_account Tok_acct_holder_t
		err=rows.Scan(&holder_account.Address,&holder_account.Balance)
		if err!=nil {
			return output,err
		}
		holders=append(holders,holder_account)
	}
	output.Holders=holders
	output.Offset=offset;
	output.Limit=limit;
	return output,nil

}
func Get_token_transfers(contract_address string,offset int,limit int) (Token_transfers_t,error) {
	var output Token_transfers_t

	if limit<default_limit {
		limit=default_limit
	}
	if offset<0 {
		offset=0;
	}
	var err error
	output.Token,err=Get_token_info(contract_address)
	if err!=nil {
		Error.Println(fmt.Sprintf("Error getting token info %v",err))
		os.Exit(2)
	}
	var query string
	query=`
		SELECT op.tokop_id,tx.tx_hash,op.block_num,op.block_ts,src.address,dst.address,op.value,op.from_balance,op.to_balance,op.kind,op.non_fungible
		FROM tokop AS op
		LEFT JOIN tokacct AS src ON op.from_id=src.account_id
		LEFT JOIN tokacct AS dst ON op.to_id=dst.account_id
		LEFT JOIN transaction AS tx ON op.tx_id=tx.tx_id
		WHERE contract_id=$1
		ORDER BY op.block_num DESC,op.tokop_id DESC
		LIMIT $2
		OFFSET $3
	`
	rows,err:=db.Query(query,output.Token.Contract_id,limit,offset)
	if err!=nil {
		Error.Println(fmt.Sprintf("Error at query %v: %v",query,err))
		os.Exit(2)
	}
	defer rows.Close()
	tokops:=make([]Tok_transfer_t,0,limit)
	for rows.Next() {
		var op Tok_transfer_t
		err=rows.Scan(&op.Tokop_id,&op.Tx_hash,&op.Block_num,&op.Ts_created,&op.From,&op.To,&op.Value,&op.From_balance,&op.To_balance,&op.Kind,&op.Non_fungible)
		if err!=nil {
			return output,err
		}
		tokops=append(tokops,op)
	}
	output.Tokops=tokops
	output.Offset=offset;
	output.Limit=limit;
	return output,nil
}
func Get_token_approvals(contract_address string,offset int,limit int) (Token_approvals_t,error) {
	var output Token_approvals_t

	if limit<default_limit {
		limit=default_limit
	}
	if offset<0 {
		offset=0;
	}
	var err error
	output.Token,err=Get_token_info(contract_address)
	if err!=nil {
		Error.Println(fmt.Sprintf("Error getting token info %v",err))
		os.Exit(2)
	}
	var query string
	query=`
		SELECT tx.tx_hash,ap.block_num,ap.block_ts,src.address,dst.address,ap.value,ap.value_consumed,ap.value-ap.value_consumed AS remaining,ap.expired
		FROM approval AS ap
		LEFT JOIN tokacct AS src ON ap.from_id=src.account_id
		LEFT JOIN tokacct AS dst ON ap.to_id=dst.account_id
		LEFT JOIN transaction AS tx ON ap.tx_id=tx.tx_id
		WHERE contract_id=$1
		ORDER BY ap.block_num DESC,ap.approval_id DESC
		LIMIT $2
		OFFSET $3
	`
	Info.Println(fmt.Sprintf("%v",query))
	Info.Println(fmt.Sprintf("contract_id=%v,offset=%v, limit=%v",output.Token.Contract_id,offset,limit))
	rows,err:=db.Query(query,output.Token.Contract_id,limit,offset)
	if err!=nil {
		Error.Println(fmt.Sprintf("Error at query %v: %v",query,err))
		os.Exit(2)
	}
	defer rows.Close()
	approvals:=make([]Tok_approval_t,0,limit)
	for rows.Next() {
		var ap Tok_approval_t
		err=rows.Scan(&ap.Tx_hash,&ap.Block_num,&ap.Timestamp,&ap.From,&ap.To,&ap.Amount_approved,&ap.Amount_transferred,&ap.Amount_remaining,&ap.Expired)
		if err!=nil {
			return output,err
		}

		approvals=append(approvals,ap)
	}
	output.Approvals=approvals
	output.Offset=offset;
	output.Limit=limit;
	return output,nil
}
func Get_tokops(contract_address,account_address string,offset int,limit int) (Token_transfers_t,error) {
	var output Token_transfers_t

	if limit<default_limit {
		limit=default_limit
	}
	if offset<0 {
		offset=0;
	}
	var err error
	output.Account_address=account_address
	output.Token,err=Get_token_info(contract_address)
	if err!=nil {
		return output,err
	}
	account_id,err:=lookup_token_account(account_address)
	if err!=nil {
		return output,errors.New(fmt.Sprintf("Can't find account %v in token accounts",account_address))
	}
	var query string
	query=`
		SELECT op.tokop_id,tx.tx_hash,op.block_num,op.block_ts,src.address,dst.address,op.value,op.from_balance,op.to_balance,op.kind,op.non_fungible
		FROM (
			(
				SELECT tokop_id,block_num,tx_id,from_id,to_id,value,from_balance,to_balance,kind,block_ts,non_fungible
				FROM tokop 
				WHERE (contract_id=$1) AND (from_id=$2)
			) UNION ALL (
				SELECT tokop_id,block_num,tx_id,from_id,to_id,value,from_balance,to_balance,kind,block_ts,non_fungible
				FROM tokop 
				WHERE (contract_id=$1) AND (to_id=$2)
			)
		) AS op
		LEFT JOIN tokacct AS src ON op.from_id=src.account_id
		LEFT JOIN tokacct AS dst ON op.to_id=dst.account_id
		LEFT JOIN transaction AS tx ON op.tx_id=tx.tx_id
		ORDER BY op.block_num DESC,op.tokop_id DESC
		LIMIT $3
		OFFSET $4
	`
	rows,err:=db.Query(query,output.Token.Contract_id,account_id,limit,offset)
	if err!=nil {
		return output,err
	}
	defer rows.Close()
	tokops:=make([]Tok_transfer_t,0,limit)
	for rows.Next() {
		var op Tok_transfer_t
		err=rows.Scan(&op.Tokop_id,&op.Tx_hash,&op.Block_num,&op.Ts_created,&op.From,&op.To,&op.Value,&op.From_balance,&op.To_balance,&op.Kind,&op.Non_fungible)
		if err!=nil {
			return output,err
		}
		tokops=append(tokops,op)
	}
	output.Tokops=tokops
	output.Offset=offset;
	output.Limit=limit;
	return output,nil
}
func Get_fungible_token_account_balance(contract_address,account_address string) (Json_ft_acct_bal_t,error) {
	var output Json_ft_acct_bal_t

	output.Contract_address=contract_address
	output.Account_address=account_address
	output.Balance="0"
	contract_id,_,_,err:=lookup_account(contract_address)
	if err!=nil {
		return output,err
	}
	account_id,err:=lookup_token_account(account_address)
	if err!=nil {
		return output,err
	}
	var query string
	query=get_fungible_token_balance_query()
	row:=db.QueryRow(query,contract_id,account_id)
	var (
		tmp_tokop_id		int64
		tmp_tx_id			int64
		tmp_contract_id		int64
		tmp_block_num		int64
		tmp_block_ts		int64
		tmp_from_id			int
		tmp_to_id			int
		tmp_from_balance	string
		tmp_to_balance		string
		tmp_value			string
		tmp_kind			byte
	)

	err=row.Scan(&tmp_tokop_id,&tmp_tx_id,&tmp_contract_id,&tmp_block_num,&tmp_block_ts,&tmp_from_id,&tmp_to_id,&tmp_from_balance,&tmp_to_balance,&tmp_value,&tmp_kind)
	if err!=nil {
		if err==sql.ErrNoRows {
			err=nil
		}
		return output,err
	}
	if tmp_to_id==tmp_from_id {// selftransfer
		output.Balance=tmp_to_balance
	} else {
		if tmp_to_id==account_id {
			output.Balance=tmp_to_balance
		} 
		if tmp_from_id==account_id {
			output.Balance=tmp_from_balance
		}
	}
	return output,nil
}
func Main_stats() (Json_main_stats_t,error) {
	var out Json_main_stats_t

	var query string
	query=`SELECT 
				round(hash_rate)::text,
				round(block_time,2)::text,
				round(tx_per_block,1)::text,
				round(tx_per_sec,1)::text,
				gas_price::text,
				tx_cost::text,
				round(supply/1000000000000000000)::text,
				round(difficulty)::text,
				round(volume,2),
				round(activity),
				last_block::text 
				FROM mainstats`
	row := db.QueryRow(query)
	err:=row.Scan(  &out.Hash_rate,
					&out.Block_time,
					&out.Tx_per_block,
					&out.Tx_per_sec,
					&out.Gas_price,
					&out.Tx_cost,
					&out.Supply,
					&out.Difficulty,
					&out.Volume,
					&out.Activity,
					&out.Last_block);
	if (err!=nil) {
		Error.Println(fmt.Sprintf("Error at Scan() in Mainstats(): %v",err))
		os.Exit(2)
	}
	return out,nil
}
func Search(search_text string) (Json_search_result_t,error) {
	var search_res Json_search_result_t
	var found bool

	slen:=len(search_text)
	Info.Println(fmt.Sprintf("Search term: %v; length=%v",search_text,slen))
	if (slen>2) {
		if (search_text[0]=='0') && (search_text[1]=='x') { // 0x hexadecimal prefix
			search_text=search_text[2:]
		}
	}
	search_res.Search_text=search_text
	if (strings.ToUpper(search_text)=="BLOCKCHAIN") {
		Info.Println(fmt.Sprintf("searching BLOCKCHAIN"))
		search_res.Search_text="BLOCKCHAIN"
		search_res.Object_type=JSON_OBJ_TYPE_ACCOUNT
		search_res.Account,_,_=Get_account("0")
		return search_res,nil
	}

	// First, GetBlockByNumber()
	block_num,err:=strconv.ParseUint(search_text,10,64)
	if (err==nil) {
		search_res.Object_type=JSON_OBJ_TYPE_BLOCK
		search_res.Block,found,err=Get_block(int(block_num),"")
		return search_res,err
	}

	// Now do 'GetBlockByHash'
	search_res.Block,found,err=Get_block(-1,search_text)
	if (found) {
		search_res.Object_type=JSON_OBJ_TYPE_BLOCK
		return search_res,err
	}
	// now by address
	addr_str:=strings.ToLower(search_text)
	account,account_found,_:=Get_account(addr_str)
	if account_found {
		search_res.Object_type=JSON_OBJ_TYPE_ACCOUNT
		search_res.Account=account
		Info.Println(fmt.Sprintf("Returning data for Account %v",search_text))
		return search_res,nil
	} else {
		if len(search_text)==40 { // it is an address
			search_res.Object_type=JSON_OBJ_TYPE_ACCOUNT
			search_res.Account=account
			Info.Println(fmt.Sprintf("Account %v was not found.",search_text))
			return search_res,nil
		}
	}
	search_res.Transaction,err=Get_transaction(search_text)
	if (err==nil) {
		search_res.Object_type=JSON_OBJ_TYPE_TRANSACTION
		search_res.Search_text=search_text
		return search_res,nil
	}
	return search_res,nil
}
func Get_block_list(up_to_block int) ([]Json_aex_bhdr_t,error) {
	var last_blocks []Json_aex_bhdr_t

	if (up_to_block < -1 ) {
		return last_blocks,errors.New("Invalid parameter")
	}
	var last_block_num int
	if (up_to_block==-1) {
		last_block_num=get_last_block()
		if (last_block_num<0) {
			return last_blocks,nil
		}
	} else {
		last_block_num=up_to_block
	}
	var query string
	query="SELECT b.block_num,b.num_tx,b.num_vt,b.val_transferred,address AS miner FROM block AS b,account WHERE (block_num<=$1) AND (miner_id=account.account_id) ORDER BY block_num DESC LIMIT 9"
	rows,err:=db.Query(query,last_block_num)
	defer rows.Close()
	if err!=nil {
		Error.Println(fmt.Sprintf("Error at Get_block_list(): %v",err))
		os.Exit(2)
	}
	for rows.Next() {
		var hdr Json_aex_bhdr_t
		err=rows.Scan(&hdr.Block_number,&hdr.Num_transactions,&hdr.Num_value_transfers,&hdr.Val_transferred,&hdr.Miner)
		if (err!=nil) {
			Error.Println(fmt.Sprintf("Error at rows.Scan() in Get_block_list: %v",err))
			os.Exit(2)
		}
		last_blocks=append(last_blocks,hdr)
	}
	return last_blocks,nil
}
func Get_account(acct_addr string) (Json_account_t,bool,error) {
	var found bool=false;
	var o Json_account_t	// output
	var query string

	acct_addr=strings.ToLower(acct_addr)
	o.Address=acct_addr
	if len(acct_addr)<2 {
		if acct_addr!="0" { // blockchain account code
			return o,found,errors.New("Invalid address")
		}
	} else {
		if (acct_addr[0]=='0') && (acct_addr[1]=='x') {
			acct_addr=acct_addr[2:]
		}
	}
	query=`
		SELECT 
			a.account_id,
			a.owner_id,
			a.last_balance,
			a.num_tx,
			a.num_vt,
			a.ts_created,
			a.block_created,
			a.deleted,
			a.block_sd,
			a.address,
			o.address as owner_address
		FROM account as a
		LEFT JOIN account as o ON a.owner_id=o.account_id
		WHERE a.address=$1
	`
	var row *sql.Row
	row=db.QueryRow(query,acct_addr);
	var tmp_owner_address sql.NullString
	err:=row.Scan(&o.Account_id,&o.Owner_id,&o.Balance,&o.Num_transactions,&o.Num_value_transfers,&o.Ts_created,&o.Block_created,&o.Deleted,&o.Block_suicided,&o.Address,&tmp_owner_address)
	if err!=nil {
		if (err==sql.ErrNoRows) {
			return o,found,nil
		} else {
			return o,found,err
		}
	}
	if tmp_owner_address.Valid {
		o.Owner_address=tmp_owner_address.String
	}
	found=true
	return o,found,nil
}
func Get_token_account(acct_addr string) (Json_tokacct_t,bool,error) {
	var found bool=false;
	var o Json_tokacct_t	// output
	var query string

	o.Address=acct_addr
	acct_addr=strings.ToLower(acct_addr)
	if len(acct_addr)<2 {
		return o,found,errors.New("Invalid address")
	}
	if (acct_addr[0]=='0') && (acct_addr[1]=='x') {
		acct_addr=acct_addr[2:]
	}
	query=`
		SELECT 
			a.account_id,
			a.ts_created,
			a.block_created,
			a.address
		FROM tokacct as a
		WHERE a.address=$1
	`
	var row *sql.Row
	row=db.QueryRow(query,acct_addr);
	err:=row.Scan(&o.Account_id,&o.Ts_created,&o.Block_created,&o.Address)
	if (err!=nil) {
		if (err==sql.ErrNoRows) {
		} else {
			found=true
		}
	}
	query=`
		SELECT contract_id FROM tokop
		WHERE to_id=$1
		ORDER BY block_num DESC LIMIT 1
		`
	o.Has_tokens=false
	var cid int
	row=db.QueryRow(query,o.Account_id);
	err=row.Scan(&cid)
	if err!=nil {
		if (err==sql.ErrNoRows) {
		} else {
			return o,found,err;
		}
	} else {
		o.Has_tokens=true
	}
	
	return o,found,nil
}
func Get_uncles(block_num int) (Json_uncles_t,error) {
	var output Json_uncles_t
	var num_uncles int = 0;
	var query string
	query=`
		SELECT 
			u.block_num,
			u.parent_num,
			u.block_hash,
			u.block_ts,
			m.address,
			u.difficulty,
			u.total_dif,
			u.gas_limit,
			u.gas_used,
			u.nonce,
			p.block_hash AS parent_hash,
			u.uncle_hash,
			u.extra
		FROM uncle AS u
			LEFT JOIN account AS m ON u.miner_id=m.account_id,
			block AS p
		WHERE 
			p.block_num=$1
		`
	rows,err:=db.Query(query,block_num)
	defer rows.Close()
	if (err!=nil) {
		Error.Println("Error getting uncles: %v",err)
		os.Exit(2)
	}
	if rows.Next() {
		err=rows.Scan(&output.Uncle1.Number,
					&output.Uncle1.Parent_num,
					&output.Uncle1.Hash,
					&output.Uncle1.Timestamp,
					&output.Uncle1.Miner,
					&output.Uncle1.Difficulty,
					&output.Uncle1.Total_difficulty,
					&output.Uncle1.Gas_limit,
					&output.Uncle1.Gas_used,
					&output.Uncle1.Nonce,
					&output.Uncle1.Parent_hash,
					&output.Uncle1.Sha3uncles,
					&output.Uncle1.Extra_data)
		if (err!=nil) {
			Error.Println(fmt.Sprintf("Error in Scan() while getting uncles: %v",err))
			os.Exit(1)
		}
		num_uncles++
	}
	if rows.Next() {
		err=rows.Scan(&output.Uncle2.Number,
					&output.Uncle2.Parent_num,
					&output.Uncle2.Hash,
					&output.Uncle2.Timestamp,
					&output.Uncle2.Miner,
					&output.Uncle2.Difficulty,
					&output.Uncle2.Total_difficulty,
					&output.Uncle2.Gas_limit,
					&output.Uncle2.Gas_used,
					&output.Uncle2.Nonce,
					&output.Uncle2.Parent_hash,
					&output.Uncle2.Sha3uncles,
					&output.Uncle2.Extra_data)
		if (err!=nil) {
			Error.Println(fmt.Sprintf("Error in Scan() while getting uncles: %v",err))
			os.Exit(1)
		}
		num_uncles++
	}
	output.Num_uncles=num_uncles;
	output.Block_num=block_num;
	return output,nil
}
func Get_transaction(hash string) (Json_transaction_t,error) {
	var tx Json_transaction_t
	var query string
	query=`
		SELECT 
			tx.tx_id,
			tx.from_id,
			tx.to_id,
			b.block_ts,
			src.address AS from_addr,
			dst.address AS to_addr,
			tx.tx_value::text,
			tx.val_transferred::text,
			tx.tx_hash,
			tx.gas_limit,
			tx.gas_used,
			tx.gas_price,
			tx.nonce::text,
			tx.block_num,
			tx.tx_index,
			tx.tx_status,
			tx.num_vt,
			tx.v,
			tx.r,
			tx.s,
			tx.tx_error,
			tx.vm_error
		FROM transaction AS tx
		LEFT JOIN account as src ON tx.from_id=src.account_id
		LEFT JOIN account as dst ON tx.to_id=dst.account_id,
		block AS b
		WHERE tx.block_num=b.block_num AND tx.tx_hash=$1 
		`
	row:=db.QueryRow(query,hash);
	err:=row.Scan(&tx.Tx_id,&tx.From_id,&tx.To_id,&tx.Tx_timestamp,&tx.From_addr,&tx.To_addr,&tx.Value,&tx.Val_transferred,&tx.Tx_hash,&tx.Gas_limit,&tx.Gas_used,&tx.Gas_price,&tx.Nonce,&tx.Block_num,&tx.Tx_index,&tx.Tx_status,&tx.Num_vt,&tx.V,&tx.R,&tx.S,&tx.Tx_error,&tx.Vm_error)
	if (err!=nil) {
		if err==sql.ErrNoRows {
			return tx,ErrNoRows;
		} else {
			Error.Println(fmt.Sprintf("Scan failed at Get_transaction(): query=%v, %v",query,err))
			os.Exit(2)
		}
	}
	gas_used:=big.NewInt(0)
	gas_used.SetString(tx.Gas_used,10)
	gas_price:=big.NewInt(0)
	gas_price.SetString(tx.Gas_price,10)
	cost:=big.NewInt(0)
	cost=cost.Mul(gas_used,gas_price)
	tx.Cost=cost.String()
	last_block_num:=get_last_block()
	tx.Confirmations=last_block_num-tx.Block_num
	return tx,nil
}
func Pending_transaction_watch(hash string) (Json_pending_tx_watch_t,error) {
	var output Json_pending_tx_watch_t
	var query string
	query=`
		SELECT  
			(SELECT block_num AS tx_block_num FROM transaction WHERE tx_hash=$1),
			(SELECT block_num AS last_block_num FROM last_block)
		`
	row:=db.QueryRow(query,hash);
	var tmp_block_num sql.NullInt64
	err:=row.Scan(&tmp_block_num,&output.Last_block_num)
	if (err!=nil) {
		if err==sql.ErrNoRows {
			return output,ErrNoRows;
		} else {
			Error.Println(fmt.Sprintf("Scan failed at Pending_transaction_watch(): query=%v, %v",query,err))
			os.Exit(2)
		}
	}
	if tmp_block_num.Valid {
		output.Processed_block_num=int(tmp_block_num.Int64)
		if output.Processed_block_num>0 {
			output.Confirmations=output.Last_block_num-output.Processed_block_num
			output.Processed=true
		}
	}
	output.Tx_hash=hash
	return output,nil
}
func Get_eth_aet_prices() Json_eth_aet_prices_t {
	var query string
	var output Json_eth_aet_prices_t 

	query=`SELECT symbol,round(price,2) as price FROM coins WHERE symbol IN ('ETH','AET')`
	rows,err:=db.Query(query)
	if err!=nil {
		Error.Println(fmt.Sprintf("Error querying coin prices: %v",err))
		os.Exit(2)
	}
	defer rows.Close()
	for rows.Next() {
		var price float32
		var symbol string
		err=rows.Scan(&symbol,&price)
		if err!=nil {
			Error.Println(fmt.Sprintf("Error at Scan() in price eth/aet loop :%v",err))
			os.Exit(2)
		}
		if symbol=="ETH" {
			output.Eth_price=price
		}
		if symbol=="AET" {
			output.Aet_price=price
		}
	}
	return output
}
func New_blocks_exist(block_num int) (bool,error) {
	var query string
	query="SELECT block_num FROM block WHERE block_num>$1"
	row := db.QueryRow(query,block_num)
	var null_block_num sql.NullInt64
	var err error
	err=row.Scan(&null_block_num);
	if (err!=nil) {
		if err==sql.ErrNoRows {
			return false,nil
		} else {
			Error.Println(fmt.Sprintf("Error in get_last_block_num(): %v",err))
			os.Exit(2)
		}
	}
	return true,nil
}
func Stats_difficulty() (Stats_array_t,error) {
	var output Stats_array_t;
	var query string
	var tmp_value,divisor float64
	var tmp_ts int
	var max_value float64
	var min_value float64

	var block_interval int =2000							// sampling is done every 200 blocks
	var blocks_to_get int = block_interval * 200		// we are going to get 100 points for the char	
	var starting_block int=0;
	last_block_num:=get_last_block()
	if (last_block_num>-2) {
		starting_block=last_block_num-blocks_to_get
		if (starting_block<0) {
			starting_block=0
		}
	} else {
		return output,errors.New("Cant get (valid) last block number")
	}
	first_block_num:=int(starting_block/block_interval)
	var table_name string="stats_difficulty_"+strconv.Itoa(first_block_num)

	query=`
		SELECT EXISTS (
	   			SELECT 1
				FROM   information_schema.tables 
				WHERE  table_schema = 'public'
				AND    table_name = $1
			);
			`
	row:=db.QueryRow(query,table_name);
	var exists bool
	err:=row.Scan(&exists)
	if (err!=nil) {
		return output,err
	}
	if !exists { // the temporary table is used as cache
		query=`
			CREATE TABLE IF NOT EXISTS `+table_name+` AS
			SELECT 
				b.difficulty,
				b.block_ts,
				b.block_num
			FROM 
				block AS b
			WHERE
				((b.block_num%$2)=0) AND (b.block_num>$1)
			ORDER BY 
				b.block_num ASC
			LIMIT 200
			`
		_,err:=db.Exec(query,starting_block,block_interval)
		if err!=nil {
			Error.Println(fmt.Sprintf("Error creating table for difficulty stats: %v",err))
			os.Exit(2)
		}
	}
	query=`SELECT difficulty,block_ts FROM `+table_name+` ORDER BY block_num`
	rows,err:=db.Query(query);
	if (err!=nil) {
		Error.Println(fmt.Sprintf("failed to execute query: %v, error=%v",query,err))
		return output,err
	}
	defer rows.Close()
	not_done:=rows.Next()
	min_max_not_set:=true
	if (not_done) {
		for (not_done) {
			err=rows.Scan(&tmp_value,&tmp_ts)
			if (err!=nil) {
				Error.Println(fmt.Sprintf("Error in get_last_block_num(): %v",err))
				return output,err
			}
			if min_max_not_set {
				max_value=tmp_value;
				min_value=tmp_value;
				min_max_not_set=false;
			}
			output.Values=append(output.Values,tmp_value)
			output.Timestamps=append(output.Timestamps,tmp_ts)
			if (tmp_value>max_value) {
				max_value=tmp_value
			}
			if (tmp_value<min_value) {
				min_value=tmp_value
			}
			not_done=rows.Next()
		}
		mid_point:=(max_value+min_value)/2.0

		divisor,output.Unit=get_units(mid_point)
		// this cycle does data normalization
		for i,entry:= range output.Values {
			output.Values[i]=entry/divisor
		}
	}

	return output,nil
}
func get_units(mid_point float64) (divisor float64,unit string) {

	if mid_point<1000.0 {
		return 1.0,""
	}
	if mid_point < 1000000.0 {
		return 1000.0,"K"
	}
	if mid_point < 1000000000.0 {
		return     1000000.0,   "M"
	}
	if mid_point < 1000000000000.0 {
		return     1000000000.0,   "G"
	}
	if mid_point < 1000000000000000.0 {
		return     1000000000000.0,   "T"
	}
	if mid_point < 1000000000000000000.0 {
		return     1000000000000000.0,   "P"
	}
	if mid_point < 1000000000000000000000.0 {
		return     1000000000000000000.0,   "E"
	}
	return 1.0,`?`
}
func get_last_block() int {

	var query string
	query="SELECT block_num FROM last_block LIMIT 1";
	row := db.QueryRow(query)
	var null_block_num sql.NullInt64
	var err error
	err=row.Scan(&null_block_num);
	if (err!=nil) {
		Error.Println(fmt.Sprintf("Error in get_last_block_num(): %v",err))
		os.Exit(2)
	}
	if (null_block_num.Valid) {
		return int(null_block_num.Int64)
	} else {
		return -2
	}
}
func lookup_account(addr_str string) (account_id int,owner_id int,last_balance string,err error) {
	query:="SELECT account_id,owner_id,last_balance FROM account WHERE address=$1"
	row:=db.QueryRow(query,addr_str);
	err=row.Scan(&account_id,&owner_id,&last_balance);
	if err!=nil {
		if err==sql.ErrNoRows {
			account_id=0
			owner_id=0
			last_balance="-1"
			err=nil
		}
	}
	return account_id,owner_id,last_balance,err
}
func lookup_token_account(addr_str string) (account_id int,err error) {
	query:="SELECT account_id FROM tokacct WHERE address=$1"
	row:=db.QueryRow(query,addr_str);
	err=row.Scan(&account_id);
	if err!=nil {
		if err==sql.ErrNoRows {
			account_id=0
			err=nil
		}
	}
	return account_id,err
}
func get_account_balance(account_id int) (string,error) {
	var query string
	query="SELECT get_balance($1,-1)"
	row:=db.QueryRow(query,account_id)
	var balance_str string
	err:=row.Scan(&balance_str)
	if (err!=nil) {
		if err==sql.ErrNoRows {
			Error.Println("get_balance() returned no rows, make sure this PL/SQL function exists")
			os.Exit(2)
		} else {
			return "-1",err
		}
	}
	return balance_str,nil
}
func get_fungible_token_balance_query() string {
	return `
		SELECT 
			tokop_id,tx_id,contract_id,block_num,block_ts,from_id,to_id,from_balance,to_balance,value,kind FROM
		(
			(
				SELECT tokop_id,tx_id,contract_id,block_num,block_ts,from_id,to_id,from_balance,to_balance,value,kind
				FROM tokop
				WHERE contract_id=$1 AND from_id=$2
			) UNION ALL (
				SELECT tokop_id,tx_id,contract_id,block_num,block_ts,from_id,to_id,from_balance,to_balance,value,kind
				FROM tokop
				WHERE contract_id=$1 AND to_id=$2
			)
		) AS op
		ORDER BY block_num DESC,tokop_id DESC
		LIMIT 1
	`
}
