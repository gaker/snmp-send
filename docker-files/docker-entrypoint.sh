#!/bin/sh

if [ ! -f "/etc/snmp-send/config.json" ]; then
    echo "Creating Config file from Environment Variables"
    export IPADDR=$(ifconfig eth0 | grep "inet addr:" | cut -d ':' -f2 | cut -d " " -f1)

    sed -e "s/{{snmp_community}}/${SNMP_COMMUNITY}/g" \
        -e "s/{{snmp_timeout}}/${SNMP_TIMEOUT}/g" \
        -e "s/{{hostname}}/${HOSTNAME}/g" \
        -e "s/{{ip_addr}}/${IPADDR}/g" \
        -e "s/{{server_type}}/${SERVER_TYPE}/g" \
        -e "s/{{influx_db}}/${INFLUX_DB}/g" \
        -e "s#{{receiver_url}}#${RECEIVER_URL}#g" \
        -e "s/{{receiver_token}}/${RECEIVER_TOKEN}/g" /etc/snmp-send/example.config.json > /etc/snmp-send/config.json
fi

/sbin/cat <<EOF > /tmp/snmp-cron
    ${CRON_TIME} /bin/snmp_send -conf=/etc/snmp-send/config.json >> /var/log/cron.log 2>&1
EOF

/sbin/crontab -c /var/spool/cron/crontabs /tmp/snmp-cron 

if [ "$1" == "" ]; then
    echo "Running crond"
    cat /etc/snmp-send/config.json
    /sbin/crond -L /var/log/cron.log && tail -f /var/log/cron.log
else
    eval "$@"
fi

