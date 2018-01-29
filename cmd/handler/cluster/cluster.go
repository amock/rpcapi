package cluster

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"

	"github.com/mitchellh/mapstructure"

	"github.com/amock/rpcapi/cmd/handler"
)

var nextClusterID uint64
var actions map[string]func(*handler.Request) interface{}
var clusterMap map[string]*cluster

const (
	statusOk    = iota
	paramsError = iota
	notFound    = iota
)

func init() {
	actions = make(map[string]func(*handler.Request) interface{})
	clusterMap = make(map[string]*cluster)
	handler.Add("cluster", HandlerFunc)
	actions["create"] = create
	actions["read"] = read
	actions["update"] = update
	actions["delete"] = deleteCluster
	actions["list"] = listClusters
}

type cluster struct {
	Name       string
	ID         string
	NodeGroups []nodeDef
}

type nodeDef struct {
	Type  string
	Count uint64
}

type createRequest struct {
	Name       string
	NodeGroups []nodeDef
}

type createResponse struct {
	Status        uint32
	StatusMessage string
	ClusterID     string
}

type readRequest struct {
	ID string
}

type readResponse struct {
	Cluster       cluster
	Status        uint32
	StatusMessage string
}

type updateRequest struct {
	ID         string
	NodeGroups []nodeDef
}

type updateResponse struct {
	Status        uint32
	StatusMessage string
}

type deleteRequest struct {
	ID string
}

type deleteResponse struct {
	Status        uint32
	StatusMessage string
}

type listRequest struct {
	// Empty
}

type listResponse struct {
	Clusters      []cluster
	Status        uint32
	StatusMessage string
}

func create(r *handler.Request) interface{} {
	params := new(createRequest)
	resp := createResponse{}
	cluster := cluster{}
	err := mapstructure.Decode(r.Params, params)
	if err != nil {
		log.Printf("Error in params %s: %s\n", r.Params, err.Error())
		resp.Status = paramsError
		resp.StatusMessage = err.Error()
	} else {
		cluster.ID = strconv.FormatUint(nextClusterID, 10)
		cluster.Name = params.Name
		cluster.NodeGroups = params.NodeGroups
		clusterMap[cluster.ID] = &cluster
		nextClusterID++
		resp.ClusterID = cluster.ID
		resp.Status = statusOk
		resp.StatusMessage = fmt.Sprintf("Successfully created with params %v", params)
	}
	return resp
}

func read(r *handler.Request) interface{} {
	params := new(readRequest)
	resp := readResponse{}
	err := mapstructure.Decode(r.Params, params)
	if err != nil {
		log.Printf("Error in params %s: %s\n", r.Params, err.Error())
		resp.Status = paramsError
		resp.StatusMessage = err.Error()
	} else {
		cluster, ok := clusterMap[params.ID]
		if !ok {
			resp.Status = notFound
			resp.StatusMessage = fmt.Sprintf("Cluster %s not found", params.ID)
		} else {
			resp.Status = statusOk
			resp.StatusMessage = fmt.Sprintf("Found cluster with ID %s", params.ID)
			resp.Cluster = *cluster
		}
	}
	return resp
}

func update(r *handler.Request) interface{} {
	params := new(updateRequest)
	resp := updateResponse{}
	err := mapstructure.Decode(r.Params, params)
	if err != nil {
		log.Printf("Error in params %s: %s\n", r.Params, err.Error())
		resp.Status = paramsError
		resp.StatusMessage = err.Error()
	} else {
		cluster, ok := clusterMap[params.ID]
		if !ok {
			resp.Status = notFound
			resp.StatusMessage = fmt.Sprintf("Cluster %s not found", params.ID)
		} else {
			cluster.NodeGroups = params.NodeGroups
			resp.Status = statusOk
			resp.StatusMessage = fmt.Sprintf("Successfully updated with params %v", params)
		}
	}
	return resp
}

func deleteCluster(r *handler.Request) interface{} {
	params := new(deleteRequest)
	resp := deleteResponse{}
	err := mapstructure.Decode(r.Params, params)
	if err != nil {
		log.Printf("Error in params %s: %s\n", r.Params, err.Error())
		resp.Status = paramsError
		resp.StatusMessage = err.Error()
	} else {
		_, ok := clusterMap[params.ID]
		if !ok {
			resp.Status = notFound
			resp.StatusMessage = fmt.Sprintf("Cluster %s not found", params.ID)
		} else {
			delete(clusterMap, params.ID)
			resp.Status = statusOk
			resp.StatusMessage = fmt.Sprintf("Successfully deleted cluter %s", params.ID)
		}
	}
	return resp
}

func listClusters(r *handler.Request) interface{} {
	params := new(listRequest)
	resp := listResponse{}
	err := mapstructure.Decode(r.Params, params)
	if err != nil {
		log.Printf("Error in params %s: %s\n", r.Params, err.Error())
		resp.Status = paramsError
		resp.StatusMessage = err.Error()
	} else {
		resp.Status = statusOk
		resp.StatusMessage = fmt.Sprintf("Found %d clusters", len(clusterMap))
		for _, cluster := range clusterMap {
			resp.Clusters = append(resp.Clusters, *cluster)
		}
	}
	return resp
}

// HandlerFunc handles cluster related API requests
func HandlerFunc(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		fmt.Fprintf(w, "Error reading body %s", err.Error())
		return
	}
	payload := new(handler.Request)
	err = json.Unmarshal(body, &payload)
	if err != nil {
		fmt.Fprintf(w, "Error unmarshalling json %s", err.Error())
		return
	}
	action, ok := actions[payload.Action]
	if !ok {
		fmt.Fprintf(w, "Unknown action %s", payload.Action)
		return
	}
	resp := action(payload)
	sResp, err := json.Marshal(resp)
	if err != nil {
		log.Printf("Error marshalling response %v: %s", resp, err.Error())
		w.WriteHeader(500)
		return
	}
	w.Write(sResp)
}
