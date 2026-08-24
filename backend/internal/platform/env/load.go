package env

import (
	"errors"
	"os"
	"runtime"
	"strings"
	"sync"
)

type env struct {
	key   string
	value string
}

func search(filename string) (bool, error) {
	currentDir, err := os.Getwd()
	if err != nil {
		return false, err
	}

	dirsEntry, err := os.ReadDir(currentDir)
	if err != nil {
		return false, err
	}

	for _, entry := range dirsEntry {
		if !entry.IsDir() && entry.Name() == filename {
			return true, nil
		}
	}

	return false, nil
}

func validator(data string) ([]env, error) {
	if len(data) == 0 || data == "" {
		return []env{}, errors.New("content mustn't be empty")
	}

	envs := strings.Split(data, "\n")
	valids := make([]env, 0, len(envs))

	for _, keyValue := range envs {
		keyValue = strings.TrimSpace(keyValue)
		key, value, found := strings.Cut(keyValue, "=")
		if !found {
			continue
		}

		key = strings.TrimSpace(key)
		value = strings.TrimSpace(value)

		if key == "" {
			return []env{}, errors.New("key mustn´t be empty")
		}

		if len(value) == 0 {
			continue
		}

		valueLen := len(value) - 1

		if value[0] == '"' && value[valueLen] == '"' {
			value = value[1:valueLen]
		}

		if value == "" {
			continue
		}

		valids = append(valids, env{strings.ToUpper(key), value})
	}

	return valids, nil
}

func Load(filenames ...string) error {
	if len(filenames) == 0 {
		filenames = append(filenames, ".env")
	}

	queueFilenames := make(chan string, len(filenames))
	queueResults := make(chan error, len(filenames))
	workers := min(3, runtime.NumCPU())

	var wg sync.WaitGroup
	for range workers {
		wg.Go(func() {
			for filename := range queueFilenames {
				queueResults <- loadEnv(filename)
			}
		})
	}

	for _, filename := range filenames {
		queueFilenames <- filename
	}
	close(queueFilenames)

	go func() {
		wg.Wait()
		close(queueResults)
	}()

	for result := range queueResults {
		if result != nil {
			return result
		}
	}

	return nil
}

func loadEnv(filename string) error {
	exist, err := search(filename)
	if err != nil {
		return err
	}

	if !exist {
		return errors.New("file not exist")
	}

	data, err := os.ReadFile(filename)
	if err != nil {
		return err
	}

	envs, err := validator(strings.TrimSpace(string(data)))
	if err != nil {
		return err
	}

	for _, env := range envs {
		if err := os.Setenv(env.key, env.value); err != nil {
			return err
		}
	}

	return nil
}
