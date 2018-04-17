package module

import (
	"sync"

	"github.com/l-dandelion/yi-ants-go/lib/constant"
)

// interface for the module registrar
type Registrar interface {
	Register(module Module) *constant.YiError // register an instance of module
	Unregister(mid MID) *constant.YiError     // unregiser an instance of module
	GetAll() map[MID]Module                   // get all instances of module
	Clear()                                   // clear all instances of module

	/*get an instance of module according to module type
	 *this function should load balancing strategy based on the instance is returned.
	 */
	Get(mtype int8) (Module, *constant.YiError)

	/*get all instance of module according to module type
	 */
	GetAllByType(mtype int8) (map[MID]Module, *constant.YiError)
}

func NewRegistrar() Registrar {
	return &myRegistrar{
		typeModuleMap: map[int8]map[MID]Module{},
	}
}

// implementation of interface registrar
type myRegistrar struct {
	typeModuleMap map[int8]map[MID]Module // type-MID-module map
	rwlock        sync.RWMutex            // read/write lock
}

/*
 * register an instance of module
 */
func (registrar *myRegistrar) Register(module Module) (yierr *constant.YiError) {
	//check whether module is valid
	if module == nil {
		return constant.NewYiErrorf(constant.ERR_REGISTER_MODULE, "Nil module instance")
	}
	mid := module.ID()
	mtype, yierr := GetType(mid)
	if yierr != nil {
		return
	}
	if !IsMatch(mtype, module) {
		return constant.NewYiErrorf(constant.ERR_REGISTER_MODULE,
			"Module type and module don't match. Incorrect module: %v", module)
	}

	//add the module
	registrar.rwlock.Lock()
	defer registrar.rwlock.Unlock()
	modules := registrar.typeModuleMap[mtype]
	if modules == nil {
		modules = map[MID]Module{}
	}
	modules[mid] = module
	registrar.typeModuleMap[mtype] = modules
	return nil
}

/*
 * unregister an instance of module
 */
func (registrar *myRegistrar) Unregister(mid MID) (yierr *constant.YiError) {
	mtype, yierr := GetType(mid)
	if yierr != nil {
		return
	}

	registrar.rwlock.Lock()
	defer registrar.rwlock.Unlock()
	var found bool
	if modules, ok := registrar.typeModuleMap[mtype]; ok {
		if _, ok := modules[mid]; ok {
			delete(modules, mid)
			found = true
		}
	}
	if !found {
		yierr = constant.NewYiErrorf(constant.ERR_MODULE_NOT_FOUND,
			"The module instance not found.(mid: %s)", mid)
	}
	return
}

/*
 * get all instance of module
 */
func (registrar *myRegistrar) GetAll() map[MID]Module {
	result := map[MID]Module{}
	registrar.rwlock.Lock()
	defer registrar.rwlock.Unlock()

	for _, modules := range registrar.typeModuleMap {
		for mid, module := range modules {
			result[mid] = module
		}
	}

	return result
}

/*
 * clear all instances of module
 */
func (registrar *myRegistrar) Clear() {
	registrar.typeModuleMap = map[int8]map[MID]Module{}
}

/*
 * get all instance of module according to module type
 */
func (registrar *myRegistrar) GetAllByType(mtype int8) (map[MID]Module, *constant.YiError) {
	if !LegalType(mtype) {
		return nil, constant.NewYiErrorf(constant.ERR_ILLEGAL_MODULE_TYPE,
			"Illegal module type: %d", mtype)
	}
	registrar.rwlock.RLock()
	defer registrar.rwlock.RUnlock()
	modules := registrar.typeModuleMap[mtype]
	if len(modules) == 0 {
		return nil, constant.NewYiErrorf(constant.ERR_MODULE_NOT_FOUND,
			"Not found the module instance.(mtype: %d)", mtype)
	}
	result := map[MID]Module{}
	for mid, module := range modules {
		result[mid] = module
	}
	return result, nil
}

/*
 *get an instance of module according to module type
 *this function should load balancing strategy based on the instance is returned.
 */
func (registrar *myRegistrar) Get(mtype int8) (module Module, yierr *constant.YiError) {
	modules, yierr := registrar.GetAllByType(mtype)
	if yierr != nil {
		return
	}

	//select a module instance with minimum score
	minScore := uint64(0)
	for _, m := range modules {
		SetScore(m)
		score := m.Score()
		if minScore == 0 || score < minScore {
			minScore = score
			module = m
		}
	}
	return
}
