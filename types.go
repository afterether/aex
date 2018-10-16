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
	Tx_timestamp	int		`json: "tx_timestamp"	gencodec:"required"`
	Confirmations	int		`json: "confirmations"	gencodec:"required"`
	Num_vt			int		`json: "num_vt"			gencodec:"required"`
	Val_transferred	string	`json: "val_transferred gencodec:"required"`
	V				string	`json: "v"				gencodec:"required"`
	R				string	`json: "r"				gencodec:"required"`
	S				string	`json: "s"				gencodec:"required"`
	Tx_error		string	`json: "tx_error"		gencodec:"required"`
	Vm_error		string	`json: "vm_error"		gencodec:"required"`
}
type Json_block_t struct {
	Number				int		`json: "number"`
	Parent_num			int		`json: "number"`
	Hash				string	`json: "hash"`
	Confirmations		int		`json: "confirmations"`
	Timestamp			int		`json: "timestamp"`
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
	Val_transferred		string	`json: "val_transferred"`
	Miner_reward		string	`json: "miner_reward"`
}
type Json_account_t struct {
	Account_id			int		`json: "account_id"`
	Owner_id			int		`json: "owner_id"`
	Address				string	`json: "address"`
	Owner_address		string	`json: "owner_address"`
	Balance				string	`json: "balance"`
	Num_transactions	int		`json: "num_tx"`
	Num_value_transfers	int		`json: "num_vt"`
	Ts_created			int		`json: "ts_created"`
	Block_created		int		`json: "block_created"`
	Deleted				int		`json: "deleted"`
	Block_suicided		int		`json: "block_suicided"`
}
type Json_tokacct_t struct {
	Account_id			int		`json: "account_id"`
	Ts_created			int		`json: "ts_created"`
	Block_created		int		`json: "block_created"`
	Has_tokens			bool	`json: "has_tokens"`
	Has_approved		bool	`json: "has_approved"`
	Address				string	`json: "address"`
}
type Json_aex_bhdr_t struct { // block header for AEX, it is not the same as types.Header, it is shorter
	Block_number				int		`json: "number"`
	Num_transactions	int		`json: "num_tx"`
	Num_value_transfers	int		`json: "num_vt"`
	Val_transferred		string `json: "val_transferred"`
	Miner				string	`json: "miner"`
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
	Tx_per_sec			string	`json: "tx_per_sec"`
	Gas_price			string	`json: "gas_price"`
	Tx_cost				string	`json: "tx_cost"`
	Supply				string	`json: "supply"`
	Difficulty			string	`json: "difficulty"`
	Volume				string	`json: "volume"`
	Activity			string	`json: "activity"`
	Last_block			int		`json: "last_block"`
}
type Json_search_result_t struct {
	Object_type		int
	Block			Json_block_t
	Transaction		Json_transaction_t
	Account			Json_account_t
	Search_text		string
}
type Json_vt_set_t struct {
	Account			Json_account_t
	Offset			int
	Limit			int
	Value_transfers	[]Json_value_transfer_t
}
type Json_tx_set_t struct {
	Account			Json_account_t
	Offset			int
	Limit			int
	Transactions	[]Json_transaction_t
}
type Token_info_t struct {
	Num_transfers	int64
	Contract_id		int
	Decimals		int
	Block_created	int
	Search_string	string
	Contract_addr	string
	Symbol			string
	Name			string
	Total_supply	string
	Tx_hash			string
	Non_fungible	bool
	i_ERC20			bool
	i_burnable		bool
	i_mintable		bool
	i_ERC223		bool
	i_ERC677		bool
	i_ERC721		bool
	i_ERC777		bool
	i_ERC827		bool
	m_erc20_name			bool
	m_erc20_symbol			bool
	m_erc20_decimals		bool
	m_erc20_total_supply	bool
	m_erc20_balance_of		bool
	m_erc20_allowance		bool
	m_erc20_transfer_from	bool
}
type Account_fungible_holding_t struct {
	Token			Token_info_t;
	Value			string
}
type Json_nonfungible_holding_t struct {
	Token			Token_info_t;
	Token_IDs		[]string
}
type Json_approval_holding_t struct {
	Token			Token_info_t;
	Amount			string
}
type Tok_approval_t struct {
	Block_num				int
	Timestamp				int
	Amount_approved			string
	Amount_transferred		string
	Amount_remaining		string
	From					string
	To						string
	Tx_hash					string
	Expired					bool
}
type Token_approvals_t  struct {
	Offset				int
	Limit				int
	Token				Token_info_t
	Approvals			[]Tok_approval_t
}
type Tok_acct_fungible_holdings_t struct {
	Offset			int
	Limit			int
	Holdings		[]Account_fungible_holding_t
	Address_list	string
}
type Tok_acct_nonfungible_holdings_t struct {
	Offset			int
	Limit			int
	Holdings		map[int]*Json_nonfungible_holding_t
	Address_list	string
}
type Tok_acct_approvals_t struct {
	Account_address		string
	Token				Token_info_t;
	Approvals			[]Tok_approval_t;
}
type Tok_transfer_t struct {
	Tokop_id		int64
	Block_num		int
	Ts_created		int
	Kind			int
	Tx_hash			string
	From			string
	To				string
	Value			string
	From_balance	string
	To_balance		string
	Non_fungible	bool
}
type Tok_acct_holder_t struct {
	Address			string
	Balance			string
}
type Token_holders_t struct {
	Offset				int
	Limit				int
	Token				Token_info_t
	Holders				[]Tok_acct_holder_t
}
type Token_transfers_t  struct {
	Offset				int
	Limit				int
	Token				Token_info_t
	Account_address		string
	Tokops				[]Tok_transfer_t
}
type Stats_array_t struct {
	Values			[]float64
	Timestamps		[]int
	Unit			string
	Starting_block	int
	Ending_block	int
}
type Json_balance_t struct {
	Address			string
	Balance			string
}
type Json_account_full_info_t struct {
	Offset			int
	Limit			int
	BAccount		Json_account_t
	TAccount		Json_tokacct_t
	Value_transfers	[]Json_value_transfer_t
	Transactions	[]Json_transaction_t
}
type Json_ft_acct_bal_t struct {
	Contract_address	string
	Account_address		string
	Balance				string
}
type Json_eth_aet_prices_t struct {
	Eth_price			float32
	Aet_price			float32
}
type Json_ftoken_sum_t struct {
	Decimals			int
	Contract_addr		string
	Symbol				string
	Balance				string
}
type Json_pending_tx_watch_t struct {
	Tx_hash					string
	Last_block_num			int
	Processed_block_num		int
	Confirmations			int
	Processed				bool
}
