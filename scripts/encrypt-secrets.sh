#!/bin/sh
set -eu
root=$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)
input=${1:-"$root/.env.jury"}
if ! command -v age >/dev/null 2>&1; then
    echo "Install age: https://github.com/FiloSottile/age#installation" >&2
    exit 1
fi
if [ ! -f "$input" ]; then
    echo "Create .env.jury from .env.example and add only jury configuration/keys." >&2
    exit 1
fi
if git -C "$root" ls-files --error-unmatch "$input" >/dev/null 2>&1; then
    echo "Plaintext environment must not be tracked by Git." >&2
    exit 1
fi
age -r age1932fu7f70rwyfhaxflyfszz4gf6avxwc95wc0yjwm6zns9tspycqpgdnc0 -o "$root/secrets.env.age" "$input"
echo "Created secrets.env.age. Decrypt with the organizer private key; plaintext was not printed."
