package util

import (
	"os"
	"path/filepath"
)

type cache struct {
	dir   string
	value map[string][]byte
}

func NewCache(dir string) *cache {
	c := &cache{
		dir:   dir,
		value: map[string][]byte{},
	}
	c.load()
	return c
}

func (c *cache) Get(key string, callback func() ([]byte, error)) ([]byte, error) {
	r, ok := c.value[Hash(key)]
	if ok {
		return r, nil
	}

	cr, err := callback()
	if err != nil {
		return nil, err
	}

	mapKey := Hash(key)
	if err := os.WriteFile(filepath.Join(c.dir, mapKey), cr, 0777); err != nil {
		panic(err)
	}

	c.value[mapKey] = cr

	return cr, nil
}

func (c *cache) load() {
	if err := os.MkdirAll(c.dir, 0777); err != nil {
		panic(err)
	}

	dir, err := os.ReadDir(c.dir)
	if err != nil {
		panic(err)
	}

	for _, e := range dir {
		if e.IsDir() {
			continue
		}

		file, err := os.ReadFile(filepath.Join(c.dir, e.Name()))
		if err != nil {
			panic(err)
		}

		c.value[e.Name()] = file
	}
}
