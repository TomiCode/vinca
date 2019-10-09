package main

import "os"
import "log"
import "encoding/json"

type Config struct {
    Database string `json:"database"`
}

func (cfg* Config) LoadConfig(file string) error {
    conf, err := os.Open(file)
    if err == os.ErrNotExist {
        log.Println("Configuration file does not exist:", file)
        return nil
    } else if err != nil {
        log.Fatal(err)
        return err
    }
    defer conf.Close()

    if err = json.NewDecoder(conf).Decode(cfg); err != nil {
        log.Fatal(err)
        return err
    }
    return nil
}