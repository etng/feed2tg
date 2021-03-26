SHELL := /bin/bash
test:
	# rm mine.opml || true
	source .env && go run main.go --tg_token=$${TG_TOKEN} --tg_channel=$${TG_CHANNEL} --proxy=$${PROXY}