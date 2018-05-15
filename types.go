package main
import (

)
const (
	JSON_OBJ_TYPE_UNDEFINED = iota
	JSON_OBJ_TYPE_BLOCK
	JSON_OBJ_TYPE_ACCOUNT
	JSON_OBJ_TYPE_TRANSACTION
	JSON_OBJ_TYPE_VALUE_TRANSFER
	JSON_OBJ_TYPE_UNCLE
)
type Json_value_transfer_t struct {
	Valtr_id		int64	`json: "valtr_id"		gencodec:"required"`
	Block_num		int		`json: "block_num"		gencodec:"required"`
	Block_id		int64	`json: "block_id"		gencodec:"required"`
	From_id			int64	`json: "from_id"		gencodec:"required"`
	To_id			int64	`json: "to_id"			gencodec:"required"`
	From_addr		string	`json: "from_addr"		gencodec:"required"`
	To_addr			string	`json: "to_addr"		gencodec:"required"`
	From_balance	string	`json: "from_balance"	gencodec:"required"`
	To_balance		string	`json: "to_balance"		gencodec:"required"`
	Value			string	`json: "value"			gencodec:"required"`
	Kind			int		`json: "kind"			gencodec:"rquired"`
	Tx_id			int64	`json: "tx_id"			gencodec:"required"`
	Tx_hash			string	`json: "tx_hash"		gencodec:"required"`
	Tx_index		int		`json: "tx_index"		gencodec:"required"`
	Direction		int		`jsoin:"direction"		gencodec:"required"`
	Error			string	`json: "error"			gencodec:"required"`
}
type Json_transaction_t  struct {
	Tx_id			int64
	From_id			int		`json: "from_id"		gencodec:"required"`
	To_id			int		`json: "to_id"			gencodec:"required"`
	From_addr		string	`json: "from_addr"		gencodec:"required"`
	To_addr			string	`json: "to_addr"		gencodec:"required"`
	Value			string	`json: "value"			gencodec:"required"`
	Tx_hash			string	`json: "tx_hash"		gencodec:"required"`
	Gas_limit		string	`json: "gas_limit"		gencodec:"required"`
	Gas_used		string	`json: "gas_used"		gencodec:"required"`
	Gas_price		string	`json: "gas_price"		gencodec:"required"`
	Cost			string	`json: "cost"			gencodec:"required"`
	Nonce			int		`json: "nonce"			gencodec:"required"`
	Block_id		int		`json: "block_id"		gencodec:"required"`
	Block_num		int		`json: "block_num"		gencodec:"required"`
	Tx_index		int		`json: "tx_index"		gencodec:"required"`
	Tx_status		int		`json: "tx_status"		gencodec:"required"`
	Confirmations	int		`json: "confirmations"	gencodec:"required"`
	V				string	`json: "v"				gencodec:"required"`
	R				string	`json: "r"				gencodec:"required"`
	S				string	`json: "s"				gencodec:"required"`
	Tx_error		string	`json: "tx_error"		gencodec:"required"`
}
type Json_block_t struct {
	Number				int		`json: "number"`
	Parent_num			int		`json: "number"`
	Hash				string	`json: "hash"`
	Confirmations		int		`json: "confirmations"`
	Timestamp			uint64	`json: "timestamp"`
	Miner				string	`jsoin: "miner"`
	Num_transactions	int		`json: "num_tx"`
	Num_value_transfers	int		`json: "num_vt"`
	Num_uncles			int		`json: "num_uncles"`
	Difficulty			string	`json: "difficulty"`
	Total_difficulty	string	`json: "total_difficulty"`
	Gas_used			uint64	`json: "gas_used"`
	Gas_limit			uint64	`json: "gas_limit"`
	Size				float64	`json: "block_size"`
	Nonce				uint64	`json: "nonce"`
	Parent_hash			string	`json: "parent_hash"`
	Sha3uncles			string	`json: "sha3uncles"`
	Extra_data			string	`json: "extra_data"`
}
type Json_uncles_t struct {
	Block_num			int					`json: "block_num"`
	Num_uncles			int					`json: "num_uncles"`
	Uncle1				Json_block_t		`json: "uncle1"`
	Uncle2				Json_block_t		`json: "uncle2"`
}
type Json_main_stats_t struct {
	Hash_rate			string	`json: "hash_rate"`
	Block_time			string	`json: "block_time"`
	Tx_per_block		string	`json: "tx_per_block"`
	Gas_price			string	`json: "gas_price"`
	Tx_cost				string	`json: "tx_cost"`
	Supply				string	`json: "supply"`
	Difficulty			string	`json: "difficulty"`
	Last_block			int		`json: "last_block"`
}
type Json_search_result_t struct {
	Object_type		int
	Block			Json_block_t
	Transaction		Json_transaction_t
	Search_text		string
	Account_balance	string
	Account_id		int
}
type Json_vt_set_t struct {
	Account_address	string
	Account_balance	string
	Account_id		int
	Offset			int
	Limit			int
	Value_transfers	[]Json_value_transfer_t
}
type Json_tx_set_t struct {
	Account_address	string
	Account_balance	string
	Account_id		int
	Offset			int
	Limit			int
	Transactions	[]Json_transaction_t
}
type Last_block_info_t struct {
	Block_number		uint64	`json: "block_number"`
	Num_transactions	int		`json: "num_transactions"`
}
type Stats_array_t struct {
	Values			[]float64
	Timestamps		[]int
	Unit			string
	Starting_block	int
	Ending_block	int
}

