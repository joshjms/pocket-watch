package isolate

import (
	"fmt"
	"log"
	"os/exec"
	"strconv"
)

type Box struct {
	Id  string
	Dir string
}

func CreateBox() (*Box, error) {
	id := GetBoxId()
	if id == -1 {
		log.Print("No box available")
		return nil, fmt.Errorf("No box available")
	}
	b := &Box{
		Id:  strconv.Itoa(id),
		Dir: "/var/local/lib/isolate/" + strconv.Itoa(id) + "/box/",
	}
	return b, nil
}

func (b *Box) Init() error {
	log.Print("Init box : ", b.Id)

	if err := b.Cleanup(); err != nil {
		log.Print("failed to cleanup : ", err)
		return err
	}
	err := exec.Command("isolate", "--init", "--cg", "--box-id="+b.Id).Run()
	return err
}

func (b *Box) Cleanup() error {
	err := exec.Command("isolate", "--cleanup", "--cg", "--box-id="+b.Id).Run()
	return err
}

func (b *Box) Compile() error {
	compileCmd := exec.Command("g++", "-o", b.Dir+"exec", b.Dir+"src.cpp")
	logs, err := compileCmd.CombinedOutput()
	if err != nil {
		log.Print("failed to compile : ", err)
		log.Print(string(logs))
		return err
	}
	return nil
}

func (b *Box) Run(command string, options *Options, stdin string, stdout string, stderr string, meta string) ([]byte, error) {
	memLimit := strconv.Itoa(options.MemoryLimit * 1024)
	timeLimit := strconv.Itoa(options.TimeLimit)
	processes := strconv.Itoa(options.Processes)

	return exec.Command("isolate",
		"--run",
		"--box-id="+b.Id,
		"./exec",
		"--cg",
		"--stdin="+stdin,
		"--stdout="+stdout,
		"--stderr="+stderr,
		"--meta="+meta,
		"--mem="+memLimit,
		"--time="+timeLimit,
		"--wall-time=2",
		"--extra-time=0.5",
		"--fsize=262144",
		"--processes="+processes,
	).CombinedOutput()
}

func (b *Box) Exit() error {
	boxId, err := strconv.Atoi(b.Id)
	if err != nil {
		return err
	}
	ReleaseBoxId(boxId)
	return nil
}
