#!/bin/bash
set -eo pipefail

# This file initializes muliple validators and a full-node for remote testnet purposes.

source "./genesis.sh"

CHAIN_ID="dydxprotocol-testnet"

# Define mnemonics for all validators.
MNEMONICS=(
	# alice
	# Consensus Address: dydxvalcons1zf9csp5ygq95cqyxh48w3qkuckmpealrw2ug4d
	"merge panther lobster crazy road hollow amused security before critic about cliff exhibit cause coyote talent happy where lion river tobacco option coconut small"

	# bob
	# Consensus Address: dydxvalcons1s7wykslt83kayxuaktep9fw8qxe5n73ucftkh4
	"color habit donor nurse dinosaur stable wonder process post perfect raven gold census inside worth inquiry mammal panic olive toss shadow strong name drum"

	# carl
	# Consensus Address: dydxvalcons1vy0nrh7l4rtezrsakaadz4mngwlpdmhy64h0ls
	"school artefact ghost shop exchange slender letter debris dose window alarm hurt whale tiger find found island what engine ketchup globe obtain glory manage"

	# dave
	# Consensus Address: dydxvalcons1stjspktkshgcsv8sneqk2vs2ws0nw2wr272vtt
	"switch boring kiss cash lizard coconut romance hurry sniff bus accident zone chest height merit elevator furnace eagle fetch quit toward steak mystery nest"
)

# Define node keys for all full nodes.
FULL_NODE_KEYS=(
	# Node ID: dfa67970296bbecce14daba6cb0da516ed60458a
	"+c9Wyy9G4VJvVmUQ41CogREJPVMDqnBxefcGoika3Qo7U7eJHVIcjPIFuS0HYm224mWMfYgdNlo5KgJ0z1x/0w=="

	# Node ID: 25dd504d86d82673b9cf94fe78c00714f236c9f8
	"6pJcb5ezfttShtuAsPfOVv5Ua4h3MdZWzHvBnLC3w3f2vdhHiKelbWsVtkzIt6WF475k+5me4n6ptiz99WxkIw=="

	# Node ID: 5b0bdffc54d3aa942ab8abc636bd9cfd0e835709
	"HU8oEKbQU5SgIUosbIPr/WBcW+LW/39eFUo1mEQRNVb3VwrJNbaG7It7hCR+6+Jc9y9IN8QIx7011zV66NDevw=="

	# Node ID: c026a2f137552f26867fb90e7f6a025d44a7781f
	"+dGignS2RfEyWDp39HFxlKZ6h9mgvRrFYEyo8aRDW+PN0XTMU5KrSD6B+1sE7/uAeDsvcth6th+6maHisRMDRg=="
)

# Define node keys for all validators.
NODE_KEYS=(
	# Node ID: 17e5e45691f0d01449c84fd4ae87279578cdd7ec
	"8EGQBxfGMcRfH0C45UTedEG5Xi3XAcukuInLUqFPpskjp1Ny0c5XvwlKevAwtVvkwoeYYQSe0geQG/cF3GAcUA=="

	# Node ID: b69182310be02559483e42c77b7b104352713166
	"3OZf5HenMmeTncJY40VJrNYKIKcXoILU5bkYTLzTJvewowU2/iV2+8wSlGOs9LoKdl0ODfj8UutpMhLn5cORlw=="

	# Node ID: 47539956aaa8e624e0f1d926040e54908ad0eb44
	"tWV4uEya9Xvmm/kwcPTnEQIV1ZHqiqUTN/jLPHhIBq7+g/5AEXInokWUGM0shK9+BPaTPTNlzv7vgE8smsFg4w=="

	# Node ID: 5882428984d83b03d0c907c1f0af343534987052
	"++C3kWgFAs7rUfwAHB7Ffrv43muPg0wTD2/UtSPFFkhtobooIqc78UiotmrT8onuT1jg8/wFPbSjhnKRThTRZg=="
)

# Define monikers for each validator. These are made up strings and can be anything.
# This also controls in which directory the validator's home will be located. i.e. `/dydxprotocol/chain/.alice`
MONIKERS=(
	"alice"
	"bob"
	"carl"
	"dave"
)

# Define all test accounts for the chain.
TEST_ACCOUNTS=(
	"dydx199tqg4wdlnu4qjlxchpd7seg454937hjrknju4" # alice
	"dydx10fx7sy6ywd5senxae9dwytf8jxek3t2gcen2vs" # bob
	"dydx1fjg6zp6vv8t9wvy4lps03r5l4g7tkjw9wvmh70" # carl
	"dydx1wau5mja7j7zdavtfq9lu7ejef05hm6ffenlcsn" # dave
	"dydx1x2hd82qerp7lc0kf5cs3yekftupkrl620te6u2"
	"dydx1z9vz6r35ejc2l7l4wse8sc08gq2rm99t9c2hwr"
	"dydx16flrr9x7wjypd8y38dqt93km7tmnhp3yx6gucx"
	"dydx1ux65hjecs2cc64sw46dy5tr39s9xay5a7xz868"
	"dydx13yrk2f4h568j8svst3nvrcf0utkumjnzrc822g"
	"dydx1rkkrvxxvnufkwq9dve63getfyr2crpkgkqzrxy"
	"dydx1ruvdhudz4tcvm9rxm3rv7npxqdkzmw30lacxzd"
	"dydx1vewh46wf76szcnlajhxythaj4nja87vgjsez40"
	"dydx1zhr6cstar6zlr0p3r70u7nt7vy0ag8er786wts"
	"dydx1azcv6c9uqvyumxp0tq4efzyufrqft3vzpa3f70"
	"dydx1ajyappxvuhwvm63vy9uk4tgd6n2nqsx5v09783"
	"dydx1xm4mn2v5mquz5r0kth4l0s0pd8lgsl6ka9nwcd"
	"dydx18gtyjf5jrnrhd2y837p7yg60n4e94kpfpmq49a"
	"dydx1xu782jvc9k7wkxsdyckanlqxcy6v2vd7ups69u"
	"dydx1rsk4kt2xw5yfdxwc55tsphj8g2yyajxpsnsh8f"
	"dydx1sje4za9ww22czp0707ycyev72pfexzlgnpg8vp"
	"dydx1ral6kmw4jpxjj3hcjxlqwxj8qek2440p2h9cca"
	"dydx12qxafhmlcwr6mtdp24sn974qtfd32ffhuq6fjp"
	"dydx16yhyu378me5enwlevsl7f9atumg5mpgvrzj5dm"
	"dydx1ey95wck3szcm8dh96gjn0w5lmv4alwpah0zxmn"
)

FAUCET_ACCOUNTS=(
	"dydx1nzuttarf5k2j0nug5yzhr6p74t9avehn9hlh8m" # main faucet
	"dydx10du0qegtt73ynv5ctenh565qha27ptzr6dz8c3" # backup #1
	"dydx1axstmx84qtv0avhjwek46v6tcmyc8agu03nafv" # backup #2
)

# Define dependencies for this script.
# `jq` and `dasel` are used to manipulate json and yaml files respectively.
install_prerequisites() {
	apk add dasel jq
}

# Create all validators for the chain including a full-node.
# Initialize their genesis files and home directories.
create_validators() {
	# Create directories for full-nodes to use.
	for i in "${!FULL_NODE_KEYS[@]}"; do
		FULL_NODE_HOME_DIR="$HOME/chain/.full-node-$i"
		FULL_NODE_CONFIG_DIR="$FULL_NODE_HOME_DIR/config"
		dydxprotocold init "full-node" -o --chain-id=$CHAIN_ID --home "$FULL_NODE_HOME_DIR"

		# Note: `dydxprotocold init` non-deterministically creates `node_id.json` for each validator.
		# This is inconvenient for persistent peering during testing in Terraform configuration as the `node_id`
		# would change with every build of this container.
		#
		# For that reason we overwrite the non-deterministically generated one with a deterministic key defined in this file here.
		new_file=$(jq ".priv_key.value = \"${FULL_NODE_KEYS[$i]}\"" "$FULL_NODE_CONFIG_DIR"/node_key.json)
		cat <<<"$new_file" >"$FULL_NODE_CONFIG_DIR"/node_key.json

		edit_config "$FULL_NODE_CONFIG_DIR"
	done

	# Create temporary directory for all gentx files.
	mkdir /tmp/gentx

	# Iterate over all validators and set up their home directories, as well as generate `gentx` transaction for each.
	for i in "${!MONIKERS[@]}"; do
		VAL_HOME_DIR="$HOME/chain/.${MONIKERS[$i]}"
		VAL_CONFIG_DIR="$VAL_HOME_DIR/config"

		# Initialize the chain and validator files.
		dydxprotocold init "${MONIKERS[$i]}" -o --chain-id=$CHAIN_ID --home "$VAL_HOME_DIR"

		# Overwrite the randomly generated `priv_validator_key.json` with a key generated deterministically from the mnemonic.
		dydxprotocold tendermint gen-priv-key --home "$VAL_HOME_DIR" --mnemonic "${MNEMONICS[$i]}"

		# Note: `dydxprotocold init` non-deterministically creates `node_id.json` for each validator.
		# This is inconvenient for persistent peering during testing in Terraform configuration as the `node_id`
		# would change with every build of this container.
		#
		# For that reason we overwrite the non-deterministically generated one with a deterministic key defined in this file here.
		new_file=$(jq ".priv_key.value = \"${NODE_KEYS[$i]}\"" "$VAL_CONFIG_DIR"/node_key.json)
		cat <<<"$new_file" >"$VAL_CONFIG_DIR"/node_key.json

		edit_config "$VAL_CONFIG_DIR"

		echo "${MNEMONICS[$i]}" | dydxprotocold keys add "${MONIKERS[$i]}" --recover --keyring-backend=test --home "$VAL_HOME_DIR"

		# Using "*" as a subscript results in a single arg: "dydx1... dydx1... dydx1..."
		# Using "@" as a subscript results in separate args: "dydx1..." "dydx1..." "dydx1..."
		# Note: `edit_genesis` must be called before `add-genesis-account`.
		edit_genesis "$VAL_CONFIG_DIR" "${TEST_ACCOUNTS[*]}" "${FAUCET_ACCOUNTS[*]}"

		for acct in "${TEST_ACCOUNTS[@]}"; do
			dydxprotocold add-genesis-account "$acct" 100000000000000000usdc,100000000000stake --home "$VAL_HOME_DIR"
		done
		for acct in "${FAUCET_ACCOUNTS[@]}"; do
			dydxprotocold add-genesis-account "$acct" 900000000000000000usdc,100000000000stake --home "$VAL_HOME_DIR"
		done

		dydxprotocold gentx "${MONIKERS[$i]}" 500000000stake --moniker="${MONIKERS[$i]}" --keyring-backend=test --chain-id=$CHAIN_ID --home "$VAL_HOME_DIR"

		# Copy the gentx to a shared directory.
		cp -a "$VAL_CONFIG_DIR/gentx/." /tmp/gentx
	done

	# Copy gentxs to the first validator's home directory to build the genesis json file
	FIRST_VAL_HOME_DIR="$HOME/chain/.${MONIKERS[0]}"
	FIRST_VAL_CONFIG_DIR="$FIRST_VAL_HOME_DIR/config"

	rm -rf "$FIRST_VAL_CONFIG_DIR/gentx"
	mkdir "$FIRST_VAL_CONFIG_DIR/gentx"
	cp -r /tmp/gentx "$FIRST_VAL_CONFIG_DIR"

	# Build the final genesis.json file that all validators and the full-nodes will use.
	dydxprotocold collect-gentxs --home "$FIRST_VAL_HOME_DIR"

	# Copy this genesis file to each of the other validators
	for i in "${!MONIKERS[@]}"; do
		if [[ "$i" == 0 ]]; then
			# Skip first moniker as it already has the correct genesis file.
			continue
		fi

		VAL_HOME_DIR="$HOME/chain/.${MONIKERS[$i]}"
		VAL_CONFIG_DIR="$VAL_HOME_DIR/config"
		rm -rf "$VAL_CONFIG_DIR/genesis.json"
		cp "$FIRST_VAL_CONFIG_DIR/genesis.json" "$VAL_CONFIG_DIR/genesis.json"
	done

	# Copy the genesis file to the full-node directories.
	for i in "${!FULL_NODE_KEYS[@]}"; do
		FULL_NODE_HOME_DIR="$HOME/chain/.full-node-$i"
		FULL_NODE_CONFIG_DIR="$FULL_NODE_HOME_DIR/config"

		cp "$FIRST_VAL_CONFIG_DIR/genesis.json" "$FULL_NODE_CONFIG_DIR/genesis.json"
	done
}

# TODO(DEC-1894): remove this function once we migrate off of persistent peers.
# Note: DO NOT add more config modifications in this method. Use `cmd/config.go` to configure
# the default config values.
edit_config() {
	CONFIG_FOLDER=$1

	# Disable pex
	dasel put bool -f "$CONFIG_FOLDER"/config.toml '.p2p.pex' 'false'

	# TODO(CORE-79): Set this parameter in the binary after we get a reasonable value from experiments.
	dasel put string -f "$CONFIG_FOLDER"/config.toml '.consensus.timeout_commit' '1s'
}

install_prerequisites
create_validators
