package apollo

import (
	"github.com/shima-park/agollo"
	"github.com/pelletier/go-toml"
)


type Apollo struct {
	url string
	appName string
	nameSpace string
	cluster string
	ag agollo.Agollo
}

func NewApollo(url,appName,nameSpace,cluster string)*Apollo{
	return &Apollo{
		url:       url,
		appName:   appName,
		nameSpace: nameSpace,
		cluster:   cluster,
	}
}

func (a *Apollo)InitApollo()error{
	ag,err :=agollo.New(a.url,a.appName,agollo.PreloadNamespaces(a.nameSpace),agollo.Cluster(a.cluster),
		agollo.AutoFetchOnCacheMiss(),
		agollo.FailTolerantOnBackupExists())
	if err!=nil{
		return err
	}
	a.ag=ag
	return nil
}

func (a *Apollo)GetConfig(key string)string {
	return a.ag.Get(key)
}

func (a *Apollo)GetObj(dev string,obj interface{})error{
	res := a.ag.Get(dev)
	err :=toml.Unmarshal([]byte(res),obj)
	if err!=nil{
		return err
	}
	return nil
}
