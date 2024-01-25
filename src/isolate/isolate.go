package isolate

import (
	"bytes"
	"log"
	"os"
	"strconv"
	"sync"

	"github.com/joshjms/pocket-watch/src/models"
)

type QueueManager struct {
	Mutex  sync.Mutex
	BoxIds []bool
	Sem    chan int

	MaxInstances int
}

var qm *QueueManager

func InitQueueManager() {
	qm = &QueueManager{}

	qm.Mutex = sync.Mutex{}
	qm.BoxIds = make([]bool, 10)
	qm.Sem = make(chan int, 10)
	qm.MaxInstances = 10
}

func GetQueueManager() *QueueManager {
	return qm
}

type IsolateInstance struct {
	Box      *Box
	Request  *IsolateRequest
	Response *IsolateResponse
}

type IsolateRequest struct {
	Code     string
	Language string
	Stdin    []string
	Options  *Options
}

type IsolateResponse struct {
	Verdict []string
	Stdout  []string
	Stderr  []string
	Time    []int
	Memory  []int
}

type Options struct {
	MemoryLimit int // in MB
	TimeLimit   int // in seconds
	Processes   int // number of processes allowed by the program; default 1
}

func GetBoxId() int {
	qm.Mutex.Lock()
	defer qm.Mutex.Unlock()

	for i, v := range qm.BoxIds {
		if !v {
			qm.BoxIds[i] = true
			return i
		}
	}

	return -1
}

func ReleaseBoxId(boxId int) {
	qm.Mutex.Lock()
	defer qm.Mutex.Unlock()

	qm.BoxIds[boxId] = false
}

func CreateInstance(req models.Request) (*IsolateInstance, error) {
	var instance *IsolateInstance
	qm.Sem <- 1

	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		box, err := CreateBox()
		if err != nil {
			log.Print(err)
			return
		}

		instance = &IsolateInstance{
			Box: box,
			Request: &IsolateRequest{
				Code:     req.Code,
				Language: req.Language,
				Stdin:    req.Stdin,
				Options: &Options{
					MemoryLimit: 256,
					TimeLimit:   1,
					Processes:   1,
				},
			},
			Response: &IsolateResponse{},
		}

		err = instance.Run()
		if err != nil {
			log.Print(err)
		}

		<-qm.Sem
	}()

	wg.Wait()

	return instance, nil
}

func (instance *IsolateInstance) Run() error {
	metaFilesDir := "meta/"
	if fdir := os.Getenv("PW_META_FILES_DIR"); fdir != "" {
		metaFilesDir = fdir
	}

	b := instance.Box
	if err := b.Init(); err != nil {
		log.Print("failed to init box : ", err)
		instance.Response.Verdict = append(instance.Response.Verdict, "Internal Error")
		return err
	}

	file, err := os.Create(b.Dir + "src.cpp")
	if err != nil {
		log.Print("failed to create src.cpp file : ", err)
		instance.Response.Verdict = append(instance.Response.Verdict, "Internal Error")
		return err
	}
	defer file.Close()

	file.WriteString(instance.Request.Code)

	if err := b.Compile(); err != nil {
		log.Print("failed to compile : ", err)
		instance.Response.Verdict = append(instance.Response.Verdict, "CE")
		return err
	}

	if err := os.Mkdir(metaFilesDir+b.Id, 0777); err != nil {
		log.Print("failed to create meta dir : ", err)
		instance.Response.Verdict = append(instance.Response.Verdict, "Internal Error")
		return err
	}

	for i, stdin := range instance.Request.Stdin {
		stdinDir := strconv.Itoa(i) + ".in"
		stdinFile, err := os.Create(b.Dir + stdinDir)
		if err != nil {
			log.Print("failed to create .in file : ", err)
			instance.HandleError("Internal Error")
			continue
		}
		defer stdinFile.Close()

		stdinFile.WriteString(stdin)

		stdoutDir := strconv.Itoa(i) + ".out"
		stdoutFile, err := os.Create(b.Dir + stdoutDir)
		if err != nil {
			log.Print("failed to create .out file : ", err)
			instance.HandleError("Internal Error")
			continue
		}
		defer stdoutFile.Close()

		stderrDir := strconv.Itoa(i) + ".err"
		stderrFile, err := os.Create(b.Dir + stderrDir)
		if err != nil {
			log.Print("failed to create .err file : ", err)
			instance.HandleError("Internal Error")
			continue
		}
		defer stderrFile.Close()

		metaFileDir := metaFilesDir + b.Id + "/" + strconv.Itoa(i) + ".txt"
		metaFile, err := os.Create(metaFileDir)
		if err != nil {
			log.Print("failed to create meta file : ", err)
			instance.HandleError("Internal Error")
			continue
		}
		defer metaFile.Close()

		logs, err := b.Run("./exec", instance.Request.Options, stdinDir, stdoutDir, stderrDir, metaFileDir)
		if err != nil {
			log.Print("failed to run : ", err)
			if string(logs) == "Time limit exceeded\n" {
				instance.HandleError("TLE")
			} else if string(logs) == "Memory limit exceeded\n" {
				instance.HandleError("MLE")
			} else {
				instance.HandleError("RTE")
			}
			log.Print(string(logs))
			continue
		}

		instance.Response.Verdict = append(instance.Response.Verdict, "OK")

		if err == nil {
			buf := bytes.NewBuffer(nil)
			buf.ReadFrom(stdoutFile)
			instance.Response.Stdout = append(instance.Response.Stdout, buf.String())
		} else {
			instance.Response.Stdout = append(instance.Response.Stdout, "")
		}

		if err != nil {
			buf := bytes.NewBuffer(nil)
			buf.ReadFrom(stderrFile)
			instance.Response.Stderr = append(instance.Response.Stderr, buf.String())
		} else {
			instance.Response.Stderr = append(instance.Response.Stderr, "")
		}

		buf := bytes.NewBuffer(nil)
		buf.ReadFrom(metaFile)
		log.Print(buf.String())

		mf := GetDetails(buf.String())
		runTime := mf.Get("time")
		runMemory := mf.Get("cg-mem")

		runTimeFloat, err := strconv.ParseFloat(runTime, 32)
		if err != nil {
			log.Print("failed to convert runTime to int : ", err)
			instance.Response.Verdict = append(instance.Response.Verdict, "Internal Error")
			return err
		}
		runTimeInt := int(runTimeFloat * 1000)

		runMemoryInt, err := strconv.Atoi(runMemory)
		if err != nil {
			log.Print("failed to convert runMemory to int : ", err)
			instance.Response.Verdict = append(instance.Response.Verdict, "Internal Error")
			return err
		}

		instance.Response.Time = append(instance.Response.Time, runTimeInt)
		instance.Response.Memory = append(instance.Response.Memory, runMemoryInt)
	}

	if err := b.Exit(); err != nil {
		log.Print("failed to cleanup box : ", err)
		instance.Response.Verdict = append(instance.Response.Verdict, "Internal Error")
		return err
	}

	if err := os.RemoveAll(metaFilesDir + b.Id); err != nil {
		return err
	}

	return nil
}

func (instance *IsolateInstance) HandleError(verdict string) {
	instance.Response.Verdict = append(instance.Response.Verdict, verdict)
	instance.Response.Stdout = append(instance.Response.Stdout, "")
	instance.Response.Stderr = append(instance.Response.Stderr, "")
	instance.Response.Time = append(instance.Response.Time, -1)
	instance.Response.Memory = append(instance.Response.Memory, -1)
}
