package main

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/go-git/go-billy/v5"
	"github.com/go-git/go-billy/v5/osfs"
	"github.com/go-git/go-git/v5"
)

func main() {
	cwd, _ := os.Getwd()

	repo, err := NewGGRepo(cwd)
	if err != nil {
		panic(err)
	}

	if err := repo.Set("foo", "bar"); err != nil {
		panic(err)
	}

	v, err := repo.Get("foo")
	if err != nil {
		panic(err)
	}

	fmt.Printf("Value for foo: %s\n", v)
}

func NewGGRepo(path string) (*ggRepo, error) {
	gp, err := detectGitPath(path)
	if err != nil {
		return nil, err
	}

	r, err := git.PlainOpen(gp)
	if err != nil {
		return nil, err
	}

	repo := &ggRepo{
		path:  gp,
		repo:  r,
		store: osfs.New(filepath.Join(path, ".git", "git-meta")),
	}

	return repo, nil
}

type ggRepo struct {
	path  string
	repo  *git.Repository
	store billy.Filesystem
}

func (r ggRepo) Set(key, value string) error {
	keySum := fmt.Sprintf("%x", sha256.Sum256([]byte(key)))

	of, err := r.store.Create(keySum + ".json")
	if err != nil {
		return err
	}
	defer of.Close()

	kv := &KVPair{
		Key: key,
		Val: value,
	}

	bytes, err := json.MarshalIndent(kv, "", "  ")
	if err != nil {
		return err
	}

	_, err = of.Write(bytes)
	return err
}

func (r ggRepo) Get(key string) (string, error) {
	keySum := fmt.Sprintf("%x", sha256.Sum256([]byte(key)))

	of, err := r.store.Open(keySum + ".json")
	if err != nil {
		return "", err
	}
	defer of.Close()

	bytes, err := ioutil.ReadAll(of)
	if err != nil {
		return "", err
	}

	var kv KVPair

	err = json.Unmarshal(bytes, &kv)
	if err != nil {
		return "", err
	}

	return kv.Val, nil
}

type KVPair struct {
	Key string `json:"key"`
	Val string `json:"val"`
}