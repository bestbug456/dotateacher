package main

import (
	"context"
	"fmt"
	"os"
	"strconv"

	wq "github.com/bestbug456/dotateacher/workingqueue"

	"github.com/aws/aws-lambda-go/lambda"
	"gopkg.in/mgo.v2"
)

var compressed map[int]int

var (
	JOBNUMBER = 100
)

func init() {
	compressed = make(map[int]int)
	for i := 0; i < len(uncompressed); i++ {
		compressed[uncompressed[i]] = int(i)
	}
}

func main() {
	trainAndValidateNN()
	lambda.Start(HandleRequest)
}

func HandleRequest(ctx context.Context, data interface{}) (string, error) {
	return trainAndValidateNN()
}

func trainAndValidateNN() (string, error) {

	address := os.Getenv("address")
	username := os.Getenv("username")
	password := os.Getenv("password")
	option := os.Getenv("option")
	ssl := os.Getenv("ssl")
	jobsNumber := os.Getenv("nrjobs")
	if jobsNumber != "" {
		var err error
		JOBNUMBER, err = strconv.Atoi(jobsNumber)
		if err != nil {
			return "", fmt.Errorf("Error while converting jobsNumber: %s", err.Error())
		}
	}

	var s *mgo.Session
	var err error
	if ssl == "false" {
		s, err = mgo.Dial(fmt.Sprintf("mongodb://%s:%s@%s/%s", username, password, address, option))
		if err != nil {
			return "", fmt.Errorf("Error while connecting without ssl: %s", err.Error())
		}
	} else {
		s, err = DialUsingSSL(address, option, username, password)
		if err != nil {
			return "", fmt.Errorf("Error while connecting via ssl: %s", err.Error())
		}
	}
	defer s.Close()

	err = s.DB("opendota-infos").C("matchs").EnsureIndex(mgo.Index{
		Unique: true,
		Key:    []string{"id"},
	})

	traindata, testdata, err := getDatasetAndTrainSet(s)
	if err != nil {
		return "", fmt.Errorf("Error while get trainset and testset: %s", err.Error())
	}

	// Get old weights if exist or create new one
	// if is the first run.
	NN, err := getActualNewNeuralNetwork(s)
	if err != nil && err != mgo.ErrNotFound {
		return "", fmt.Errorf("Error while get actual weights: %s", err.Error())
	}
	var created bool
	if err != nil {
		created = true
		NN, err = trainNewNeuralNetwork(traindata, 25)
	}

	// Check how many correct prediction we have
	storedNNResult := checkWeightsQuality(NN, testdata)

	accuracy, sensitivity := getStastFromMatrixQA(storedNNResult)

	// If we didn't have at least 70% of accuracy
	// try to find a better weights.
	if accuracy < 0.8 || sensitivity < 0.8 {

		wq := wq.NewWorkingQueue(50, 200, nil)
		wq.Run()

		responseChan := make(chan NNmessage, 0)
		for i := 0; i < JOBNUMBER; i++ {
			wq.SendJob(CreateNewNeuralNetworkAndValidate, &JobArgs{
				TrainData:    traindata,
				TestData:     testdata,
				NeuronNumber: i,
				result:       responseChan,
			})
		}

		var bestaccuracy float64
		var bestsensitivity float64
		var bestResult NNmessage
		for i := 0; i < JOBNUMBER; i++ {
			result := <-responseChan
			// Ignore invalid matrix
			if result.MatrixQA == nil ||
				result.MatrixQA.ConfusionMatrix == nil ||
				len(result.MatrixQA.ConfusionMatrix) != 2 ||
				len(result.MatrixQA.ConfusionMatrix[0]) != 2 ||
				len(result.MatrixQA.ConfusionMatrix[1]) != 2 {
				continue
			}
			resultaccuracy, resultsensitivity := getStastFromMatrixQA(result.MatrixQA)
			if resultaccuracy >= bestaccuracy && resultsensitivity >= bestsensitivity {
				bestaccuracy = resultaccuracy
				bestsensitivity = resultsensitivity
				bestResult = result
			}
		}

		// If the new weights have at least 70% of accuracy
		// OR is better then the actual save it to the database
		if bestResult.MatrixQA == nil {
			return "", fmt.Errorf("MatrixQA have zero len (%+v)", bestResult.MatrixQA)
		}

		if bestaccuracy >= accuracy && bestsensitivity >= sensitivity {
			err = storeNewNeuralNetworkAndQAResults(bestResult, s)
			if err != nil {
				return "", fmt.Errorf("Error while storing actual weights: %s", err.Error())
			}
		}
		// We just created the new NN, store it to the database
		if created {
			if float64(bestResult.MatrixQA.CorrectPrediction)/float64(len(testdata)) > float64(storedNNResult.CorrectPrediction)/float64(len(testdata)) {
				err = storeNewNeuralNetworkAndQAResults(bestResult, s)
				if err != nil {
					return "", fmt.Errorf("Error while storing actual weights: %s", err.Error())
				}
			} else {
				err = storeNewNeuralNetworkAndQAResults(NNmessage{
					NN:       NN,
					MatrixQA: storedNNResult,
				}, s)
				if err != nil {
					return "", fmt.Errorf("Error while storing actual weights: %s", err.Error())
				}
			}
		}
	}
	return "", nil
}
