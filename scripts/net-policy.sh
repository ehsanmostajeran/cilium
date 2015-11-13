#!/usr/bin/env bash
set -e
set -o nounset
dir=$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )
cd "${dir}/.."

if [ ${#} -lt 3 ]; then
    echo "Usage: net-policy.sh SRC-SVC DST-SVC { allow | drop }"
    exit 0
fi

src="${1}"
dst="${2}"
policy="${3}"

if [ "${src}" != "any" ]; then
    src=$(./scripts/svc2grp.sh "${src}") || {
        echo "Unknown source service ${1}"
        exit 1
    }
fi

if [ "${dst}" != "any" ]; then
    dst=$(./scripts/svc2grp.sh "${dst}") || {
        echo "Unknown destination service ${2}"
        exit 1
    }
fi

if [ "${policy}" != "allow" -a  "${policy}" != "drop" ]; then
    echo "Unknown policy action ${policy}"
    exit 1
fi


echo "Applying policy: ${1} (${src}) <=> ${2} (${dst}): ${policy}"

sudo ./backend/apply-net-policy.sh "${src}" "${dst}" "${policy}"
vagrant ssh -c "sudo cilium/backend/apply-net-policy.sh ${src} ${dst} ${policy}" node1
vagrant ssh -c "sudo cilium/backend/apply-net-policy.sh ${src} ${dst} ${policy}" node2

exit 0
