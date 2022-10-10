USER := pi
HOST := pm5
DIR := /home/pi

.PHONY: pi4
pi4:
	GOOS=linux GOARCH=arm GOARM=7 go build
	
.PHONY: pi3
pi3:
	GOOS=linux GOARCH=arm GOARM=7 go build

.PHONY: pi0
pi0:
	rm pm5
	GOOS=linux GOARCH=arm GOARM=6 go build

.PHONY: deploy
deploy:
	ssh $(USER)@$(HOST) "sudo systemctl stop pm5"
	scp pm5 $(USER)@$(HOST):$(DIR)
	ssh $(USER)@$(HOST) "sudo systemctl start pm5"

