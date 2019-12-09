#!/bin/sh

# Detect Hardware that will not change state

SYSTTEM_INFO=$(dmidecode -t system)

MANUFACTURER=$(echo "${SYSTTEM_INFO}" | grep Manufacturer | awk -F': ' '{print $2}')
PRODUCT_NAME=$(echo "${SYSTTEM_INFO}" | grep Product\ Name | awk -F': ' '{print $2}')
SERIAL_NUMBER=$(echo "${SYSTTEM_INFO}" | grep Serial\ Number | awk -F': ' '{print $2}')
SKU_NUMBER=$(echo "${SYSTTEM_INFO}" | grep SKU\ Number | awk -F': ' '{print $2}')
FAMILY=$(echo "${SYSTTEM_INFO}" | grep Family | awk -F': ' '{print $2}')

SYSTEM_TYPE=$(cat /proc/cpuinfo  | grep -o hypervisor | head -n1)
if [ "${SYSTEM_TYPE}" == "" ]; then
	SYSTEM_TYPE="bare-metal"
fi

GPUS_PCI_ADDRESS=$(lspci | grep -i 'vga' | awk '{print $1}')
for GPU_PCI_ADDRESS in ${GPUS_PCI_ADDRESS}; do
	GPU=$(lspci -v -s ${GPU_PCI_ADDRESS} | grep Kernel | awk -F ': ' '{print $2}')
	GRAPHICS="${GRAPHICS} \"graphics.$GPU=true\""
done

if [ "${GPUS}" != "" ]; then
	VIDEO="\"VIDEO=default\""
fi

VT_ENABLED_COUNT=$(egrep -c '(vmx)' /proc/cpuinfo)
if [ ${VT_ENABLED_COUNT} -gt 0 ]; then
	VT_ENABLED="true"
else
	VT_ENABLED="false"
fi

echo -e "{
	\"labels\": [\"system_type=${SYSTEM_TYPE}\", \"vt_enabled=${VT_ENABLED}\", \"manufacturer=${MANUFACTURER}\", \"product_name=${PRODUCT_NAME}\", \"serial_number=${SERIAL_NUMBER}\", \"sku_number=${SKU_NUMBER}\", \"family=${FAMILY}\", ${GRAPHICS}],
	\"node-generic-resources\": [${VIDEO}]
}" > /etc/docker/daemon.json

# until $FINISHED; do
	# echo ""
	# sleep 5
# done
