package main

import (
	"fmt"
	"log"
	"os"

	"github.com/unixpickle/neuralstruct"
	"github.com/unixpickle/seqtasks"
	"github.com/unixpickle/weakai/neuralnet"
)

type Task struct {
	Name  string
	Task  seqtasks.Task
	Model seqtasks.Model

	MaxEpochs    int
	MaxScore     float64
	TrainingSize int
	TestingBatch int
	TestingCount int
}

var StructNames = []string{"stack"}
var Structs = map[string]neuralstruct.RStruct{
	"stack": &neuralstruct.Stack{VectorSize: 40},
}
var Activations = map[string]neuralnet.Layer{
	"stack": &StackActivation{VecSize: 40},
}

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintln(os.Stderr, "Usage:", os.Args[0], "<struct>")
		fmt.Fprintln(os.Stderr, "Available structs:")
		for _, s := range StructNames {
			fmt.Fprintln(os.Stderr, " -", s)
		}
		os.Exit(1)
	}

	structure, ok := Structs[os.Args[1]]
	if !ok {
		fmt.Fprintln(os.Stderr, "Unknown struct:", os.Args[1])
		os.Exit(1)
	}
	activation := Activations[os.Args[1]]

	tasks := []Task{
		{
			Name: "XOR last",
			Task: &seqtasks.XORLastTask{SeqLen: 50},
			Model: &SeqFuncModel{
				SeqFunc: AddActivation(NewSeqFunc(structure, 1, 40, 40, 1),
					activation),
				Cost:          &neuralnet.SigmoidCECost{},
				OutActivation: &neuralnet.Sigmoid{},
			},
			MaxEpochs:    50,
			MaxScore:     1,
			TrainingSize: 100,
			TestingBatch: 10,
			TestingCount: 10,
		},
		{
			Name: "Repeat",
			Task: &seqtasks.RepeatTask{
				MinString: 2,
				MaxString: 5,
				MinGap:    0,
				MaxGap:    6,
			},
			Model: &SeqFuncModel{
				SeqFunc: AddActivation(NewSeqFunc(structure, 3, 100, 100, 1),
					activation),
				Cost:          &neuralnet.SigmoidCECost{},
				OutActivation: &neuralnet.Sigmoid{},
			},
			MaxEpochs:    100,
			MaxScore:     1,
			TrainingSize: 300,
			TestingBatch: 10,
			TestingCount: 30,
		},
		{
			Name: "Match Open",
			Task: &seqtasks.MatchOpenTask{
				MinLen:  1,
				MaxLen:  15,
				MaxOpen: 6,
			},
			Model: &SeqFuncModel{
				SeqFunc: AddActivation(NewSeqFunc(structure, 3, 40, 40, 1),
					activation),
				Cost:          &neuralnet.SigmoidCECost{},
				OutActivation: &neuralnet.Sigmoid{},
			},
			MaxEpochs:    50,
			MaxScore:     1,
			TrainingSize: 100,
			TestingBatch: 10,
			TestingCount: 30,
		},
		{
			Name: "Match Multi",
			Task: &seqtasks.MatchMultiTask{
				TypeCount: 4,
				MinLen:    1,
				MaxLen:    8,
				CloseProb: 0.3,
			},
			Model: &SeqFuncModel{
				SeqFunc: AddActivation(NewDeepSeqFunc(structure, 4*2+1, 100, 2, 100, 4+1),
					activation),
				Cost:          &neuralnet.SigmoidCECost{},
				OutActivation: &neuralnet.Sigmoid{},
			},
			MaxEpochs:    1000,
			MaxScore:     1,
			TrainingSize: 3000,
			TestingBatch: 20,
			TestingCount: 100,
		},
	}
	for _, task := range tasks {
		log.Println("Running task", task.Name, "...")
		samples := task.Task.NewSamples(task.TrainingSize)
		for i := 0; i < task.MaxEpochs; i++ {
			task.Model.Train(samples)
			score := task.Task.Score(task.Model, task.TestingBatch, task.TestingCount)
			log.Printf("epoch %d: score=%f", i, score)
			if score >= task.MaxScore {
				break
			}
		}
	}
}