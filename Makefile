build:
	go build

run:
	go build && DATA_FOLDER=./data TOKEN=YourTelegramBotToken CHECKING_INTERVAL=10 ./bazaraki_notifier
