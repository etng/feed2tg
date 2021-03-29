SHELL := /bin/bash
only_tg:
	# rm mine.opml || true
	source .env && go run main.go --tg_token=$${TG_TOKEN} --tg_channel=$${TG_CHANNEL} --proxy=$${PROXY}
no_notify:
	# rm mine.opml || true
	source .env && go run main.go --proxy=$${PROXY}
only_pp:
	# rm mine.opml || true
	source .env && go run main.go -pp_token=$${PP_TOKEN} -pp_topic=$${PP_TOPIC} --proxy=$${PROXY}