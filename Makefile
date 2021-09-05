install:
	cp pm5.yml /etc
	cp pm5.service /etc/systemd/system
	systemctl daemon-reload
	systemctl enable pm5
	systemctl start pm5

.PHONY install
