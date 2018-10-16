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
/// "CONSTANTS"
var ticker_symbol='ETH'
var currency_name='Ethereum'
var network_name='Ethereum Main Net'
var block_reward=3
var block_scrolling_num=9
var default_limit=20
var check_new_blocks_interval=10; // seconds
/// VARIABLES
var current_tab_num=0
var current_account_tab_num=0
var current_block_num=-1
var current_VT_offset=0
var current_TX_offset=0
var current_account_VT_offset=0;
var current_account_TX_offset=0;
var first_block_num_in_line=-1;
var listening_mode=true;
var highest_block_num=-1;
function init() {
	var elt=document.getElementById('currency_name')
	elt.innerHTML=currency_name
	elt=document.getElementById('currency_ticker_sym')
	elt.innerHTML=ticker_symbol
	var elt=document.getElementById('network_name')
	elt.innerHTML=network_name
	elt=document.getElementById('top_stats_block_reward')
	elt.innerHTML=block_reward+' '+ticker_symbol
	elt=document.getElementById('next_block_line')
	elt.addEventListener("click",next_block_line)
	elt=document.getElementById('prev_block_line')
	elt.addEventListener("click",prev_block_line)
	setTimeout(check_for_new_blocks,check_new_blocks_interval*1000)
	var path_array = window.location.href.split( '?' );
	if (path_array.length==2) {
		var search_term=path_array[1]
		if (search_term.length>0) {
			console.log("search term = "+search_term)
			search(search_term)
		}
	}
}
function short_search(e) {
	e.preventDefault()
	let clicked_object = e.target;
	search(clicked_object.dataset.search_term)
}
function get_last_blocks(last_block_num) {
	
	Ajax_GET('/blist/'+last_block_num,function(data) {
			var response=JSON.parse(data)
			set_last_blocks(response.result)
	});
}
function get_main_stats() {
	Ajax_GET('/mainstats',function(data) {
		var response,data;
		response=JSON.parse(data)
		load_main_stats(response.result)
	});
}
function get_block(block_number) {

	Ajax_GET('/block/'+block_number,function(data) {
		var result,response;
		response=JSON.parse(data)
		result=response.result
		load_block(result)
		show_section(1)
		select_and_show_pane(1)
	});
}
function get_uncles(block_number) {
	var elt;
	elt=document.getElementById("uncles_container")
	elt.style.display="none";
	Ajax_GET('/uncles/'+block_number,function(data) {
		var result,response;
		response=JSON.parse(data)
		result=response.result
		load_uncles(result)
		var elt
		elt=document.getElementById("uncles_container")
		elt.style.display="block";
		show_section(1)
		select_and_show_pane(4)
	});
}
function get_block_transactions(block_number) {
	elt=document.getElementById("block_transactions_loading_image")
	elt.style.display="inline-block";
	elt=document.getElementById("block_transaction_table")
	elt.style.display="none";
	Ajax_GET('/btx/'+block_number,function(data) {
		response=JSON.parse(data)
		load_block_transactions(response.result)
		elt=document.getElementById("block_transactions_loading_image")
		elt.style.display="none";
	});
}
function get_block_value_transfers(block_num) {
	var elt
	var table_element=document.getElementById('block_value_transfers_table')
	var elts=table_element.getElementsByTagName("TBODY")
	var tbody=elts[0]
	while (tbody.lastElementChild) {
		tbody.removeChild(tbody.lastElementChild)
	}
	elt=document.getElementById("block_value_transfers_loading_image")
	elt.style.display="inline-block";
	elt=document.getElementById("block_value_transfers_table")
	elt.style.display="none";
	Ajax_GET('/bvt/'+block_num,function(data) {
		response=JSON.parse(data)
		result=response.result
		var table=document.getElementById("block_value_transfers_table")
		if (table) {
			load_block_value_transfers(table,result)
		}
		elt=document.getElementById("block_value_transfers_loading_image")
		elt.style.display="none";
		elt=document.getElementById("block_value_transfers_table")
		elt.style.display="block";
	});
}
function get_account_value_transfers(account_address,offset,limit) {

	var table_element=document.getElementById("account_value_transfers_table")
	var elts=table_element.getElementsByTagName("TBODY")
	var tbody=elts[0]
	while (tbody.lastElementChild) {
		tbody.removeChild(tbody.lastElementChild)
	}

	var elt
	elt=document.getElementById("account_value_transfers_loading_image")
	elt.style.display="inline-block";
	elt=document.getElementById("account_value_transfers_container")
	elt.style.display="none";
	Ajax_GET('/avt/'+account_address+'/'+offset+'/'+limit,function(data) {
		var elt
		var response=JSON.parse(data)
		var result=response.result
		var table=document.getElementById("account_value_transfers_table")
		load_account_value_transfers(table,result)
		elt=document.getElementById("account_value_transfers_loading_image")
		elt.style.display="none";
		elt=document.getElementById("account_value_transfers_container")
		elt.style.display="block";
	});
}
function get_transaction_value_transfers(tx_hash) {
	var table_element=document.getElementById('transaction_value_transfers_table')
	var elts=table_element.getElementsByTagName("TBODY")
	var tbody=elts[0]
	while (tbody.lastElementChild) {
		tbody.removeChild(tbody.lastElementChild)
	}
	Ajax_GET('/tvt/'+tx_hash,function(data) {
		var result,response;
		response=JSON.parse(data)
		result=response.result
		var table=document.getElementById("transaction_value_transfers_table")
		if (table) {
			load_transaction_value_transfers(table,result)
		}
	});
}
function get_account_transactions(account_address,offset,limit) {
	elt=document.getElementById("account_transactions_loading_image")
	elt.style.display="inline-block";
	elt=document.getElementById("account_transaction_table")
	elt.style.display="none";
	Ajax_GET('/atx/'+account_address+'/'+offset+'/'+limit,function(data) {
		var result,response;
		response=JSON.parse(data)
		result=response.result
		var table=document.getElementById("account_transaction_table")
		load_account_transactions(table,result)
		elt=document.getElementById("account_transactions_loading_image")
		elt.style.display="none";
		elt=document.getElementById("account_transaction_table")
		elt.style.display="block";
	});
}
function load_main_stats(main_stats) {
	var elt
	elt=document.getElementById("top_stats_last_block")
	if (elt) {
		elt.innerHTML=main_stats.Last_block
	}
	elt=document.getElementById("top_stats_difficulty")
	if (elt) {
		var value=format_big_number(main_stats.Difficulty)
		elt.innerHTML=value
		elt.title=add_commas(main_stats.Difficulty)
	}
	elt=document.getElementById("top_stats_block_time")
	if (elt) {
		elt.innerHTML=main_stats.Block_time+' sec'
	}
	elt=document.getElementById("top_stats_tx_per_block")
	if (elt) {
		elt.innerHTML=main_stats.Tx_per_block
	}
	elt=document.getElementById("top_stats_gas_price")
	if (elt) {
		elt.innerHTML=main_stats.Gas_price+' GWei'
	}
	elt=document.getElementById("top_stats_tx_cost")
	if (elt) {
		elt.innerHTML=main_stats.Tx_cost+' '+ticker_symbol
	}
	elt=document.getElementById("top_stats_hash_rate")
	if (elt) {
		var value=format_big_number(main_stats.Hash_rate)
		elt.innerHTML=value
		elt.title=add_commas(main_stats.Hash_rate)
	}
	elt=document.getElementById("top_stats_supply")
	if (elt) {
		var value=format_value(main_stats.Supply)
		elt.innerHTML=value.hi
	}
}
function load_block(block) {
			var elt
	set_block_header(block.Number)
			elt=document.getElementById("block_block_num")
			elt.innerHTML=block.Number
			elt=document.getElementById("block_hash")
			elt.innerHTML=block.Hash
			elt=document.getElementById("block_confirmations")
			elt.innerHTML=block.Confirmations
			elt=document.getElementById("block_timestamp")
			if (elt) {
				elt.innerHTML=block.Timestamp
				elt=document.getElementById("block_datetime")
				if (elt) {
					var date = new Date(block.Timestamp*1000);
				    var date_str=('0' + date.getDate()).slice(-2) + '/' + ('0' + (date.getMonth() + 1)).slice(-2) + '/' + date.getFullYear() + ' ' + ('0' + date.getHours()).slice(-2) + ':' + ('0' + date.getMinutes()).slice(-2) + ':' +('0' + date.getSeconds()).slice(-2);
					elt.innerHTML=date_str
				}
			}
			elt=document.getElementById("block_miner")
			while (elt.lastElementChild) {
				elt.removeChild(elt.lastElementChild)
			}
			var a=document.createElement('A')
			var link_text = document.createTextNode(block.Miner);
			a.appendChild(link_text);
			a.title = format_address4link(block.Miner)
			a.className="link"
			a.href = '/index.html?'+format_address4link(block.Miner)
			a.dataset.search_term=format_address4link(block.Miner)
			a.addEventListener("click",short_search)
			elt.appendChild(a)
		//	elt.innerHTML=block.Miner
		//
			elt=document.getElementById("block_num_transactions")
			elt.innerHTML=block.Num_transactions
			elt=document.getElementById("block_difficulty")
			elt.innerHTML=block.Difficulty
			elt=document.getElementById("block_total_difficulty")
			elt.innerHTML=block.Total_difficulty
			elt=document.getElementById("block_gas_used")
			elt.innerHTML=block.Gas_used
			elt=document.getElementById("block_gas_limit")
			elt.innerHTML=block.Gas_limit
			elt=document.getElementById("block_size")
			elt.innerHTML=block.Size
			elt=document.getElementById("block_nonce")
			elt.innerHTML=block.Nonce
			elt=document.getElementById("block_parent_hash")
			elt.innerHTML=block.Parent_hash
			elt=document.getElementById("block_sha3uncles")
			elt.innerHTML=block.Sha3uncles
			elt=document.getElementById("block_extra_data")
			elt.innerHTML=block.Extra_data
			current_block_num=block.Number
}
function load_uncles(udata) {
	var elt
	elt=document.getElementById("block_uncles_empty")
	if (udata.Num_uncles==0) {
		elt.style.display="block";
	} else {
		elt.style.display="none";
	}
	elt=document.getElementById("uncle1")
	elt.style.display="none";
	elt=document.getElementById("uncle2")
	elt.style.display="none";
	/// UNCLE 1
	if (udata.Num_uncles<1) return;
	elt=document.getElementById("uncle1_block_num");			elt.innerHTML=udata.Uncle1.Number
	elt=document.getElementById("uncle1_parent_num");			elt.innerHTML=udata.Uncle1.Parent_num
	elt=document.getElementById("uncle1_hash");					elt.innerHTML=udata.Uncle1.Hash
	elt=document.getElementById("uncle1_parent_hash");			elt.innerHTML=udata.Uncle1.Parent_hash
	elt=document.getElementById("uncle1_timestamp");			elt.innerHTML=udata.Uncle1.Timestamp
	elt=document.getElementById("uncle1_datetime");
	var date = new Date(udata.Uncle1.Timestamp*1000);
    var date_str=('0' + date.getDate()).slice(-2) + '/' + ('0' + (date.getMonth() + 1)).slice(-2) + '/' + date.getFullYear() + ' ' + ('0' + date.getHours()).slice(-2) + ':' + ('0' + date.getMinutes()).slice(-2) + ':' +('0' + date.getSeconds()).slice(-2);
	elt.innerHTML=date_str
	elt=document.getElementById("uncle1_miner");				elt.innerHTML=udata.Uncle1.Miner
	elt=document.getElementById("uncle1_difficulty");			elt.innerHTML=udata.Uncle1.Difficulty
	elt=document.getElementById("uncle1_total_difficulty");		elt.innerHTML=udata.Uncle1.Total_difficulty
	elt=document.getElementById("uncle1_gas_used");				elt.innerHTML=udata.Uncle1.Gas_used
	elt=document.getElementById("uncle1_gas_limit");			elt.innerHTML=udata.Uncle1.Gas_limit
	elt=document.getElementById("uncle1_nonce");				elt.innerHTML=udata.Uncle1.Nonce
	elt=document.getElementById("uncle1_sha3uncles");			elt.innerHTML=udata.Uncle1.Sha3uncles
	elt=document.getElementById("uncle1_extra_data");			elt.innerHTML=udata.Uncle1.Extra_data
	elt=document.getElementById("uncle1");						elt.style.display="block"
	if (udata.Num_uncles<2) return

	/// UNCLE 2
	elt=document.getElementById("uncle2_block_num");			elt.innerHTML=udata.Uncle2.Number
	elt=document.getElementById("uncle2_parent_num");			elt.innerHTML=udata.Uncle2.Parent_num
	elt=document.getElementById("uncle2_hash");					elt.innerHTML=udata.Uncle2.Hash
	elt=document.getElementById("uncle2_parent_hash");			elt.innerHTML=udata.Uncle2.Parent_hash
	elt=document.getElementById("uncle2_timestamp");			elt.innerHTML=udata.Uncle2.Timestamp
	elt=document.getElementById("uncle2_datetime");
	var date = new Date(udata.Uncle1.Timestamp*1000);
    var date_str=('0' + date.getDate()).slice(-2) + '/' + ('0' + (date.getMonth() + 1)).slice(-2) + '/' + date.getFullYear() + ' ' + ('0' + date.getHours()).slice(-2) + ':' + ('0' + date.getMinutes()).slice(-2) + ':' +('0' + date.getSeconds()).slice(-2);
	elt.innerHTML=date_str
	elt=document.getElementById("uncle2_miner");				elt.innerHTML=udata.Uncle2.Miner
	elt=document.getElementById("uncle2_difficulty");			elt.innerHTML=udata.Uncle2.Difficulty
	elt=document.getElementById("uncle2_total_difficulty");		elt.innerHTML=udata.Uncle2.Total_difficulty
	elt=document.getElementById("uncle2_gas_used");				elt.innerHTML=udata.Uncle2.Gas_used
	elt=document.getElementById("uncle2_gas_limit");			elt.innerHTML=udata.Uncle2.Gas_limit
	elt=document.getElementById("uncle2_nonce");				elt.innerHTML=udata.Uncle2.Nonce
	elt=document.getElementById("uncle2_sha3uncles");			elt.innerHTML=udata.Uncle2.Sha3uncles
	elt=document.getElementById("uncle2_extra_data");			elt.innerHTML=udata.Uncle2.Extra_data
	elt=document.getElementById("uncle2");						elt.style.display="block"
}
function load_block_transactions(transactions) {
	var table=document.getElementById("block_transaction_table")
	if (!table) {
		console.log("missing element `transaction_table`")
		return
	}
	var elt=document.getElementById("block_transactions_empty")
	if (transactions) {
		elt.style.display="none"
		table.style.display="block";
	} else {
		elt.style.display="block";
		table.style.display="none";
		return;
	}
	var elts=table.getElementsByTagName("TBODY")
	var tbody=elts[0]
	while (tbody.lastElementChild) {
		tbody.removeChild(tbody.lastElementChild)
	}
	for (i in transactions) {
		var a,link_text
		tx=transactions[i]
		tr=document.createElement('TR')
		if ((i%2)==1) {
			tr.style.backgroundColor="#eeeeee"
		}
		if (tx.Vm_error.length>0) {
			tr.style.backgroundColor="#de5d5d"
		}
		// TX Number
		td=document.createElement('TD')
		a=document.createElement('A')
		link_text = document.createTextNode(tx.Tx_index+1);
		a.appendChild(link_text);
		a.title = tx.Tx_hash
		a.className="link"
		a.href = '/index.html?'+tx.Tx_hash
		a.dataset.search_term=tx.Tx_hash
		a.addEventListener("click",short_search)
		td.appendChild(a)
		td.className="tx_td"
		tr.appendChild(td)
		// From addr
		td=document.createElement('TD')
		td.innerHTML='<a class="link" href="javascript: search(\''+tx.From_addr+'\')">'+format_address(tx.From_addr)+'</a';
		td.className="tx_td"
		tr.appendChild(td)
		// To addr
		td=document.createElement('TD')
		td.innerHTML='<a class="link" href="javascript: search(\''+tx.To_addr+'\')">'+format_address(tx.To_addr)+'</a';
		td.className="tx_td"
		tr.appendChild(td)
		// Value
		var val=format_value(tx.Value)
		// high part
		td=document.createElement('TD')
		td.className="tx_td_font value_hi"
		td.innerHTML=val.hi
		tr.appendChild(td)
		// comma
		td=document.createElement('TD')
		td.className="tx_td_font value_dot"
		td.innerHTML='.'
		tr.appendChild(td)
		// low part
		td=document.createElement('TD')
		td.className="tx_td_font value_lo"
		td.innerHTML=val.lo
		tr.appendChild(td)
		tbody.appendChild(tr)
	}
}
function load_account_transactions(table,tx_set) {

	var elts=table.getElementsByTagName("TBODY")
	var tbody=elts[0]
	while (tbody.lastElementChild) {
		tbody.removeChild(tbody.lastElementChild)
	}
	var transactions=tx_set.Transactions
	if (transactions==null) {
		return
	}
	for (i in transactions) {
		tx=transactions[i]
		tr=document.createElement('TR')
		if ((i%2)==1) {
			tr.style.backgroundColor="#eeeeee"
		}
		// Block
		td=document.createElement('TD')
		td.innerHTML='<a class="link" href="javascript: search(\''+tx.Block_num+'\')">'+tx.Block_num+'</a';
		td.className="tx_td"
		tr.appendChild(td)
		// From addr
		td=document.createElement('TD')
		td.innerHTML='<a class="link" href="javascript: search(\''+tx.From_addr+'\')">'+format_address(tx.From_addr)+'</a';
		td.className="tx_td"
		tr.appendChild(td)
		// To addr
		td=document.createElement('TD')
		td.innerHTML='<a class="link" href="javascript: search(\''+tx.To_addr+'\')">'+format_address(tx.To_addr)+'</a';
		td.className="tx_td"
		tr.appendChild(td)
		// Value
		var val=format_value(tx.Value)
		// high part
		td=document.createElement('TD')
		td.className="tx_td value_hi"
		td.innerHTML=val.hi
		tr.appendChild(td)
		// comma
		td=document.createElement('TD')
		td.className="tx_td value_dot"
		td.innerHTML='.'
		tr.appendChild(td)
		// low part
		td=document.createElement('Td')
		td.className="tx_td value_lo"
		td.innerHTML=val.lo
		tr.appendChild(td)

		td=document.createElement('TD')
		td.className="tx_td"
		td.innerHTML='<a href="javascript: search(\''+tx.Tx_hash+'\')"><img src="imgs/tx_details.png"></a>'
		tr.appendChild(td)

		tbody.appendChild(tr)
	}
	var elt;
	elt=document.getElementById("acct_transactions_nav")
	if (elt) {
		tx_set.method_name='get_account_transactions'
		add_page_navigation_elts(elt,tx_set,transactions.length)
	}
}
function load_account(account) {
	var elt=document.getElementById("account_info_account_address")
	elt.innerHTML=account.Address
	elt=document.getElementById("account_balance")
	var val_obj=format_value(account.Balance)
	elt.innerHTML='<span class="value_hi">'+val_obj.hi+'</span><span class="value_dot">.</span><span class="value_lo">'+val_obj.lo+ticker_symbol+'</span>'
	elt=document.getElementById("account_type") 
	if (account.Owner_id==0) {
		elt.innerHTML="Externally Owned Account (EOA)"
	} else {
		elt.innerHTML="Contract.<br/>Owner: "+'<a class="link" href="javascript: search(\''+account.Owner_address+'\')">'+account.Owner_address+'</a>';
	}
	elt=document.getElementById("account_num_tx")
	elt.innerHTML=account.Num_transactions

	var date = new Date(account.Ts_created*1000);
    var date_str=('0' + date.getDate()).slice(-2) + '/' + ('0' + (date.getMonth() + 1)).slice(-2) + '/' + date.getFullYear() + ' ' + ('0' + date.getHours()).slice(-2) + ':' + ('0' + date.getMinutes()).slice(-2) + ':' +('0' + date.getSeconds()).slice(-2);
	elt=document.getElementById("account_created")
	elt.innerHTML=date_str+', Block: '+account.Block_created

	elt=document.getElementById("account_deleted")
	if (account.Deleted==1) {
		elt.innerHTML='Yes'
	} else {
		elt.innerHTML='No'
	}
}
function load_account_value_transfers(table_element,vt_set) {


	var value_transfers=vt_set.Value_transfers
	var account_id=vt_set.Account_id
	var table=document.getElementById("account_value_transfers_table")
	if (!table) {
		console.log("missing element `account_value_transfers_table`")
		return
	}
	
	var elts=table.getElementsByTagName("TBODY")
	var tbody=elts[0]
	while (tbody.lastElementChild) {
		tbody.removeChild(tbody.lastElementChild)
	}
	if (value_transfers==null) {
		return;
	}
	for (i in value_transfers) {
		vt=value_transfers[i]
		tr=document.createElement('TR')
		if ((i%2)==1) {
			tr.style.backgroundColor="#eeeeee"
		}
		if (vt.Error.length>0) {
			tr.style.backgroundColor="#de5d5d"
		}
		var balance
		var value_color
		var val_elt
		if (vt.Direction==0) { // selftransfer
			balance=vt.To_balance
			value_color=''
		}  else {
			if (vt.Direction<0) {
				balance=vt.From_balance
				value_color='vt_out'
			} else {
				balance=vt.To_balance
				value_color='vt_in'
			}
		}
		// Block
		td=document.createElement('TD')
		a=document.createElement('A')
		link_text = document.createTextNode(vt.Block_num);
		a.appendChild(link_text);
		a.title = format_address4link(vt.Block_num)
		a.className="link"
		a.href = '/index.html?'+vt.Block_num
		a.dataset.search_term=vt.Block_num
		a.addEventListener("click",short_search)
		td.appendChild(a)
		td.className="vt_td"
		tr.appendChild(td)

		// From addr
		td=document.createElement('TD')
		a=document.createElement('A')
		link_text = document.createTextNode(format_address(vt.From_addr));
		a.appendChild(link_text);
		a.title = format_address4link(vt.From_addr)
		a.className="link"
		a.href = '/index.html?'+format_address4link(vt.From_addr)
		a.dataset.search_term=format_address4link(vt.From_addr)
		a.addEventListener("click",short_search)
		td.appendChild(a)
		td.className="vt_td"
		tr.appendChild(td)

		// To addr
		td=document.createElement('TD')
		a=document.createElement('A')
		link_text = document.createTextNode(format_address(vt.To_addr));
		a.appendChild(link_text);
		a.title = format_address4link(vt.To_addr)
		a.className="link"
		a.href = '/index.html?'+format_address4link(vt.To_addr)
		a.dataset.search_term=vt.To_addr
		a.addEventListener("click",short_search)
		td.appendChild(a)
		td.className="vt_td"
		tr.appendChild(td)

		// Kind
		var kind=lookup_kind(vt.Kind)
		td=document.createElement('TD')
		td.innerHTML=kind.code
		td.title=kind.description
		td.className="vt_td vt_td_kind"
		tr.appendChild(td)

		// Value
		var trsf_val=format_value(vt.Value)

		// High part
		td=document.createElement('TD')
		td.className="vt_td_font value_hi "+value_color
		td.innerHTML=trsf_val.hi
		tr.appendChild(td)
		// dot
		td=document.createElement('TD')
		td.className="vt_td_font value_dot "+value_color
		td.innerHTML='.'
		tr.appendChild(td)
		// Lo part
		td=document.createElement('TD')
		td.className="vt_td_font value_lo "+value_color
		td.innerHTML=trsf_val.lo
		tr.appendChild(td)

		// Balance
		var bal=format_value(balance)

		// High part
		td=document.createElement('TD')
		td.className="vt_td_font value_hi"
		td.innerHTML=bal.hi
		tr.appendChild(td)
		// dot
		td=document.createElement('TD')
		td.className="vt_td_font value_dot"
		td.innerHTML='.'
		tr.appendChild(td)
		// Lo part
		td=document.createElement('TD')
		td.className="vt_td_font value_lo"
		td.innerHTML=bal.lo
		tr.appendChild(td)

		tbody.appendChild(tr)
	}
	var elt;
	elt=document.getElementById("acct_value_transfers_nav")
	if (elt) {
		vt_set.method_name='get_account_value_transfers'
		add_page_navigation_elts(elt,vt_set,value_transfers.length)
	}
}
function load_transaction_value_transfers(table_element,vt_set) {

	show_section(3)

	var value_transfers=vt_set.Value_transfers

	var elts=table_element.getElementsByTagName("TBODY")
	var tbody=elts[0]
	while (tbody.lastElementChild) {
		tbody.removeChild(tbody.lastElementChild)
	}
	for (i in value_transfers) {
		vt=value_transfers[i]
		tr=document.createElement('TR')
		if ((i%2)==1) {
			tr.style.backgroundColor="#eeeeee"
		}
		if (vt.Error.length>0) {
			tr.style.backgroundColor="#de5d5d"
		}
		var balance
		var value_color=''
		var val_elt
		var a,link_text
		// Block
		td=document.createElement('TD')

		a=document.createElement('A')
		link_text = document.createTextNode(vt.Block_num);
		a.appendChild(link_text);
		a.title = 'Block '+vt.Block_num
		a.className="link"
		a.href = '/index.html?'+vt.Block_num
		a.dataset.search_term=vt.Block_num
		a.addEventListener("click",short_search)
		td.appendChild(a)
		td.className="vt_td"
		tr.appendChild(td)
		// From addr
		td=document.createElement('TD')

		a=document.createElement('A')
		link_text = document.createTextNode(format_address(vt.From_addr));
		a.appendChild(link_text);
		a.title = vt.From_addr 
		a.className="link"
		a.href = '/index.html?'+format_address4link(vt.From_addr)
		a.dataset.search_term=format_address4link(vt.From_addr)
		a.addEventListener("click",short_search)
		td.appendChild(a)
		td.className="vt_td"
		tr.appendChild(td)
		// To addr
		td=document.createElement('TD')

		a=document.createElement('A')
		link_text = document.createTextNode(format_address(vt.To_addr));
		a.appendChild(link_text);
		a.title = vt.To_addr 
		a.className="link"
		a.href = '/index.html?'+format_address4link(vt.To_addr)
		a.dataset.search_term=format_address4link(vt.To_addr)
		a.addEventListener("click",short_search)
		td.appendChild(a)

		td.className="vt_td"
		tr.appendChild(td)

		// Kind
		var kind=lookup_kind(vt.Kind)
		td=document.createElement('TD')
		td.innerHTML=kind.code
		td.title=kind.description
		td.className="vt_td vt_td_kind"
		tr.appendChild(td)

		add_value_to_table(tr,vt)

		tbody.appendChild(tr)
	}
}
function add_page_navigation_elts(elt,data_set,num_items) {
	var next_offset
	next_offset=(data_set.Offset+default_limit)
	var prev_btn=''
	if (data_set.Offset>0) {
		var new_offset=data_set.Offset-default_limit
		if (new_offset<0) {
			new_offset=0
		}
		prev_btn='<a class="navigator_btn prev_btn" href="javascript: '+data_set.method_name+'(\''+data_set.Account.Address+'\','+new_offset+','+default_limit+')"><img class="browser_nav_btns" src="imgs/previous.png"></a>';
	}
	var next_btn='';
	if (num_items==default_limit) {	
		next_btn='<a class="navigator_btn next_btn" href="javascript: '+data_set.method_name+'(\''+data_set.Account.Address+'\','+(data_set.Offset+default_limit)+','+default_limit+')"><img class="browser_nav_btns" src="imgs/next.png"></a>'
	}
	elt.innerHTML=prev_btn+next_btn
}
function load_single_transaction(account_address,transaction) {
	
	get_transaction_value_transfers(transaction.Tx_hash)

	var elt
	elt=document.getElementById("transaction_hash")
	if (elt) {
		elt.innerHTML=transaction.Tx_hash
	}
	elt=document.getElementById("transaction_block_num")
	if (elt) {
		elt.innerHTML=transaction.Block_num
	}
	elt=document.getElementById("transaction_from_addr")
	if (elt) {
		while (elt.lastElementChild) {
			elt.removeChild(elt.lastElementChild)
		}
		var a=document.createElement('A')
		var link_text = document.createTextNode(format_address4link(transaction.From_addr));
		a.appendChild(link_text);
		a.title = format_address4link(transaction.From_addr)
		a.className="link"
		a.href = '/index.html?'+format_address4link(transaction.From_addr)
		a.dataset.search_term=format_address4link(transaction.From_addr)
		a.addEventListener("click",short_search)
		elt.appendChild(a)
	}
	elt=document.getElementById("transaction_to_addr")
	if (elt) {
		while (elt.lastElementChild) {
			elt.removeChild(elt.lastElementChild)
		}
		var a=document.createElement('A')
		var link_text = document.createTextNode(format_address4link(transaction.To_addr));
		a.appendChild(link_text);
		a.title = format_address4link(transaction.To_addr)
		a.className="link"
		a.href = '/index.html?'+format_address4link(transaction.To_addr)
		a.dataset.search_term=format_address4link(transaction.To_addr)
		a.addEventListener("click",short_search)
		elt.appendChild(a)
	}
	elt=document.getElementById("transaction_value")
	if (elt) {
		var val=format_value(transaction.Value)
		elt.innerHTML='<span class="value_hi">'+val.hi+'</span>'+'<span class="value_dot">.</span>'+'<span class="value_lo">'+val.lo+ticker_symbol+'</span>';
	}
	elt=document.getElementById("transaction_status")
	if (elt) {
		elt.innerHTML=transaction.Tx_status
	}
	elt=document.getElementById("transaction_confirmations")
	if (elt) {
		elt.innerHTML=transaction.Confirmations
	}
	elt=document.getElementById("transaction_gas_limit")
	if (elt) {
		elt.innerHTML=transaction.Gas_limit
	}
	elt=document.getElementById("transaction_gas_used")
	if (elt) {
		elt.innerHTML=transaction.Gas_used
	}
	elt=document.getElementById("transaction_gas_price")
	if (elt) {
		var val=format_value(transaction.Gas_price)
		elt.innerHTML='<span class="value_hi">'+val.hi+'</span>'+'<span class="value_dot">.</span>'+'<span class="value_lo">'+val.lo+ticker_symbol+'</span>';
	}
	elt=document.getElementById("transaction_cost")
	if (elt) {
		var val=format_value(transaction.Cost)
		elt.innerHTML='<span class="value_hi">'+val.hi+'</span>'+'<span class="value_dot">.</span>'+'<span class="value_lo">'+val.lo+ticker_symbol+'</span>';
	}
	elt=document.getElementById("transaction_v")
	if (elt) {
		elt.innerHTML=transaction.V
	}
	elt=document.getElementById("transaction_r")
	if (elt) {
		elt.innerHTML=transaction.R
	}
	elt=document.getElementById("transaction_s")
	if (elt) {
		elt.innerHTML=transaction.S
	}
	elt=document.getElementById("transaction_error")
	if (elt) {
		elt.innerHTML=transaction.Vm_error
	}
}
function load_block_value_transfers(table_element,value_transfers) {

	var elts=table_element.getElementsByTagName("TBODY")
	var tbody=elts[0]
	while (tbody.lastElementChild) {
		tbody.removeChild(tbody.lastElementChild)
	}
	var i=0
	for (i in value_transfers) {
		var a,link_text
		vt=value_transfers[i]
		tr=document.createElement('TR')
		if ((i%2)==1) {
			tr.style.backgroundColor="#eeeeee"
		}
		if (vt.Error.length>0) {
			tr.style.backgroundColor="#de5d5d"
		}
		// Tx number
		td=document.createElement('TD')
		a=document.createElement('A')
		link_text = document.createTextNode(vt.Tx_index);
		a.appendChild(link_text);
		a.title = vt.Tx_hash
		a.className="link"
		a.href = '/index.html?'+vt.Tx_hash
		a.dataset.search_term=vt.Tx_hash
		a.addEventListener("click",short_search)
		td.appendChild(a)
		td.className="vt_td td_tx_index"
		tr.appendChild(td)

		td=document.createElement('TD')
		a=document.createElement('A')
		link_text = document.createTextNode(format_address(vt.From_addr));
		a.appendChild(link_text);
		a.title = vt.From_addr 
		a.className="link"
		a.href = '/index.html?'+format_address4link(vt.From_addr)
		a.dataset.search_term=format_address4link(vt.From_addr)
		a.addEventListener("click",short_search)
		td.appendChild(a)
		td.className="vt_td"
		tr.appendChild(td)
		// To addr
		td=document.createElement('TD')
		a=document.createElement('A')
		link_text = document.createTextNode(format_address(vt.To_addr));
		a.appendChild(link_text);
		a.title = vt.To_addr 
		a.className="link"
		a.href = '/index.html?'+format_address4link(vt.To_addr)
		a.dataset.search_term=format_address4link(vt.To_addr)
		a.addEventListener("click",short_search)
		td.appendChild(a)
		td.className="vt_td"
		tr.appendChild(td)
		// Kind
		var kind=lookup_kind(vt.Kind)
		td=document.createElement('TD')
		td.innerHTML=kind.code
		td.className="vt_td vt_td_kind"
		td.title=kind.description
		tr.appendChild(td)

		add_value_to_table(tr,vt)

		tbody.appendChild(tr)
	}
}
function add_value_to_table(tr,vt) {
		// Value
		val_entry=format_value(vt.Value)
		// higher value part
		td=document.createElement('TD')
		td.className="vt_td_font value_hi"
		if (vt.Error.length>0) {
			td.innerHTML='!'
		} else {
			td.innerHTML=val_entry.hi
		}
		tr.appendChild(td)

		// add decimal point now
		td=document.createElement('TD')
		td.className="vt_td_font value_dot"
		if (vt.Error.length>0) {
			td.innerHTML=''
		} else {
			td.innerHTML='.'
		}
		tr.appendChild(td)

		// add lower part of the value
		td=document.createElement('TD')
		td.className="vt_td_font value_lo"
		if (vt.Error.length>0) {
			var escaped_err=vt.Error.replace(/"/g, '\\\"');
			var a=document.createElement('A')
			a.className="link"
			a.dataset.error=vt.Error;
			a.addEventListener("click",show_error)
			a.innerHTML='ERROR'
			td.appendChild(a)
		} else {
			td.innerHTML=val_entry.lo
		}
		tr.appendChild(td)
}
function show_error(evt_obj) {
	
	evt_obj.preventDefault();
	let clicked_object = evt_obj.target;
	clicked_object.className="vt_td_font value_lo";
	var newstr=clicked_object.dataset.error
	newstr=newstr.split(',').join("<br/>\n");
	clicked_object.innerHTML="<code>"+newstr+"</code>";
	clicked_object.removeEventListener("click",show_error)
}
function search(search_text) {
	unselect_current_block()
	Ajax_GET('/search/'+search_text,function(data) {
		var result,response;
		response=JSON.parse(data)
		result=response.result
		search_callback(result)
	});
}
function exec_search() {
	var elt
	elt=document.getElementById("search_text")
	if (elt) {
		search(elt.value)
	}
	
}
function search_callback(result) {
	switch(result.Object_type) {
		case 0: // not found
			var elt=document.getElementById("object_not_found")
			elt.style.display="block";
			elt.innerHTML='"'+result.Search_text+"\" was not found"
			setTimeout(function() {
				var elt=document.getElementById("object_not_found")
					elt.style.display="none";
			},2000);
		break;
		case 1:	// block by number
			if (current_block_num!=-1) {
				block_element=find_block_in_line(current_block_num)
				if (block_element) {
					block_element.className="block_info link"
				}
			}
			load_block(result.Block)
			show_section(1)
			select_current_tab(1)
			select_and_show_pane(1)
		break;
		case 2: // account by address
			show_section(2)
			select_current_account_tab(2)
			select_and_show_account_pane(2)
			get_account_value_transfers(result.Search_text,0,default_limit)
			load_account(result.Account)

		break;
		case 3: // transaction by hash
			load_single_transaction(result.Search_text,result.Transaction)
			show_section(3)
		break;

		default:
			console.log("Unknown object type returned from EthBot: "+result.Object_type)
	}

}
function find_block_in_line(block_num) {
	var elt
	elt=document.getElementById("last_blocks_container")
	var elements
	elements=document.getElementsByName("b"+block_num)
	return elements[0]
}
function create_new_block_box(parent_container,block_info,new_block) {
	var block_elt
	block_elt=document.createElement("A")
	if (block_elt) {
			if (block_info.Block_number>highest_block_num) {
				highest_block_num=block_info.Block_number
			}
			block_elt.name="b"+block_info.Block_number
			block_elt.href="javascript: block_load_tab_data("+block_info.Block_number+")"
			block_elt.className="block_info link"
			if ((first_block_num_in_line>-1) && (current_block_num>-1) && (current_block_num==block_info.Block_number)) {
				block_elt.className="block_info link_selected"
			}
			block_elt.style.padding="0.5em";
			setTimeout(function() {
				block_elt.style.opacity=1;
			},10);
			if (new_block==true) {
				block_elt.style.background="yellow";
			}
			block_number_elt=document.createElement("DIV")
			block_number_elt.className="block_info_num_block"
			block_number_elt.innerHTML=block_info.Block_number
			num_tx_elt=document.createElement("DIV")
			num_tx_elt.className="block_info_num_transactions"
			num_tx_elt.innerHTML=block_info.Num_transactions
			block_elt.appendChild(block_number_elt)
			block_elt.appendChild(num_tx_elt)
			parent_container.insertBefore(block_elt,parent_container.firstChild)
	} else {
			console.log("failed to create new DIV")
	}
}
function set_last_blocks(last_blocks) {
	
	var elt
	elt=document.getElementById("last_blocks_container")
	if (!elt) {
		console.log("error: no last_blocks_container element found")
		return;
	}
	while (elt.lastElementChild) {
		elt.removeChild(elt.lastElementChild)
	}
	if (last_blocks.length>0) {
		first_block_num_in_line=last_blocks[0].Block_number;
	} else {
		first_block_num_in_line=-1
	}
	var block_info
	var i=last_blocks.length-1
	for (;i>-1;i--) {
		block_info=last_blocks[i]
		create_new_block_box(elt,block_info,false)
	}
}
function select_and_show_pane(tab_num) {
	var elt
	elt=document.getElementById("block_info")
	elt.style.display="none";
	if (tab_num==1) {
		elt.style.display="block"
	}
	elt=document.getElementById("block_transactions")
	elt.style.display="none";
	if (tab_num==2) {
		elt.style.display="block"
	}
	elt=document.getElementById("block_value_transfers")
	elt.style.display="none";
	if (tab_num==3) {
		elt.style.display="block"
	}
	elt=document.getElementById("block_uncles")
	elt.style.display="none"
	if (tab_num==4) {
		elt.style.display="block"
	}
	current_tab_num=tab_num
}
function select_current_tab(tab_num) {
	var elt
	elt=document.getElementById("tab_block_info")
	elt.className="block_tab tab_unselected"
	elt=document.getElementById("tab_block_transactions")
	elt.className="block_tab tab_unselected"
	elt=document.getElementById("tab_block_value_transfers")
	elt.className="block_tab tab_unselected"
	elt=document.getElementById("tab_block_uncles")
	elt.className="block_tab tab_unselected"
	switch (tab_num) {
		case 1:
			elt=document.getElementById("tab_block_info")
			elt.className="block_tab tab_selected"
		break;
		case 2:
			elt=document.getElementById("tab_block_transactions")
			elt.className="block_tab tab_selected"
		break;
		case 3:
			elt=document.getElementById("tab_block_value_transfers")
			elt.className="block_tab tab_selected"
		break;
		case 4:
			elt=document.getElementById("tab_block_uncles")
			elt.className="block_tab tab_selected"
		break;
	}
}
function tab_click(tab_num) {
	var elt=document.getElementById('block_children')
	elt.style.display='inline-block';
	show_section(1)
	select_current_tab(tab_num)
	select_and_show_pane(tab_num)
	switch (tab_num) {
		case 1:
			get_block(current_block_num)
		break;
		case 2:
			get_block_transactions(current_block_num)
		break;
		case 3:
			get_block_value_transfers(current_block_num)
		break;
		case 4:
			get_uncles(current_block_num)
		break;
		case 5:
			// nothing
		break;
	}

}
function lookup_kind(kind_code) {
	var kind={}
	switch(kind_code) {
		case 1:
			kind.code='GEN';
			kind.description='Initialization at Genesis block'
		break;
		case 2:
			kind.code='TXN'
			kind.description='Transfer between accounts'
		break;
		case 3:
			kind.code='FEE'
			kind.description='Transaction fee paid to the miner'
		break;
		case 4:
			kind.code='RWD'
			kind.description='Block reward paid to the miner'
		break;
		case 5:
			kind.code='NEW'
			kind.description='Contract creation operation'
		break
		case 6:
			kind.code='CTX'
			kind.description='Transfer of value made by a contract'
		break;
		case 7:
			kind.code='KIL'
			kind.description='Contract destruction'
		break;
		case 8:
			kind.code='FRK'
			kind.description='Transfer caused by a hard fork'
		break;
		default:
			kind.code='???'
			kind.description="Unknown value transfer kind"
	}
	return kind
}
function add_commas(input,skip) {
	var output=""
	var slen=input.length;
	var triplet_counter=0;
	for (var i=(slen-1);i>=0;i--) {
		output=input.charAt(i)+output
		triplet_counter++
		if (triplet_counter==3) {
			if (i==0) {	// skip < 0 means skip first comma
				// dont do it
			} else {
				if (i==(slen-1)) { // skip >0 means skip last comma
					// dont do it
				} else {
					output=","+output
					triplet_counter=0
				}
			}
		}
	}
	return output
}
function format_beautify_zeros(input) { 
	// cuts last zeros from the string as they look ugly
	var i=input.length-1
	for (;i>=0;i--) {
		if (input.charAt(i)!='0') {
			if (input.charAt(i)!=',') {
				break;
			}
		}
	}
	if (input.charAt(i)=='.') {
		i++
	}
	var output=input.substr(0,i+1)
	if (output.length==0) {
		output="0"
	}
	return output
}
function format_value(value) {
	var val_out=new Object()
	var len=value.length
	var value_str
	if (len<=18) {
		val_out.hi=0
		var new_str='000000000000000000'+value
		var completed_str=new_str.substr(-18,18)
		val_out.lo=format_beautify_zeros(completed_str)
	} else {
		var upper=value.substr(0,len-18)
		var lower=value.substr(-18,18)
		val_out.lo=format_beautify_zeros(lower)
		val_out.hi=add_commas(upper,1)
	}
	return val_out
}
function format_address(address) {
	if (address=="0") return 'BLOCKCHAIN';
	return address.substr(0,6)+'...'+address.substr(-6,6)
}
function format_address4link(address) {
	if (address=="0") {
		return "BLOCKCHAIN"
	} else {
		return address
	}
}
function format_big_number(big_number) {

	var output=""
	var slen=big_number.length;
	var abbrev_counter=0
	var i=(slen-1)
	for (;i>=0;i--) {
		if (((i%3)==0) && (i>0)) {
			abbrev_counter++
		} else {
			continue
		}
	}
	var abbrev_str=""
	switch(abbrev_counter) {
		case 1:
			// nothing
			abbrev_str='K'
		break;
		case 2:
			abbrev_str='M'
		break
		case 3:
			abbrev_str='G'
		break;
		case 4:
			abbrev_str='T'
		break;
		case 5:
			abbrev_str='P'
		break;
		case 6:
			abbrev_str='E'
		break;
	}
	var reminder=slen-(abbrev_counter*3)
	output=big_number.substr(0,reminder)+'.'+big_number.substr(reminder,2)+abbrev_str
	return output
}
function search_account(obj) {
	search(obj.innerHTML)
}
function show_section(section_num) {
	var elt

	elt=document.getElementById("block_children")
	if (section_num==1) {
		elt.style.display="inline-block"
	} else {
		elt.style.display="none"
	}
	elt=document.getElementById("account_info")
	if (section_num==2) {
		elt.style.display="inline-block"
	} else {
		elt.style.display="none"
	}
	elt=document.getElementById("single_transaction")
	if (section_num==3) {
		elt.style.display="inline-block"
	} else {
		elt.style.display="none"
	}
}
function select_current_account_tab(tab_num) {
	var elt
	elt=document.getElementById("tab_account_tx")
	elt.className="block_tab tab_unselected"
	elt=document.getElementById("tab_account_vt")
	elt.className="block_tab tab_unselected"
	switch (tab_num) {
		case 1:
			elt=document.getElementById("tab_account_tx")
			elt.className="block_tab tab_selected"
		break;
		case 2:
			elt=document.getElementById("tab_account_vt")
			elt.className="block_tab tab_selected"
		break;
	}
}
function select_and_show_account_pane(tab_num) {
	var elt
	elt=document.getElementById("account_transactions")
	elt.style.display="none";
	if (tab_num==1) {
		elt.style.display="block"
	}
	elt=document.getElementById("account_value_transfers")
	elt.style.display="none";
	if (tab_num==2) {
		elt.style.display="block"
	}
	current_account_tab_num=tab_num
}
function account_info_tab_click(tab_num) {
	select_current_account_tab(tab_num)
	select_and_show_account_pane(tab_num)
	switch (tab_num) {
		case 1:
			var elt=document.getElementById("account_address")
			get_account_transactions(elt.innerHTML,0,default_limit)
		break;
		case 2:
			var elt=document.getElementById("account_address")
			get_account_value_transfers(elt.innerHTML,0,default_limit)
		break;
	}
}
function set_block_header(block_num) {
	elt=document.getElementById("block_header_block_num")
	elt.innerHTML="Block " + block_num
}
function unselect_current_block() {
	var block_element
	if (current_block_num!=-1) {
		block_element=find_block_in_line(current_block_num)
		if (block_element) {
			block_element.className="block_info link"
		}
	}
}
function block_load_tab_data(block_num) {

	clear_new_block_flags()
	unselect_current_block()
	current_block_num=block_num
	
	var block_element
	block_element=find_block_in_line(block_num)
	block_element.className="block_info link_selected"
	if (current_tab_num==0) {
		current_tab_num=1
	}

	set_block_header(block_num)
	tab_click(current_tab_num)
}
function prev_block_line() {
	var new_block_offset
	if (first_block_num_in_line<0) {
		console.log('Error: no blocks loaded')
		return
	}
	new_block_offset=first_block_num_in_line-block_scrolling_num
	if (new_block_offset<0) {
		new_block_offset=0
	}
	get_last_blocks(new_block_offset)
}
function next_block_line() {
	var new_block_offset
	if (first_block_num_in_line<0) {
		console.log('Error: no blocks loaded')
		return
	}
	new_block_offset=first_block_num_in_line+block_scrolling_num
	if (new_block_offset<0) {
		new_block_offset=0
	}
	get_last_blocks(new_block_offset)
}
function clear_new_block_flags() {
	var container
	container=document.getElementById("last_blocks_container")
	var children=container.childNodes
	var i=0;
	for (;i<children.length;i++) {
		var elt=children[i]
		elt.style.removeProperty('background')
	}
}
function check_for_new_blocks() {
	if (highest_block_num==-1) {
		setTimeout(check_for_new_blocks,check_new_blocks_interval*1000)
		return; // something is wrong, block list should be loaded
	}
	if (first_block_num_in_line==-1) {
		return;
	}
	Ajax_GET('/nbe/'+highest_block_num,function(data) {
		var response=JSON.parse(data)
		if (response) {
			if (response.result) { // there are new blocks
				var new_block_limit
				new_block_limit=highest_block_num+block_scrolling_num
				if (new_block_limit<=first_block_num_in_line) {
					return;
				}
				Ajax_GET('/blist/'+new_block_limit,function(data) {
					var response=JSON.parse(data)
					if (response) {
						insert_new_blocks_in_front(response.result)
					}
				});
				
			}
		}
	});
	setTimeout(check_for_new_blocks,check_new_blocks_interval*1000)
}
function insert_new_blocks_in_front(new_blocks) {
	
	var container
	container=document.getElementById("last_blocks_container")
	if (!container) {
		console.log("error: no last_blocks_container element found")
		return;
	}
	var block_info
	var i=new_blocks.length-1
	for (;i>-1;i--) {
		block_info=new_blocks[i]
		if (block_info.Block_number>highest_block_num) {
			highest_block_num=block_info.Block_number
		}
		if (block_info.Block_number<=first_block_num_in_line) {
			continue
		}
		var block_out=container.lastChild;
		container.removeChild(block_out)
		create_new_block_box(container,block_info,true)
	}
	if (new_blocks.length>0) {
		first_block_num_in_line=new_blocks[0].Block_number;
	} else {
		first_block_num_in_line=-1
	}
}
