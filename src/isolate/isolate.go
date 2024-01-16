package isolate

import (
	"bytes"
	"log"
	"os"
	"os/exec"
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
	BoxId string
	Dir   string

	Request  *IsolateRequest
	Response *IsolateResponse
}

type IsolateRequest struct {
	Code     string
	Language string
	Stdin    []string
}

type IsolateResponse struct {
	Verdict []string
	Stdout  []string
	Stderr  []string

	Time   []string
	Memory []string
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
		boxId := GetBoxId()

		if boxId == -1 {
			log.Print("No box available")
			panic("No box available")
		}

		instance = &IsolateInstance{
			BoxId: strconv.Itoa(boxId),
			Dir:   "/var/local/lib/isolate/" + strconv.Itoa(boxId) + "/box/",
			Request: &IsolateRequest{
				Code:     req.Code,
				Language: req.Language,
				Stdin:    req.Stdin,
			},
			Response: &IsolateResponse{},
		}

		err := instance.Run()
		if err != nil {
			log.Print(err)
		}

		ReleaseBoxId(boxId)

		<-qm.Sem
	}()

	wg.Wait()

	return instance, nil
}

func (instance *IsolateInstance) Run() error {
	initCleanupCmd := exec.Command("isolate", "--cleanup", "--cg", "--box-id="+instance.BoxId)
	initCleanupCmd.Run()

	initCmd := exec.Command("isolate", "--init", "--cg", "--box-id="+instance.BoxId)
	logs, err := initCmd.CombinedOutput()
	if err != nil {
		log.Print("failed to init : ", err)
		log.Print(string(logs))
		instance.Response.Verdict = append(instance.Response.Verdict, "Internal Error")
		return err
	}

	file, err := os.Create(instance.Dir + "src.cpp")
	if err != nil {
		log.Print("failed to create : ", err)
		instance.Response.Verdict = append(instance.Response.Verdict, "Internal Error")
		return err
	}
	defer file.Close()

	file.WriteString(instance.Request.Code)

	compileCmd := exec.Command("g++", "-o", instance.Dir+"exec", instance.Dir+"src.cpp")
	logs, err = compileCmd.CombinedOutput()
	if err != nil {
		log.Print("failed to compile : ", err)
		log.Print(string(logs))

		instance.Response.Verdict = append(instance.Response.Verdict, "CE")

		return err
	}

	if err := os.Mkdir("meta/"+instance.BoxId, 0777); err != nil {
		log.Print("failed to create meta dir : ", err)

		instance.Response.Verdict = append(instance.Response.Verdict, "Internal Error")
		return err
	}

	for i, stdin := range instance.Request.Stdin {
		stdinFile, err := os.Create(instance.Dir + strconv.Itoa(i) + ".in")
		if err != nil {
			log.Print("failed to create .in file : ", err)
			instance.HandleError()
			continue
		}
		defer stdinFile.Close()

		stdinFile.WriteString(stdin)

		stdoutFile, err := os.Create(instance.Dir + strconv.Itoa(i) + ".out")
		if err != nil {
			log.Print("failed to create .out file : ", err)
			instance.HandleError()
			continue
		}
		defer stdoutFile.Close()

		stderrFile, err := os.Create(instance.Dir + strconv.Itoa(i) + ".err")
		if err != nil {
			log.Print("failed to create .err file : ", err)
			instance.HandleError()
			continue
		}
		defer stderrFile.Close()

		metaFile, err := os.Create("meta/" + instance.BoxId + "/" + strconv.Itoa(i))
		if err != nil {
			log.Print("failed to create meta file : ", err)
			instance.HandleError()
			continue
		}
		defer metaFile.Close()

		runCmd := exec.Command("isolate",
			"--run",
			"--box-id="+instance.BoxId,
			"./exec",
			"--cg",
			"--stdin="+strconv.Itoa(i)+".in",
			"--stdout="+strconv.Itoa(i)+".out",
			"--stderr="+strconv.Itoa(i)+".err",
			"--meta="+"meta/"+instance.BoxId+"/"+strconv.Itoa(i),
			"--mem=262144", // 256 MB
			"--time=1",
			"--wall-time=2",
			"--extra-time=0.5",
			"--fsize=262144",
			"--processes=1",
		)

		logs, err = runCmd.CombinedOutput()
		if err != nil {
			log.Print("failed to run : ", err)
			instance.HandleError()
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

		instance.Response.Time = append(instance.Response.Time, runTime)
		instance.Response.Memory = append(instance.Response.Memory, runMemory)

	}

	if err := instance.Cleanup(); err != nil {
		log.Print("failed to cleanup : ", err)
		return err
	}

	return nil
}

func (instance *IsolateInstance) HandleError() {
	instance.Response.Verdict = append(instance.Response.Verdict, "RE")
	instance.Response.Stdout = append(instance.Response.Stdout, "")
	instance.Response.Stderr = append(instance.Response.Stderr, "")
	instance.Response.Time = append(instance.Response.Time, "")
	instance.Response.Memory = append(instance.Response.Memory, "")
}

func (instance *IsolateInstance) Cleanup() error {
	cleanupCmd := exec.Command("isolate", "--cleanup", "--cg", "--box-id="+instance.BoxId)
	if err := cleanupCmd.Run(); err != nil {
		return err
	}

	if err := os.RemoveAll("meta/" + instance.BoxId); err != nil {
		return err
	}

	return nil
}
