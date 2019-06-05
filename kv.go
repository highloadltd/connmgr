package connmgr

type kvPair struct {
	key   string
	value interface{}
}

type KV struct {
	kv []kvPair
}

func (kv *KV) Get(key string) (value interface{}) {
	for _, pair := range kv.kv {
		if pair.key == key {
			return pair.value
		}
	}
	return nil
}

func (kv *KV) Set(key string, value interface{}) {
	for i, pair := range kv.kv {
		if pair.key == key {
			kv.kv[i].value = value
			return
		}
	}

	kv.kv = append(kv.kv, kvPair{
		key:   key,
		value: value,
	})
}

func (kv *KV) GetWithDefault(key string, defaultValue interface{}) (value interface{}) {
	if x := kv.Get(key); x != nil {
		return x
	}
	return defaultValue
}
