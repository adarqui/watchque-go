all:
	make clean
	go build

deps:
	go get github.com/adarqui/fsnotify
	go get github.com/gosexy/redis
	go get github.com/adarqui/fsmonitor

clean:
	rm -f watchque-go

test_pre:
	./test/test_pre

test_post:
	./test/test_post

test:
	./test/challenge.sh

run_redis:
	make
	./watchque-go 127.0.0.1:6379 --debug=on "CreateClass:CreateQueue:c:/tmp/wgo-test/c" "UpdateClass:UpdateQueue:u:/tmp/wgo-test/u" "DeleteClass:DeleteQueue:d:/tmp/wgo-test/d" "AllClass:AllQueue:a:/tmp/wgo-test/a" "AllClass2:AllQueue2:cudr:/tmp/wgo-test/cudr" "RenameClass:RenameQueue:r:/tmp/wgo-test/r"

run_redis_nodebug:
	make
	./watchque-go 127.0.0.1:6379 "CreateClass:CreateQueue:c:/tmp/wgo-test/c" "UpdateClass:UpdateQueue:u:/tmp/wgo-test/u" "DeleteClass:DeleteQueue:d:/tmp/wgo-test/d" "AllClass:AllQueue:a:/tmp/wgo-test/a" "AllClass2:AllQueue2:cudr:/tmp/wgo-test/cudr" "RenameClass:RenameQueue:r:/tmp/wgo-test/r"

run_local:
	make
	./watchque-go /tmp/wgo-test/bin --debug=on "CreateClass:CreateQueue:c:/tmp/wgo-test/c" "UpdateClass:UpdateQueue:u:/tmp/wgo-test/u" "DeleteClass:DeleteQueue:d:/tmp/wgo-test/d" "AllClass:AllQueue:a:/tmp/wgo-test/a" "AllClass2:AllQueue2:cudr:/tmp/wgo-test/cudr" "RenameClass:RenameQueue:r:/tmp/wgo-test/r"
