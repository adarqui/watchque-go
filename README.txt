TODO:

	- Solid redis re-connect and maybe disabling of fsnotify events until reconnected


Check Makefile for usage.

Installing:

	export GOPATH=`pwd`
	make deps
	make
	make run_redis


IMPORTANT:

	Modified fsmonitor code (one import line). fsmonitor was pointing to the old code repo rather than the new fsnotify repo.


Some quick tests.

The first test is running ./test/challenge_async.sh 10000. This performs 6 simultaneous file operations (10k c, C, u, d, r, cudr, a). We then check the resque queues to see how many events got enqueue'd.

Event flags:

	c - create
	C - close write (somewhat like a create event)
	u - update/modify
	d - delete
	r - rename/moved
	a - all


Without debugging:

d@o:~/dev/watchque-go/test$ ./challenge_async.sh 10000
d@o:~/dev/watchque-go/test$ ./test_post
resque:queue:CreateQueue
(integer) 10000
resque:queue:DeleteQueue
(integer) 10000
resque:queue:UpdateQueue
(integer) 10000
resque:queue:RenameQueue
(integer) 10000
resque:queue:AllQueue
(integer) 39956
resque:queue:AllQueue2
(integer) 39968


With --debug=on:

d@o:~/dev/watchque-go/test$ ./challenge_async.sh 10000
d@o:~/dev/watchque-go/test$ ./test_post
resque:queue:CreateQueue
(integer) 10000
resque:queue:DeleteQueue
(integer) 10000
resque:queue:UpdateQueue
(integer) 9985
resque:queue:RenameQueue
(integer) 10000
resque:queue:AllQueue
(integer) 39735
resque:queue:AllQueue2
(integer) 39723




Now a big 100k test (--debug=off):

d@o:~/dev/watchque-go/test$ ./challenge_async.sh 100000
d@o:~/dev/watchque-go/test$ ./test_post
resque:queue:CreateQueue
(integer) 100000
resque:queue:DeleteQueue
(integer) 100000
resque:queue:UpdateQueue
(integer) 99908
resque:queue:RenameQueue
(integer) 100000
resque:queue:AllQueue
(integer) 397865
resque:queue:AllQueue2
(integer) 397805



Two sync tests.

d@o:~/dev/watchque-go/test$ ./challenge_sync.sh 1000
d@o:~/dev/watchque-go/test$ ./test_post
resque:queue:CreateQueue
(integer) 1000
resque:queue:DeleteQueue
(integer) 1000
resque:queue:UpdateQueue
(integer) 1000
resque:queue:RenameQueue
(integer) 1000
resque:queue:AllQueue
(integer) 4000
resque:queue:AllQueue2
(integer) 4000


d@o:~/dev/watchque-go/test$ ./challenge_sync.sh 1000
d@o:~/dev/watchque-go/test$ ./test_post
resque:queue:CreateQueue
(integer) 10000
resque:queue:DeleteQueue
(integer) 10000
resque:queue:UpdateQueue
(integer) 10000
resque:queue:RenameQueue
(integer) 10000
resque:queue:AllQueue
(integer) 39999
resque:queue:AllQueue2
(integer) 39968
