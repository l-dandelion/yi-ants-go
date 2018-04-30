package action

type Watcher interface{
	Stop()
	IsStop() bool
	Pause()
	UnPause()
	Start()
	Run()
}