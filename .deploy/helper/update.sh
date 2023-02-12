#!/bin/bash

ord=(
"1.2.somename.net"
"3.4.somename.net"
)

ams=(
"5.6.somename.net"
"7.8.somename.net"
)

execute_commands()
{
    servers=("$@")
    for i in "${servers[@]}"; do
        printf "===   $i \n"
        # config file
        # ssh adbroker@${i} "mkdir ~/log/notification-service/"

        # scp adbroker@${i}:~/apps/notification-service/notification-service.toml ${i}.toml
        # scp ${i}.toml adbroker@${i}:~/apps/notification-service/notification-service.toml

        # ssh adbroker@${i} "chmod 777 ~/log/notification-service/"

        # ssh adbroker@${i} "supervisorctl status tjdsp:"

        # scp ${i}.toml adbroker@${i}:~/apps/tj-dsp/tj-dsp.toml
    done
}

printf "=== ORD ===\n"
# update config file to set datacenter to ORD
# sed -i 's/DataCenter=./DataCenter=1/' config.toml
# execute_commands "${ord[@]}"

printf "=== AMS ===\n"
# update config file to set datacenter to AMS
# sed -i 's/DataCenter=./DataCenter=2/' config.toml
execute_commands "${ams[@]}"