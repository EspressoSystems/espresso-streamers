#!/usr/bin/env bash
function generate_go_bindings() {
  local abi_file="$1"
  local contract_name_full
  local contract_name
  local base_name

  if [[ -z "$abi_file" ]]; then
    echo "Error: Please provide the path to the ABI .json file." >&2
    return 1
  fi

  # If exact file doesn't exist, try to find a versioned variant.
  if [[ ! -f "$abi_file" ]]; then
    local dir name versioned_file
    dir=$(dirname "$abi_file")
    name=$(basename "$abi_file" .json)
    versioned_file=$(ls "$dir/$name".*.json 2>/dev/null | sort -V | head -n1)
    if [[ -n "$versioned_file" ]]; then
      echo "Note: $abi_file not found, using versioned artifact: $versioned_file" >&2
      abi_file="$versioned_file"
    fi
  fi

  if [[ ! -f "$abi_file" ]]; then
    echo "Error: File not found: $abi_file" >&2
    return 1
  fi

  base_name=$(basename "$abi_file")
  contract_name_full="${base_name%.json}"
  contract_name="${contract_name_full#I}"   # Remove leading 'I' if present
  IFS='.' read -r contract_name _ <<< "$contract_name"

  abigen --abi "$abi_file" --pkg bindings --type "$contract_name"
  local abigen_status=$?
  if [[ $abigen_status -ne 0 ]]; then
    echo "Error running abigen for $contract_name (exit code: $abigen_status)" >&2
    return $abigen_status
  fi

  return 0
}

# Main execution block
if [[ $# -ne 1 ]]; then
  echo "Usage: $0 <path_to_abi_json>" >&2
  exit 1
fi

if bindings=$(generate_go_bindings "$1"); then
  echo "$bindings"
else
  exit 1
fi
