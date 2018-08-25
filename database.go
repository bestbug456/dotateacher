package main

import (
	"crypto/tls"
	"fmt"
	"net"
	"strings"

	rprop "github.com/bestbug456/gorpropplus"

	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

func DialUsingSSL(addresses string, dboption string, username string, password string) (*mgo.Session, error) {
	listaddresses := make([]string, 0)
	for _, str := range strings.Split(addresses, ",") {
		if str != "" {
			listaddresses = append(listaddresses, str)
		}
	}
	dboptions := strings.Split(dboption, "=")
	if len(dboption) < 2 {
		return nil, fmt.Errorf("can not found authSource keyword in order to permit SSL connection, aborting")
	}
	tlsConfig := &tls.Config{}
	dialInfo := &mgo.DialInfo{
		Addrs:    listaddresses,
		Database: dboptions[1],
		Username: username,
		Password: password,
	}

	dialInfo.DialServer = func(addr *mgo.ServerAddr) (net.Conn, error) {
		conn, err := tls.Dial("tcp", addr.String(), tlsConfig)
		return conn, err
	}
	session, err := mgo.DialWithInfo(dialInfo)
	if err != nil {
		return nil, err
	}
	session.EnsureSafe(&mgo.Safe{
		W:     1,
		FSync: false,
	})
	return session, nil
}

func getDatasetAndTrainSet(s *mgo.Session) ([]MatchInfos, []MatchInfos, error) {

	var results0 []MatchInfos
	err := s.DB("opendota-infos").C("matchs").Find(bson.M{"win": 0}).All(&results0)
	if err != nil {
		return nil, nil, err
	}

	var results1 []MatchInfos
	err = s.DB("opendota-infos").C("matchs").Find(bson.M{"win": 1}).All(&results1)
	if err != nil {
		return nil, nil, err
	}

	traindata := results0[0 : len(results0)/2]
	traindata1 := results1[0 : len(results1)/2]

	testdata := results0[len(traindata) : len(results0)-1]
	testdata1 := results1[len(traindata) : len(results1)-1]

	traindata = append(traindata, traindata1...)
	testdata = append(testdata, testdata1...)
	return traindata, testdata, nil
}

func storeNewNeuralNetworkAndQAResults(msg NNmessage, s *mgo.Session) error {
	count, err := s.DB("neuralnetwork").C("weights").Count()
	if err != nil {
		return err
	}
	if count != 0 {
		err = s.DB("neuralnetwork").C("weights").DropCollection()
		if err != nil {
			return err
		}
	}

	err = s.DB("neuralnetwork").C("weights").Insert(msg.NN)
	if err != nil {
		return err
	}

	count, err = s.DB("neuralnetwork").C("score").Count()
	if err != nil {
		return err
	}

	if count != 0 {
		err = s.DB("neuralnetwork").C("score").DropCollection()
		if err != nil {
			return err
		}
	}

	err = s.DB("neuralnetwork").C("score").Insert(NNStats{
		MatrixQA: msg.MatrixQA,
	})
	if err != nil {
		return err
	}

	return nil
}

func getActualNewNeuralNetwork(s *mgo.Session) (*rprop.NeuralNetwork, error) {
	var NN rprop.NeuralNetwork
	err := s.DB("neuralnetwork").C("weights").Find(nil).One(&NN)
	if err != nil {
		return nil, err
	}

	NN.ActivationFunction = rprop.Logistic
	NN.DerivateActivation = rprop.DerivateLogistic
	NN.ErrorFunction = rprop.SSE
	NN.DerivateError = rprop.DerivateSSE

	return &NN, nil
}
