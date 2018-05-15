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
			tx.block_id,
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
		err:=rows.Scan(&tx.Tx_id,&tx.From_id,&tx.To_id,&tx.From_addr,&tx.To_addr,&tx.Value,&tx.Tx_hash,&tx.Gas_limit,&tx.Gas_used,&tx.Gas_price,&tx.Nonce,&tx.Block_id,&tx.Block_num,&tx.Tx_index,&tx.Tx_status,&tx.V,&tx.R,&tx.S,&tx.Tx_error)
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
			v.block_id,
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
		ORDER BY v.block_num,v.valtr_id
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
		err:=rows.Scan(&vt.Valtr_id,&tx_id,&tx_hash,&tx_index,&vt.Block_id,&vt.Block_num,&vt.To_addr,&vt.To_id,&vt.From_addr,&vt.To_addr,&vt.Value,&vt.From_balance,&vt.To_balance,&vt.Kind,&vt.Error)
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
func Get_account_transactions(acct_addr string,offset int,limit int) (Json_tx_set_t,error) {
	var output Json_tx_set_t

	if limit==0 {
		limit=default_limit
	}
	acct_addr=strings.ToLower(acct_addr)
	account_id,_,_:=lookup_account(acct_addr)
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
			tx.block_id,
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
		ORDER BY tx.block_num DESC,tx.tx_index DESC
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
		err:=rows.Scan(&tx.Tx_id,&tx.From_id,&tx.To_id,&tx.From_addr,&tx.To_addr,&tx.Value,&tx.Tx_hash,&tx.Gas_limit,&tx.Gas_used,&tx.Gas_price,&tx.Nonce,&tx.Block_id,&tx.Block_num,&tx.Tx_index,&tx.Tx_status,&tx.V,&tx.R,&tx.S,&tx.Tx_error)
		if (err!=nil) {
			if err==sql.ErrNoRows {
				break;
			} else {
				Error.Println(fmt.Sprintf("Scan failed at Getblocktransactions(): %v",err))
				os.Exit(2)
			}
		}
		output.Transactions=append(output.Transactions,tx)
	}
	output.Offset=offset;
	output.Limit=limit;
	output.Account_id=int(account_id);
	output.Account_address=acct_addr
	output.Account_balance,err=get_account_balance(account_id)
	return output,err
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
	account_id,_,_:=lookup_account(acct_addr)
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
			t.tx_hash,
			v.error
		FROM 
			value_transfer AS v
		LEFT JOIN account AS src ON v.from_id=src.account_id
		LEFT JOIN account AS dst ON v.to_id=dst.account_id
		LEFT JOIN transaction AS t ON v.tx_id=t.tx_id
		WHERE 
			(
				(v.from_id=$1) OR (v.to_id=$1)
			)  
			ORDER BY block_num DESC,valtr_id DESC
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
		output.Value_transfers=append(output.Value_transfers,vt)

	}
	output.Offset=offset;
	output.Limit=limit;
	output.Account_id=int(account_id);
	output.Account_address=acct_addr
	output.Account_balance,err=get_account_balance(account_id)
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
		ORDER BY v.valtr_id ASC
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
	output.Account_id=0;
	output.Account_address=`N/A`
	return output,nil
}
func Main_stats() (Json_main_stats_t,error) {
	var out Json_main_stats_t

	var query string
	query="SELECT round(hash_rate)::text,round(block_time,2)::text,round(tx_per_block,1)::text,gas_price::text,tx_cost::text,supply::text,round(difficulty)::text,last_block::text FROM mainstats"
	row := db.QueryRow(query)
	err:=row.Scan(&out.Hash_rate,&out.Block_time,&out.Tx_per_block,&out.Gas_price,&out.Tx_cost,&out.Supply,&out.Difficulty,&out.Last_block);
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
	if (slen>2) {
		if (search_text[0]=='0') && (search_text[1]=='x') { // 0x hexadecimal prefix
			search_text=search_text[2:]
		}
	}

	search_res.Search_text=search_text
	if (strings.ToUpper(search_text)=="BLOCKCHAIN") {
		search_res.Account_id,_,search_res.Account_balance=lookup_account("0")
		if (search_res.Account_id==-1) {
			search_res.Search_text="BLOCKCHAIN"
			search_res.Object_type=JSON_OBJ_TYPE_ACCOUNT
			return search_res,nil
		}
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
	search_res.Account_id,_,search_res.Account_balance=lookup_account(addr_str)
	if (search_res.Account_id!=0) {
		search_res.Object_type=JSON_OBJ_TYPE_ACCOUNT
		return search_res,err
	}
	search_res.Transaction,err=Get_transaction(search_text)
	if (err==nil) {
		search_res.Object_type=JSON_OBJ_TYPE_TRANSACTION
		search_res.Search_text=search_text
		return search_res,nil
	}
	return search_res,nil
}
func Get_block_list(up_to_block int) ([]Last_block_info_t,error) {
	var last_blocks []Last_block_info_t

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
	query="SELECT block_num,num_tx FROM block WHERE block_num<=$1 ORDER BY block_num DESC LIMIT 9"
	rows,err:=db.Query(query,last_block_num)
	defer rows.Close()
	for rows.Next() {
		var block_info Last_block_info_t
		err=rows.Scan(&block_info.Block_number,&block_info.Num_transactions)
		if (err!=nil) {
			Error.Println(fmt.Sprintf("Error at rows.Scan() in Get_block_list: %v",err))
			os.Exit(2)
		}
		last_blocks=append(last_blocks,block_info)
	}
	return last_blocks,nil
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
			b.extra
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
	err:=row.Scan(&output.Number,&output.Hash,&output.Timestamp,&output.Miner,&output.Num_transactions,&output.Num_value_transfers,&output.Num_uncles,&output.Difficulty,&output.Total_difficulty,&output.Gas_used,&output.Gas_limit,&output.Size,&output.Nonce,&parent_id,&output.Sha3uncles,&output.Extra_data)
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
	query="SELECT block_num,block_hash FROM block WHERE block_id=$1"
	row=db.QueryRow(query,parent_id)
	if (err==sql.ErrNoRows) {
		return output,true,errors.New(fmt.Sprintf("Parent block (id=%v) not found",parent_id))
	}
	return output,true,nil
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
			(u.block_id=p.block_id) AND (p.block_num=$1)
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
			src.address AS from_addr,
			dst.address AS to_addr,
			tx.tx_value::text,
			tx.tx_hash,
			tx.gas_limit,
			tx.gas_used,
			tx.gas_price,
			tx.nonce::text,
			tx.block_id,
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
		WHERE tx.tx_hash=$1 
		`
	row:=db.QueryRow(query,hash);
	err:=row.Scan(&tx.Tx_id,&tx.From_id,&tx.To_id,&tx.From_addr,&tx.To_addr,&tx.Value,&tx.Tx_hash,&tx.Gas_limit,&tx.Gas_used,&tx.Gas_price,&tx.Nonce,&tx.Block_id,&tx.Block_num,&tx.Tx_index,&tx.Tx_status,&tx.V,&tx.R,&tx.S,&tx.Tx_error)
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
func New_blocks_exist(block_num int) (bool,error) {
// returns true if new blocks higher than asking block_num exist in the DB
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
func lookup_account(addr_str string) (account_id int,owner_id int,last_balance string) {
	query:="SELECT account_id,owner_id,last_balance FROM account WHERE address=$1"
	row:=db.QueryRow(query,addr_str);
	err:=row.Scan(&account_id,&owner_id,&last_balance);
	if (err==sql.ErrNoRows) {
		return 0,0,"-1"
	} else {
		return account_id,owner_id,last_balance
	}
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
