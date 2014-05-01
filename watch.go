package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gosexy/redis"
	"github.com/howeyc/fsnotify"
	"github.com/romanoff/fsmonitor"
	"log"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

type ResqueArgs struct {
	FilePath string `json:"filePath"`
	Event    string `json:"event"`
}

type ResquePacket struct {
	Class string       `json:"class"`
	Args  []ResqueArgs `json:"args"`
}

const (
	DEST_REDIS = iota
	DEST_LOCAL
)

const (
	MASK_CREATE = 0x01
	MASK_UPDATE = 0x02
	MASK_DELETE = 0x04
	MASK_RENAME = 0x08
	MASK_CLOSE_WRITE = 0x10
)

type Watcher struct {
	Dest              string
	Class             string
	Queue             string
	QueuePreFormatted string
	Events            string
	Source            string
	destType          int
	mask              uint32
	Redis             *Redis
	Local             *Local
}

type Redis struct {
	Host string
	Port uint
	red  *redis.Client
}

type Local struct {
	Base string
	Bin  string
}

type Opts struct {
	debug bool
}

var opts Opts

func usage() {
	log.Fatal("usage: ./watchque-go [<redishost:port>|</path/to/bin/dir>] <Class1>:<Queue1>:<Events>:<Directory1,...,DirectoryN> ... <ClassN>:<QueueN>:<Events>:<Directory1, ...,DirectoryN>")
}

func ParseOption(arg string) {
	switch arg {
	case "--debug=on":
		opts.debug = true
	case "--debug=off":
		opts.debug = false
	}
}

func Parse(dest, arg string) ([]*Watcher, error) {
	arr := make([]*Watcher, 0)

	tokens := strings.Split(arg, ":")
	if len(tokens) != 4 {
		return nil, errors.New("Invalid argument")
	}

	class := tokens[0]
	queue := tokens[1]
	events := tokens[2]
	sources := tokens[3]
	destType := DEST_REDIS

	/*
	 * Destination parsing: Redis vs Local scripts
	 */
	dest_tokens := strings.Split(dest, ":")
	if len(dest_tokens) == 0 {
		return nil, errors.New("Invalid destination: Specify a path or redis:port combo")
	}

	dest_0 := dest_tokens[0]
	if dest_0[0] == '/' {
		destType = DEST_LOCAL
	}

	/*
	 * Directory/file (sources) parsing. These sources are what we monitor for events
	 */
	dir_tokens := strings.Split(sources, ",")
	if len(dir_tokens) <= 0 {
		return nil, errors.New("Invalid source directories")
	}

	/*
	 * Event parsing: Events we care about can be c (create), u (update), d (delete) or a (c u d)
	 */
	var mask uint32 = 0
	for _, event := range events {
		switch event {
		case 'c':
			{
				mask |= MASK_CREATE
			}
		case 'C':
			{
				mask |= MASK_CLOSE_WRITE
			}
		case 'u':
			{
				mask |= MASK_UPDATE
			}
		case 'd':
			{
				mask |= MASK_DELETE
			}
		case 'r':
			{
				mask |= MASK_RENAME
			}
		case 'a':
			{
				mask = MASK_CREATE | MASK_UPDATE | MASK_DELETE | MASK_RENAME | MASK_CLOSE_WRITE
			}
		default:
			{
				log.Fatal("Parse: Invalid event", event)
			}
		}
	}

	for _, source := range dir_tokens {
		a := new(Watcher)
		a.Dest = dest
		a.Class = class
		a.Queue = queue
		a.QueuePreFormatted = fmt.Sprintf("resque:queue:%s", a.Queue)
		a.Events = events
		a.Source = source
		a.destType = destType
		a.mask = mask
		switch destType {
		case DEST_REDIS:
			{
				a.Redis = new(Redis)
				a.Redis.Host = dest_0
				a.Redis.Port = 6379
				if len(dest_tokens) > 1 {
					ui, _ := strconv.Atoi(dest_tokens[1])
					a.Redis.Port = uint(ui)
				}
			}
		case DEST_LOCAL:
			{
				a.Local = new(Local)
				a.Local.Base = dest
				a.Local.Bin = fmt.Sprintf("%s/%s/%s", dest, a.Class, a.Queue)
			}
		default:
			{
				return nil, errors.New("Invalid destination: Specify a path or redis:port combo")
			}
		}

		arr = append(arr, a)
	}

	return arr, nil
}

/*
 * Checks to see if we have this event in our mask. If so, return true & the string translation of the event type
 */
func isDesiredEvent(mask uint32, ev *fsnotify.FileEvent) (bool, string) {

	if (mask&MASK_CREATE > 0) && ev.IsCreate() {
		return true, "CREATE"
	} else if (mask&MASK_CLOSE_WRITE > 0) && ev.IsCloseWrite() {
		return true, "CLOSE_WRITE"
	} else if (mask&MASK_DELETE > 0) && ev.IsDelete() {
		return true, "DELETE"
	} else if (mask&MASK_UPDATE > 0) && ev.IsModify() {
		return true, "UPDATE"
	} else if (mask&MASK_RENAME > 0) && ev.IsRename() {
		return true, "RENAME"
	}

	return false, ""
}

func (this *Watcher) Dump() {
	Debug("Dest=%s, Class=%s, Queue=%s, Events=%s, Source=%s, destType=%i, mask=%i\n",
		this.Dest,
		this.Class,
		this.Queue,
		this.Events,
		this.Source,
		this.destType,
		this.mask)
}

func Launch(watchers []*Watcher) {

	ch := make(chan *fsnotify.FileEvent, 255)

	go watchers[0].Transponder(ch)

	for _, watcher := range watchers {
		go watcher.Launch(ch)
	}
}

func (this *Watcher) Transponder(ch chan *fsnotify.FileEvent) {
	switch this.destType {
	case DEST_REDIS:
		this.TransponderRedis(ch)
	case DEST_LOCAL:
		this.TransponderLocal(ch)
	default:
		{
			log.Fatal("Transponder: Unknown transponder type")
		}
	}
}

func (this *Watcher) TransponderRedis(ch chan *fsnotify.FileEvent) {
	DebugLn("TransponderRedis: Entered")

	this.Redis.red = redis.New()

	err := this.Redis.red.Connect(this.Redis.Host, this.Redis.Port)
	if err != nil {
		log.Fatal("TransponderRedis: Connect failed: %s\n", err.Error())
	}

	/* Quick and dirty sadd.. want to add this to a per minute/per 30s interval. No need to sadd every time we rpush */
	this.Redis.red.SAdd("resque:queues", this.Queue)

	for ev := range ch {
		Debug("TransponderRedis: Received event: %v\n", ev)

		boo, event := isDesiredEvent(this.mask, ev)
		if boo == false {
			Debug("TransponderRedis: Received undesired event: %v %s\n", ev, event)
			continue
		}

		rpkt := ResquePacket{}
		rpkt.Class = this.Class
		rpkt.Args = make([]ResqueArgs, 1)
		rpkt.Args[0].FilePath = ev.Name
		rpkt.Args[0].Event = event

		Debug("TransponderRedis: Received desired event: %v %s\n", ev, event)

		js, err := json.Marshal(&rpkt)
		if err != nil {
			log.Println(err)
		}

		_, _ = this.Redis.red.RPush(this.QueuePreFormatted, string(js))
	}
}

func (this *Watcher) TransponderLocal(ch chan *fsnotify.FileEvent) {
	Debug("TransponderLocal: Entered")
	for ev := range ch {
		Debug("TransponderLocal: Received event: %v\n", ev)

		boo, event := isDesiredEvent(this.mask, ev)
		if boo == false {
			Debug("TransponderLocal: Received undesired event: %v %s\n", ev, event)
			continue
		}

		cmd := exec.Command(this.Local.Bin, event, this.Class, this.Queue, ev.Name)
		err := cmd.Start()
		if err != nil {
			log.Printf("TransponderLocal: Error executing %s\n", this.Local.Bin)
			continue
		}

		Debug("TransponderLocal: Executed %s\n", this.Local.Bin)
	}
}

func (this *Watcher) Launch(ch chan *fsnotify.FileEvent) {
	this.Dump()
	mon, err := fsmonitor.NewWatcher()
	if err != nil {
		log.Fatal("Launch:fsmonitor.NewWatcher:Error:", err)
	}

	err = mon.Watch(this.Source)
	if err != nil {
		log.Fatal("Launch:mon.Watch:Error:", err)
	}

	for {
		ev := <-mon.Event
		Debug("Launch: Received event: %v\n", ev)
		ch <- ev
	}
}

func main() {

	if len(os.Args) < 3 {
		usage()
	}

	dest := os.Args[1]

	arr := make([][]*Watcher, 0)

	for k, arg := range os.Args {
		if k <= 1 {
			continue
		}
		/*
		 * Crude option processing
		 */
		if strings.HasPrefix(arg, "--") == true {
			ParseOption(arg)
			continue
		}

		watchers, err := Parse(dest, arg)
		if err != nil {
			log.Fatal(err)
		}
		arr = append(arr, watchers)
	}

	for _, watchers := range arr {
		go Launch(watchers)
	}

	select {}
}
