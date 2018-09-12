echo "[Unit]
Description=RenEx's Swapper Daemon
After=network.target

[Service]
ExecStart=${HOME}/.swapper/bin/swapper --config ${HOME}/.swapper/config.json --keystore ${HOME}/.swapper/keystore.json
Restart=on-failure
StartLimitBurst=0

# Specifies which signal to use when killing a service. Defaults to SIGTERM.
# SIGHUP gives parity time to exit cleanly before SIGKILL (default 90s)
KillSignal=SIGHUP

[Install]
WantedBy=default.target" >> swapper.service

mv swapper.service /etc/