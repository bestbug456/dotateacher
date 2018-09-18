package main

import (
	"context"
	"fmt"
	"os"

	wq "github.com/bestbug456/dotateacher/workingqueue"

	"github.com/aws/aws-lambda-go/lambda"
	"gopkg.in/mgo.v2"
)

var compressed map[int]int

const (
	JOBNUMBER = 100
)

func init() {
	compressed = make(map[int]int)
	for i := 0; i < len(uncompressed); i++ {
		compressed[uncompressed[i]] = int(i)
	}
}

func main() {
	lambda.Start(HandleRequest)
}

func HandleRequest(ctx context.Context, data interface{}) (string, error) {

	address := os.Getenv("address")
	username := os.Getenv("username")
	password := os.Getenv("password")
	option := os.Getenv("option")
	ssl := os.Getenv("ssl")

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

	// If we didn't have at least 70% of accuracy
	// try to find a better weights.
	if float64(storedNNResult.CorrectPrediction)/float64(len(testdata)) < 0.7 {

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

		var max int
		var bestResult NNmessage
		for i := 0; i < JOBNUMBER; i++ {
			result := <-responseChan
			if result.MatrixQA.CorrectPrediction > max || max == 0 {
				max = result.MatrixQA.CorrectPrediction
				bestResult = result
			}
		}

		// If the new weights have at least 70% of accuracy
		// OR is better then the actual save it to the database
		if bestResult.MatrixQA.CorrectPrediction == 0 {
			return "", fmt.Errorf("MatrixQA have zero len (%+v)", bestResult)
		}
		if float64(bestResult.MatrixQA.CorrectPrediction)/float64(len(testdata)) > float64(storedNNResult.CorrectPrediction)/float64(len(testdata)) {
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
