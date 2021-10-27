.PHONY: setup
setup:
	cp pm5.service /etc/systemd/system
	systemctl daemon-reload
	systemctl enable pm5
	systemctl start pm5

.PHONY: build-pi4
build-pi4:
	GOOS=linux GOARCH=arm GOARM=7 go build
	
.PHONY: build-pi3
build-pi3:
	GOOS=linux GOARCH=arm GOARM=7 go build

.PHONY: build-pi0
build-pi0:
	GOOS=linux GOARCH=arm GOARM=6 go build

.PHONY: deploy
deploy:
	ssh pi@pm5 'sudo systemctl stop pm5'
	scp pm5 pi@pm5:/home/pi
	ssh pi@pm5 'sudo systemctl start pm5'
