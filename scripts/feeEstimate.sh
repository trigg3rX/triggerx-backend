ABI='[
                {
                  "inputs": [
                    { "internalType": "address", "name": "safeAddress", "type": "address" },
                    { "internalType": "address", "name": "actionTarget", "type": "address" },
                    { "internalType": "uint256", "name": "actionValue", "type": "uint256" },
                    { "internalType": "bytes", "name": "actionData", "type": "bytes" },
                    { "internalType": "uint8", "name": "operation", "type": "uint8" }
                  ],
                  "name": "execJobFromHub",
                  "outputs": [
                    { "internalType": "bool", "name": "success", "type": "bool" }
                  ],
                  "stateMutability": "nonpayable",
                  "type": "function"
                }
              ]'

ENCODED_ABI=$(python3 -c "import urllib.parse,sys; print(urllib.parse.quote(sys.argv[1]))" "$ABI")

curl "http://localhost:9002/api/fees?ipfs_url=https://ipfs.io/ipfs/bafkreidpwqyuev5vzpodttc4kt5tl6gk6ycjztacsya45ilhvx26s4ysgq&task_definition_id=2&target_chain_id=421614&target_contract_address=0xa0bC1477cfc452C05786262c377DE51FB8bc4669&target_function=execJobFromHub&abi=$ENCODED_ABI"