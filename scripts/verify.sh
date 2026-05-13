#!/bin/bash
set -e

NETWORK="${1:-testnet}"

echo "=== Moistello Contract Verification ==="
echo "Network: $NETWORK"
echo ""

CONTRACTS=(
    "circle_factory"
    "circle"
    "reputation_registry"
    "governance_token"
    "treasury"
)

WASM_DIR="../moistello-contracts/target/wasm32v1-none/release"

for contract in "${CONTRACTS[@]}"; do
    wasm_path="${WASM_DIR}/${contract}.wasm"
    if [ -f "$wasm_path" ]; then
        size=$(ls -lh "$wasm_path" | awk '{print $5}')
        echo "  ✓ $contract.wasm ($size)"
    else
        echo "  ✗ $contract.wasm — NOT FOUND"
    fi
done

echo ""
echo "To deploy contracts:"
echo "  cd ../moistello-contracts && make deploy-testnet"
echo ""
echo "To verify on-chain:"
echo "  soroban contract read --id <CONTRACT_ID> --network $NETWORK"
